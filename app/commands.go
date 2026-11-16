package app

import (
	"errors"
	"fmt"
	"slices"
	"strings"

	"girc/irc"
	"girc/ui"
)

// errUsage is returned by a command handler when its arguments are invalid;
// the caller replaces it with the command's usage string.
var errUsage = errors.New("invalid arguments")

var errNoChannel = errors.New("you need to join a channel first, use /join #channel")

type command struct {
	usage       string
	description string
	run         func(args string) error
}

func (a *App) newCommands() map[string]command {
	return map[string]command{
		"join":  {"/join #channel [key]", "join a channel", a.cmdJoin},
		"part":  {"/part [#channel] [reason]", "leave a channel", a.cmdPart},
		"nick":  {"/nick newnick", "change your nickname", a.cmdNick},
		"msg":   {"/msg nickname message", "send a private message", a.cmdMsg},
		"names": {"/names [#channel]", "list the users in a channel", a.cmdNames},
		"quit":  {"/quit [reason]", "quit the server", a.cmdQuit},
		"help":  {"/help", "show this help", a.cmdHelp},
	}
}

// handleInput processes a line typed by the user. It runs on the UI goroutine.
func (a *App) handleInput(input string) {
	if err := a.execute(strings.TrimSpace(input)); err != nil {
		a.printError(err)
	}
}

func (a *App) execute(input string) error {
	if input == "" {
		return nil
	}
	if !strings.HasPrefix(input, "/") {
		return a.say(input)
	}

	name, args, _ := strings.Cut(input[1:], " ")
	cmd, ok := a.commands[strings.ToLower(name)]
	if !ok {
		return fmt.Errorf("unknown command /%s, use /help to list commands", name)
	}

	err := cmd.run(strings.TrimSpace(args))
	if errors.Is(err, errUsage) {
		return fmt.Errorf("usage: %s", cmd.usage)
	}
	return err
}

// say sends text to the current channel.
func (a *App) say(text string) error {
	channel := a.state.Channel()
	if channel == "" {
		return errNoChannel
	}
	if err := a.conn.Send("PRIVMSG", channel, text); err != nil {
		return err
	}
	// The server does not echo our own messages back.
	a.view.Println(formatChat(a.state.Nick(), text))
	return nil
}

func (a *App) cmdJoin(args string) error {
	fields := strings.Fields(args)
	if len(fields) == 0 || len(fields) > 2 {
		return errUsage
	}
	return a.conn.Send("JOIN", fields...)
}

func (a *App) cmdPart(args string) error {
	channel, msg := a.state.Channel(), args
	if first, rest, _ := strings.Cut(args, " "); irc.IsChannel(first) {
		channel, msg = first, strings.TrimSpace(rest)
	}
	if channel == "" {
		return errNoChannel
	}
	if msg == "" {
		return a.conn.Send("PART", channel)
	}
	return a.conn.Send("PART", channel, msg)
}

func (a *App) cmdNick(args string) error {
	fields := strings.Fields(args)
	if len(fields) != 1 {
		return errUsage
	}
	return a.conn.Send("NICK", fields[0])
}

func (a *App) cmdMsg(args string) error {
	target, text, _ := strings.Cut(args, " ")
	text = strings.TrimSpace(text)
	if target == "" || text == "" {
		return errUsage
	}
	if err := a.conn.Send("PRIVMSG", target, text); err != nil {
		return err
	}
	a.view.Println(fmt.Sprintf("[green]<%s>[-][yellow](to %s)[-] %s",
		ui.Escape(a.state.Nick()), ui.Escape(target), ui.Escape(text)))
	return nil
}

func (a *App) cmdNames(args string) error {
	channel := args
	if channel == "" {
		channel = a.state.Channel()
	}
	if channel == "" {
		return errUsage
	}
	return a.conn.Send("NAMES", channel)
}

func (a *App) cmdQuit(args string) error {
	if args == "" {
		args = defaultQuitMessage
	}
	a.quitting.Store(true)
	return a.conn.Send("QUIT", args)
}

func (a *App) cmdHelp(string) error {
	names := make([]string, 0, len(a.commands))
	for name := range a.commands {
		names = append(names, name)
	}
	slices.Sort(names)

	a.view.Println("[yellow]Commands:[-]")
	for _, name := range names {
		cmd := a.commands[name]
		a.view.Println(fmt.Sprintf("  %-28s %s", ui.Escape(cmd.usage), cmd.description))
	}
	a.view.Println("  Anything not starting with / is sent to the current channel.")
	return nil
}
