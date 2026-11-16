package app

import (
	"errors"
	"fmt"
	"slices"
	"strconv"
	"strings"
	"time"

	"girc/irc"
	"girc/ui"
)

// errUsage is returned by a command handler when its arguments are invalid;
// the caller replaces it with the command's usage string.
var errUsage = errors.New("invalid arguments")

var (
	errNoTarget  = errors.New("this buffer has no channel or conversation, use /join #channel or /query nick")
	errNoChannel = errors.New("this is not a channel buffer, use /join #channel")
)

type command struct {
	usage       string
	description string
	run         func(in input, args string) error
}

func (a *App) newCommands() map[string]command {
	return map[string]command{
		"help":    {"/help", "show this help", a.cmdHelp},
		"join":    {"/join #channel[,#other] [key]", "join channels", a.cmdJoin},
		"part":    {"/part [#channel] [reason]", "leave a channel", a.cmdPart},
		"close":   {"/close [reason]", "close the buffer, leaving its channel", a.cmdClose},
		"query":   {"/query nick [message]", "open a private conversation", a.cmdQuery},
		"msg":     {"/msg target message", "send a message", a.cmdMsg},
		"me":      {"/me action", "send an action", a.cmdMe},
		"notice":  {"/notice target message", "send a notice", a.cmdNotice},
		"ctcp":    {"/ctcp nick command [args]", "send a CTCP request", a.cmdCTCP},
		"nick":    {"/nick newnick", "change your nickname", a.cmdNick},
		"topic":   {"/topic [new topic]", "show or set the channel topic", a.cmdTopic},
		"mode":    {"/mode [target] [modes [params]]", "show or change modes", a.cmdMode},
		"op":      {"/op nick...", "give channel operator status", a.modeCmd("+o")},
		"deop":    {"/deop nick...", "take channel operator status", a.modeCmd("-o")},
		"voice":   {"/voice nick...", "give voice", a.modeCmd("+v")},
		"devoice": {"/devoice nick...", "take voice", a.modeCmd("-v")},
		"ban":     {"/ban nick|mask...", "ban users from the channel", a.banCmd("+b")},
		"unban":   {"/unban mask...", "remove bans", a.banCmd("-b")},
		"kick":    {"/kick nick [reason]", "kick a user from the channel", a.cmdKick},
		"invite":  {"/invite nick [#channel]", "invite a user to a channel", a.cmdInvite},
		"whois":   {"/whois nick", "show information about a user", a.cmdWhois},
		"who":     {"/who mask", "list users matching a mask", a.cmdWho},
		"away":    {"/away [message]", "set, or clear without a message, your away status", a.cmdAway},
		"list":    {"/list [pattern]", "list channels on the server", a.cmdList},
		"names":   {"/names [#channel]", "list the users in a channel", a.cmdNames},
		"quote":   {"/quote command", "send a raw command to the server", a.cmdQuote},
		"clear":   {"/clear", "clear the buffer", a.cmdClear},
		"quit":    {"/quit [reason]", "quit the server", a.cmdQuit},
	}
}

// handleInput processes a line typed by the user.
func (a *App) handleInput(in input) {
	if err := a.execute(in); err != nil {
		a.printError(in.buffer, err)
	}
}

func (a *App) execute(in input) error {
	text := strings.TrimSpace(in.text)
	switch {
	case text == "":
		return nil
	case strings.HasPrefix(text, "//"):
		return a.say(in, text[1:]) // "//text" sends a message starting with "/"
	case !strings.HasPrefix(text, "/"):
		return a.say(in, text)
	}

	name, args, _ := strings.Cut(text[1:], " ")
	cmd, ok := a.commands[strings.ToLower(name)]
	if !ok {
		return fmt.Errorf("unknown command /%s, use /help to list commands", name)
	}

	err := cmd.run(in, strings.TrimSpace(args))
	if errors.Is(err, errUsage) {
		return fmt.Errorf("usage: %s", cmd.usage)
	}
	return err
}

// target returns the channel or nick the buffer talks to.
func (a *App) target(buffer string) (string, error) {
	if c := a.s.channel(buffer); c != nil {
		return c.name, nil
	}
	if name, ok := a.s.query(buffer); ok {
		return name, nil
	}
	if irc.IsChannel(buffer) {
		return "", fmt.Errorf("you are not in %s, use /join %s", buffer, buffer)
	}
	return "", errNoTarget
}

// channelArg splits an optional leading channel name off args, defaulting to
// the buffer's channel.
func (a *App) channelArg(in input, args string) (channel, rest string, err error) {
	if first, rest, _ := strings.Cut(args, " "); irc.IsChannel(first) {
		return first, strings.TrimSpace(rest), nil
	}
	if c := a.s.channel(in.buffer); c != nil {
		return c.name, args, nil
	}
	return "", args, errNoChannel
}

