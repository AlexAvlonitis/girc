package app

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"girc/irc"
	"girc/ui"
)

const clientVersion = "girc - a simple IRC client in Go"

// handleMessage processes a message from the server.
func (a *App) handleMessage(m irc.Message) {
	switch m.Command {
	case "PING":
		if err := a.send("PONG", m.Params...); err != nil {
			a.printError(ui.StatusBuffer, err)
		}
	case "PRIVMSG":
		a.onPrivmsg(m)
	case "NOTICE":
		a.onNotice(m)
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
	case "TOPIC":
		a.onTopic(m)
	case "MODE":
		a.onMode(m)
	case "INVITE":
		a.printInfo(ui.ActiveBuffer, fmt.Sprintf("%s invites you to %s, use /join %[2]s", m.Nick(), m.Param(1)))
		a.view.Highlight(ui.StatusBuffer)
	case "ERROR":
		a.view.Println(ui.StatusBuffer, "[red]server:[-] "+ui.Escape(m.Param(0)))
	case irc.RplWelcome:
		// The server confirms the nickname it actually registered us with.
		a.s.nick = m.Param(0)
		a.s.registered = true
		a.onNumeric(m)
	case irc.RplISupport:
		if len(m.Params) > 2 {
			a.s.applyISupport(m.Params[1 : len(m.Params)-1])
		}
		a.onNumeric(m)
	case irc.RplTopic, irc.RplNoTopic, irc.RplTopicWhoTime:
		a.onTopicReply(m)
	case irc.RplNamReply:
		a.onNamReply(m)
	case irc.RplEndOfNames:
		if c := a.s.channel(m.Param(1)); c != nil {
			a.s.commitNames(c)
			a.refreshUsers(c)
		}
	case irc.ErrNicknameInUse:
		a.onNickInUse(m)
	default:
		if m.IsNumeric() {
			a.onNumeric(m)
		} else {
			a.view.Println(ui.StatusBuffer, formatChat(m.Nick(), strings.Join(m.Params, " "), false))
		}
	}
}

// channelTarget resolves a message target to a joined channel. It accepts
// STATUSMSG targets such as "@#channel" (sent to the channel's operators).
func (a *App) channelTarget(target string) *channel {
	if c := a.s.channel(target); c != nil {
		return c
	}
	return a.s.channel(strings.TrimLeft(target, a.s.prefixSymbols))
}

func (a *App) onPrivmsg(m irc.Message) {
	from, target, text := m.Nick(), m.Param(0), m.Param(1)

	cmd, args, isCTCP := irc.ParseCTCP(text)
	if isCTCP && cmd != "ACTION" {
		a.onCTCPRequest(from, cmd, args)
		return
	}

	var buffer string
	private := false
	if c := a.channelTarget(target); c != nil {
		buffer = c.name
	} else if irc.IsChannel(target) {
		buffer = ui.StatusBuffer // a channel we are not in
	} else {
		buffer = a.s.openQuery(from)
		a.view.OpenBuffer(buffer, false)
		private = true
	}

	line := formatChat(from, text, false)
	if isCTCP {
		line = formatAction(from, args)
	}
	if buffer == ui.StatusBuffer {
		line = "[gray]" + ui.Escape(target) + "[-] " + line
	}
	a.view.Println(buffer, line)

	if private || a.mentionsMe(text) {
		a.view.Highlight(buffer)
	}
}

// onCTCPRequest answers CTCP requests such as VERSION and PING.
func (a *App) onCTCPRequest(from, cmd, args string) {
	var reply string
	switch cmd {
	case "VERSION":
		reply = clientVersion
	case "PING":
		reply = args
	case "TIME":
		reply = time.Now().Format(time.RFC1123Z)
	case "CLIENTINFO":
		reply = "ACTION CLIENTINFO PING TIME VERSION"
	default:
		a.printInfo(ui.StatusBuffer, fmt.Sprintf("Unsupported CTCP %s request from %s", cmd, from))
		return
	}
	if err := a.send("NOTICE", from, irc.CTCP(cmd, reply)); err != nil {
		a.printError(ui.StatusBuffer, err)
	}
	a.printInfo(ui.StatusBuffer, fmt.Sprintf("CTCP %s request from %s", cmd, from))
}

