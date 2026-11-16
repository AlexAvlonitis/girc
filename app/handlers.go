package app

import (
	"fmt"
	"strings"

	"girc/irc"
	"girc/ui"
)

const ctcpDelim = "\x01"

// handleMessage processes a message from the server. It runs on the network
// goroutine.
func (a *App) handleMessage(m irc.Message) {
	switch m.Command {
	case "PING":
		if err := a.conn.Send("PONG", m.Params...); err != nil {
			a.printError(err)
		}
	case "PRIVMSG", "NOTICE":
		a.onPrivmsg(m)
	case "JOIN":
		a.onJoin(m)
	case "PART":
		a.onPart(m)
	case "KICK":
		a.onKick(m)
	case "QUIT":
		a.onQuit(m)
	case "NICK":
		a.onNick(m)
	case "ERROR":
		a.view.Println("[red]server:[-] " + ui.Escape(m.Param(0)))
	case irc.RplWelcome:
		// The server confirms the nickname it actually registered us with.
		a.state.SetNick(m.Param(0))
		a.state.SetRegistered()
		a.onOther(m)
	case irc.RplNamReply:
		a.onNamReply(m)
	case irc.RplEndOfNames:
		if a.state.IsCurrent(m.Param(1)) {
			a.state.CommitNames()
			a.refreshUsers()
		}
	case irc.ErrNicknameInUse:
		a.onNickInUse(m)
	default:
		a.onOther(m)
	}
}

func (a *App) onPrivmsg(m irc.Message) {
	nick, target, text := m.Nick(), m.Param(0), m.Param(1)

	var line string
	switch {
	case m.Command == "NOTICE":
		line = fmt.Sprintf("[purple]-%s-[-] %s", ui.Escape(nick), ui.Escape(text))
	case strings.HasPrefix(text, ctcpDelim+"ACTION "):
		action := strings.TrimSuffix(strings.TrimPrefix(text, ctcpDelim+"ACTION "), ctcpDelim)
		line = fmt.Sprintf("[green]* %s[-] %s", ui.Escape(nick), ui.Escape(action))
	case strings.HasPrefix(text, ctcpDelim):
		return // other CTCP requests (VERSION, PING, ...) are not supported
	case a.state.IsMe(target):
		line = fmt.Sprintf("[green]<%s>[-][yellow](Private)[-] %s", ui.Escape(nick), ui.Escape(text))
	default:
		line = formatChat(nick, text)
	}

	if irc.IsChannel(target) && !a.state.IsCurrent(target) {
		line = "[gray]" + ui.Escape(target) + "[-] " + line
	}
	a.view.Println(line)
}

func (a *App) onJoin(m irc.Message) {
	nick, channel := m.Nick(), m.Param(0)

	switch {
	case a.state.IsMe(nick):
		// NAMES replies that follow will fill in the user list.
		a.state.Join(channel)
		a.refreshUsers()
	case a.state.IsCurrent(channel):
		a.state.AddUser(nick)
		a.refreshUsers()
	}
	a.printInfo(fmt.Sprintf("%s has joined %s", nick, channel))
}

func (a *App) onPart(m irc.Message) {
	nick, channel := m.Nick(), m.Param(0)
	a.userLeft(nick, channel)
	a.printInfo(fmt.Sprintf("%s has left %s%s", nick, channel, reason(m.Param(1))))
}

func (a *App) onKick(m irc.Message) {
	channel, victim := m.Param(0), m.Param(1)
	a.userLeft(victim, channel)
	a.printInfo(fmt.Sprintf("%s was kicked from %s by %s%s", victim, channel, m.Nick(), reason(m.Param(2))))
}

// userLeft updates the state after nick leaves channel.
func (a *App) userLeft(nick, channel string) {
	if !a.state.IsCurrent(channel) {
		return
	}
	if a.state.IsMe(nick) {
		a.state.Leave()
	} else {
		a.state.RemoveUser(nick)
	}
	a.refreshUsers()
}

func (a *App) onQuit(m irc.Message) {
	// Only report quits of users we can see, to avoid noise.
	if a.state.RemoveUser(m.Nick()) {
		a.refreshUsers()
		a.printInfo(fmt.Sprintf("%s has quit%s", m.Nick(), reason(m.Param(0))))
	}
}

func (a *App) onNick(m irc.Message) {
	oldNick, newNick := m.Nick(), m.Param(0)

	if a.state.IsMe(oldNick) {
		a.state.SetNick(newNick)
	}
	if a.state.RenameUser(oldNick, newNick) {
		a.refreshUsers()
	}
	a.printInfo(fmt.Sprintf("%s is now known as %s", oldNick, newNick))
}

// onNamReply handles RPL_NAMREPLY: "<me> <symbol> <channel> :<nicks>".
func (a *App) onNamReply(m irc.Message) {
	channel, list := m.Param(2), m.Param(3)
	if !a.state.IsCurrent(channel) {
		a.printInfo(fmt.Sprintf("Users in %s: %s", channel, list))
		return
	}

	fields := strings.Fields(list)
	nicks := make([]string, 0, len(fields))
	for _, name := range fields {
		// Strip channel membership prefixes (owner, admin, op, halfop, voice).
		if nick := strings.TrimLeft(name, "~&@%+"); nick != "" {
			nicks = append(nicks, nick)
		}
	}
	a.state.AddNames(nicks)
}

// onNickInUse picks an alternative nickname if registration has not completed;
// otherwise it reports that a /nick change failed.
func (a *App) onNickInUse(m irc.Message) {
	if a.state.Registered() {
		a.onOther(m)
		return
	}
	alt := a.state.Nick() + "_"
	a.state.SetNick(alt)
	a.printInfo(fmt.Sprintf("Nickname %s is already in use, trying %s", m.Param(1), alt))
	if err := a.conn.Send("NICK", alt); err != nil {
		a.printError(err)
	}
}

// onOther displays messages the client has no special handling for.
func (a *App) onOther(m irc.Message) {
	params := m.Params
	if m.IsNumeric() && len(params) > 0 {
		params = params[1:] // the first parameter of a numeric is our own nick
	}
	if len(params) == 0 {
		return
	}
	a.view.Println(formatChat(m.Nick(), strings.Join(params, " ")))
}

func formatChat(nick, text string) string {
	return fmt.Sprintf("[green]<%s>[-] %s", ui.Escape(nick), ui.Escape(text))
}

func reason(r string) string {
	if r == "" {
		return ""
	}
	return " (" + r + ")"
}