// say sends text to the buffer's channel or conversation.
func (a *App) say(in input, text string) error {
	target, err := a.target(in.buffer)
	if err != nil {
		return err
	}
	if err := a.send("PRIVMSG", target, text); err != nil {
		return err
	}
	// The server does not echo our own messages back.
	a.view.Println(in.buffer, formatChat(a.s.nick, text, true))
	return nil
}

func (a *App) cmdHelp(input, string) error {
	names := make([]string, 0, len(a.commands))
	for name := range a.commands {
		names = append(names, name)
	}
	slices.Sort(names)

	lines := []string{"[yellow]Commands:[-]"}
	for _, name := range names {
		cmd := a.commands[name]
		lines = append(lines, fmt.Sprintf("  %-34s %s", ui.Escape(cmd.usage), cmd.description))
	}
	lines = append(lines,
		"  Other text is sent to the current channel or conversation; start it with // to send a leading /.",
		"[yellow]Keys:[-] Ctrl+N/Ctrl+P next/previous buffer, Alt+1..9 jump to buffer, PgUp/PgDn scroll, Tab move focus.",
	)
	a.view.Println(ui.ActiveBuffer, strings.Join(lines, "\n"))
	return nil
}

func (a *App) cmdJoin(_ input, args string) error {
	fields := strings.Fields(args)
	if len(fields) == 0 || len(fields) > 2 {
		return errUsage
	}
	return a.send("JOIN", fields...)
}

func (a *App) cmdPart(in input, args string) error {
	channel, reason, err := a.channelArg(in, args)
	if err != nil {
		return err
	}
	if reason == "" {
		return a.send("PART", channel)
	}
	return a.send("PART", channel, reason)
}

func (a *App) cmdClose(in input, args string) error {
	switch {
	case in.buffer == ui.StatusBuffer:
		return errors.New("the status buffer cannot be closed, use /quit to exit")
	case a.s.channel(in.buffer) != nil:
		// The buffer closes when the server confirms the PART.
		return a.cmdPart(in, args)
	default:
		a.s.closeQuery(in.buffer)
		a.view.CloseBuffer(in.buffer)
		return nil
	}
}

func (a *App) cmdQuery(_ input, args string) error {
	nick, text, _ := strings.Cut(args, " ")
	if nick == "" || irc.IsChannel(nick) {
		return errUsage
	}
	name := a.s.openQuery(nick)
	a.view.OpenBuffer(name, true)
	if text = strings.TrimSpace(text); text != "" {
		return a.say(input{buffer: name}, text)
	}
	return nil
}

func (a *App) cmdMsg(in input, args string) error {
	target, text, _ := strings.Cut(args, " ")
	text = strings.TrimSpace(text)
	if target == "" || text == "" {
		return errUsage
	}
	if err := a.send("PRIVMSG", target, text); err != nil {
		return err
	}

	buffer := ui.ActiveBuffer
	switch {
	case a.s.channel(target) != nil:
		buffer = a.s.channel(target).name
	case !irc.IsChannel(target):
		buffer = a.s.openQuery(target)
		a.view.OpenBuffer(buffer, false)
	}
	line := formatChat(a.s.nick, text, true)
	if buffer == ui.ActiveBuffer {
		line = fmt.Sprintf("[yellow]-> %s[-] %s", ui.Escape(target), ui.Escape(text))
	}
	a.view.Println(buffer, line)
	return nil
}

func (a *App) cmdMe(in input, args string) error {
	if args == "" {
		return errUsage
	}
	target, err := a.target(in.buffer)
	if err != nil {
		return err
	}
	if err := a.send("PRIVMSG", target, irc.CTCP("ACTION", args)); err != nil {
		return err
	}
	a.view.Println(in.buffer, formatAction(a.s.nick, args))
	return nil
}

func (a *App) cmdNotice(in input, args string) error {
	target, text, _ := strings.Cut(args, " ")
	text = strings.TrimSpace(text)
	if target == "" || text == "" {
		return errUsage
	}
	if err := a.send("NOTICE", target, text); err != nil {
		return err
	}
	a.view.Println(in.buffer, fmt.Sprintf("[purple]-> -%s-[-] %s", ui.Escape(target), ui.Escape(text)))
	return nil
}

func (a *App) cmdCTCP(in input, args string) error {
	fields := strings.SplitN(args, " ", 3)
	if len(fields) < 2 {
		return errUsage
	}
	nick, cmd, ctcpArgs := fields[0], strings.ToUpper(fields[1]), ""
	if len(fields) == 3 {
		ctcpArgs = fields[2]
	}
	if cmd == "PING" && ctcpArgs == "" {
		ctcpArgs = strconv.FormatInt(time.Now().UnixMilli(), 10)
	}
	if err := a.send("PRIVMSG", nick, irc.CTCP(cmd, ctcpArgs)); err != nil {
		return err
	}
	a.printInfo(in.buffer, fmt.Sprintf("CTCP %s request sent to %s", cmd, nick))
	return nil
}