func (a *App) onNotice(m irc.Message) {
	from, target, text := m.Nick(), m.Param(0), m.Param(1)

	if cmd, args, ok := irc.ParseCTCP(text); ok {
		if cmd == "PING" {
			if sent, err := strconv.ParseInt(args, 10, 64); err == nil {
				args = time.Since(time.UnixMilli(sent)).Round(time.Millisecond).String()
			}
		}
		a.printInfo(ui.ActiveBuffer, fmt.Sprintf("CTCP %s reply from %s: %s", cmd, from, args))
		return
	}

	buffer := ui.ActiveBuffer
	if c := a.channelTarget(target); c != nil {
		buffer = c.name
	} else if name, ok := a.s.query(from); ok {
		buffer = name
	} else if !a.s.registered || !strings.Contains(m.Prefix, "!") {
		buffer = ui.StatusBuffer // server notices
	}
	a.view.Println(buffer, fmt.Sprintf("[purple]-%s-[-] %s", ui.Escape(from), ui.Escape(text)))
}

func (a *App) onJoin(m irc.Message) {
	nick, name := m.Nick(), m.Param(0)

	if a.s.isMe(nick) {
		c := a.s.join(name)
		a.view.OpenBuffer(c.name, true)
		a.view.SetTopic(c.name, "")
		a.refreshUsers(c) // filled in by the NAMES replies that follow
		a.printInfo(c.name, "You have joined "+c.name)
		return
	}

	c := a.s.channel(name)
	if c == nil {
		return
	}
	c.members[irc.Fold(nick)] = &member{nick: nick}
	a.refreshUsers(c)
	a.printInfo(c.name, fmt.Sprintf("%s has joined %s", nick, c.name))
}

func (a *App) onPart(m irc.Message) {
	nick, name := m.Nick(), m.Param(0)
	c := a.s.channel(name)
	if c == nil {
		return
	}

	if a.s.isMe(nick) {
		a.s.part(c.name)
		a.view.CloseBuffer(c.name)
		a.printInfo(ui.StatusBuffer, "You have left "+c.name+reason(m.Param(1)))
		return
	}
	delete(c.members, irc.Fold(nick))
	a.refreshUsers(c)
	a.printInfo(c.name, fmt.Sprintf("%s has left %s%s", nick, c.name, reason(m.Param(1))))
}

func (a *App) onKick(m irc.Message) {
	name, victim := m.Param(0), m.Param(1)
	c := a.s.channel(name)
	if c == nil {
		return
	}

	if a.s.isMe(victim) {
		// Keep the buffer open so the user can see what happened.
		a.s.part(c.name)
		a.view.SetUsers(c.name, nil)
		a.printInfo(c.name, fmt.Sprintf("You were kicked from %s by %s%s", c.name, m.Nick(), reason(m.Param(2))))
		a.view.Highlight(c.name)
		return
	}
	delete(c.members, irc.Fold(victim))
	a.refreshUsers(c)
	a.printInfo(c.name, fmt.Sprintf("%s was kicked by %s%s", victim, m.Nick(), reason(m.Param(2))))
}

func (a *App) onQuit(m irc.Message) {
	nick := m.Nick()
	text := fmt.Sprintf("%s has quit%s", nick, reason(m.Param(0)))
	for _, c := range a.s.channelsWith(nick) {
		delete(c.members, irc.Fold(nick))
		a.refreshUsers(c)
		a.printInfo(c.name, text)
	}
	if name, ok := a.s.query(nick); ok {
		a.printInfo(name, text)
	}
}

func (a *App) onNick(m irc.Message) {
	oldNick, newNick := m.Nick(), m.Param(0)
	text := fmt.Sprintf("%s is now known as %s", oldNick, newNick)

	if a.s.isMe(oldNick) {
		a.s.nick = newNick
		a.printInfo(ui.StatusBuffer, text)
	}
	for _, c := range a.s.channelsWith(oldNick) {
		mem := c.member(oldNick)
		delete(c.members, irc.Fold(oldNick))
		mem.nick = newNick
		c.members[irc.Fold(newNick)] = mem
		a.refreshUsers(c)
		a.printInfo(c.name, text)
	}
	if name, ok := a.s.renameQuery(oldNick, newNick); ok {
		a.view.RenameBuffer(name, newNick)
		a.printInfo(newNick, text)
	}
}

func (a *App) onTopic(m irc.Message) {
	c := a.s.channel(m.Param(0))
	if c == nil {
		return
	}
	c.topic = m.Param(1)
	a.view.SetTopic(c.name, c.topic)
	a.printInfo(c.name, fmt.Sprintf("%s changed the topic to: %s", m.Nick(), c.topic))
}

// onTopicReply handles RPL_TOPIC, RPL_NOTOPIC and RPL_TOPICWHOTIME, which
// all start with "<me> <channel>".
func (a *App) onTopicReply(m irc.Message) {
	c := a.s.channel(m.Param(1))
	if c == nil {
		a.onNumeric(m)
		return
	}

	switch m.Command {
	case irc.RplTopic:
		c.topic = m.Param(2)
		a.view.SetTopic(c.name, c.topic)
		a.printInfo(c.name, "Topic: "+c.topic)
	case irc.RplNoTopic:
		a.printInfo(c.name, "No topic is set")
	case irc.RplTopicWhoTime:
		text := "Topic set by " + m.Param(2)
		if ts, err := strconv.ParseInt(m.Param(3), 10, 64); err == nil {
			text += " on " + time.Unix(ts, 0).Format("2006-01-02 15:04")
		}
		a.printInfo(c.name, text)
	}
}

func (a *App) onMode(m irc.Message) {
	target := m.Param(0)
	text := fmt.Sprintf("%s sets mode %s", m.Nick(), strings.Join(m.Params[min(1, len(m.Params)):], " "))

	c := a.s.channel(target)
	if c == nil {
		a.printInfo(ui.StatusBuffer, text) // user modes
		return
	}
	if len(m.Params) > 1 {
		a.s.applyModes(c, m.Params[1], m.Params[2:])
		a.refreshUsers(c)
	}
	a.printInfo(c.name, text)
}

// onNamReply handles RPL_NAMREPLY: "<me> <symbol> <channel> :<nicks>".
func (a *App) onNamReply(m irc.Message) {
	name, list := m.Param(2), m.Param(3)
	c := a.s.channel(name)
	if c == nil {
		a.printInfo(ui.ActiveBuffer, fmt.Sprintf("Users in %s: %s", name, list))
		return
	}
	a.s.addNames(c, strings.Fields(list))
}

// onNickInUse picks an alternative nickname if registration has not completed;
// otherwise it reports that a /nick change failed.
func (a *App) onNickInUse(m irc.Message) {
	if a.s.registered {
		a.onNumeric(m)
		return
	}
	a.s.nick += "_"
	a.printInfo(ui.StatusBuffer, fmt.Sprintf("Nickname %s is already in use, trying %s", m.Param(1), a.s.nick))
	if err := a.send("NICK", a.s.nick); err != nil {
		a.printError(ui.StatusBuffer, err)
	}
}

// onNumeric displays a numeric reply in the most relevant buffer: the channel
// it is about, the active buffer for replies to user commands, or the status
// buffer.
func (a *App) onNumeric(m irc.Message) {
	params := m.Params
	if len(params) > 0 {
		params = params[1:] // the first parameter of a numeric is our own nick
	}
	if len(params) == 0 {
		return
	}

	buffer := ui.StatusBuffer
	switch c := a.s.channel(params[0]); {
	case c != nil:
		buffer = c.name
	case a.s.registered && irc.IsReplyToUser(m.Command):
		buffer = ui.ActiveBuffer
	}

	text := ui.Escape(strings.Join(params, " "))
	if m.Command[0] == '4' || m.Command[0] == '5' {
		text = "[red]" + text + "[-]"
	}
	a.view.Println(buffer, "[blue]*[-] "+text)
}

// mentionsMe reports whether text contains the user's nick as a word.
func (a *App) mentionsMe(text string) bool {
	nick := irc.Fold(a.s.nick)
	folded := irc.Fold(text)
	for i := strings.Index(folded, nick); i >= 0; {
		end := i + len(nick)
		if (i == 0 || !isNickChar(folded[i-1])) && (end == len(folded) || !isNickChar(folded[end])) {
			return true
		}
		next := strings.Index(folded[i+1:], nick)
		if next < 0 {
			break
		}
		i += 1 + next
	}
	return false
}

func isNickChar(c byte) bool {
	return c >= 'a' && c <= 'z' || c >= '0' && c <= '9' || strings.IndexByte("-_{}|^`", c) >= 0
}