func (a *App) cmdNick(_ input, args string) error {
	fields := strings.Fields(args)
	if len(fields) != 1 {
		return errUsage
	}
	return a.send("NICK", fields[0])
}

func (a *App) cmdTopic(in input, args string) error {
	channel, topic, err := a.channelArg(in, args)
	if err != nil {
		return err
	}
	if topic == "" {
		return a.send("TOPIC", channel) // ask the server for the topic
	}
	return a.send("TOPIC", channel, topic)
}

func (a *App) cmdMode(in input, args string) error {
	fields := strings.Fields(args)
	if len(fields) == 0 || strings.ContainsAny(fields[0][:1], "+-") {
		target, err := a.target(in.buffer)
		if err != nil {
			return errUsage
		}
		fields = append([]string{target}, fields...)
	}
	return a.send("MODE", fields...)
}

// modeCmd returns a handler that sets a prefix mode such as "+o" on the
// given nicks, in as few MODE commands as the server allows.
func (a *App) modeCmd(mode string) func(input, string) error {
	return func(in input, args string) error {
		nicks := strings.Fields(args)
		if len(nicks) == 0 {
			return errUsage
		}
		return a.setModes(in, mode, nicks)
	}
}

// banCmd returns a handler that adds or removes bans. Plain nicks are turned
// into nick!*@* masks.
func (a *App) banCmd(mode string) func(input, string) error {
	return func(in input, args string) error {
		masks := strings.Fields(args)
		if len(masks) == 0 {
			return errUsage
		}
		for i, mask := range masks {
			if !strings.ContainsAny(mask, "!@*") {
				masks[i] = mask + "!*@*"
			}
		}
		return a.setModes(in, mode, masks)
	}
}

func (a *App) setModes(in input, mode string, params []string) error {
	c := a.s.channel(in.buffer)
	if c == nil {
		return errNoChannel
	}
	sign, letter := mode[:1], mode[1:]
	for len(params) > 0 {
		n := min(len(params), a.s.maxModes)
		chunk := params[:n]
		params = params[n:]
		modes := sign + strings.Repeat(letter, n)
		if err := a.send("MODE", append([]string{c.name, modes}, chunk...)...); err != nil {
			return err
		}
	}
	return nil
}

func (a *App) cmdKick(in input, args string) error {
	c := a.s.channel(in.buffer)
	if c == nil {
		return errNoChannel
	}
	nick, reason, _ := strings.Cut(args, " ")
	if nick == "" {
		return errUsage
	}
	if reason = strings.TrimSpace(reason); reason != "" {
		return a.send("KICK", c.name, nick, reason)
	}
	return a.send("KICK", c.name, nick)
}

func (a *App) cmdInvite(in input, args string) error {
	fields := strings.Fields(args)
	switch len(fields) {
	case 1:
		c := a.s.channel(in.buffer)
		if c == nil {
			return errNoChannel
		}
		return a.send("INVITE", fields[0], c.name)
	case 2:
		return a.send("INVITE", fields...)
	default:
		return errUsage
	}
}

func (a *App) cmdWhois(_ input, args string) error {
	fields := strings.Fields(args)
	if len(fields) != 1 {
		return errUsage
	}
	return a.send("WHOIS", fields[0])
}

func (a *App) cmdWho(in input, args string) error {
	fields := strings.Fields(args)
	if len(fields) == 0 {
		target, err := a.target(in.buffer)
		if err != nil {
			return errUsage
		}
		fields = []string{target}
	}
	return a.send("WHO", fields...)
}

func (a *App) cmdAway(_ input, args string) error {
	if args == "" {
		return a.send("AWAY")
	}
	return a.send("AWAY", args)
}

func (a *App) cmdList(_ input, args string) error {
	return a.send("LIST", strings.Fields(args)...)
}

func (a *App) cmdNames(in input, args string) error {
	channel := args
	if channel == "" {
		c := a.s.channel(in.buffer)
		if c == nil {
			return errUsage
		}
		channel = c.name
	}
	return a.send("NAMES", channel)
}

func (a *App) cmdQuote(_ input, args string) error {
	m, err := irc.ParseMessage(args)
	if err != nil {
		return errUsage
	}
	return a.send(m.Command, m.Params...)
}

func (a *App) cmdClear(in input, _ string) error {
	a.view.ClearBuffer(in.buffer)
	return nil
}

func (a *App) cmdQuit(_ input, args string) error {
	if args == "" {
		args = defaultQuitMessage
	}
	a.quitting.Store(true)
	return a.send("QUIT", args)
}
