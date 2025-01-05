package connection

import (
	"errors"
	"girc/interfaces"
	"girc/ui"
	"strings"
)

type Message struct {
	Source  string
	Command string
	Args    []string
}

type MessageParser struct {
	Client interfaces.Client
	Ui     *ui.UI
}

func NewMessageParser(c interfaces.Client, ui *ui.UI) *MessageParser {
	return &MessageParser{Client: c, Ui: ui}
}

// Parse formats the messages received from the server, and categorizes them
func (p *MessageParser) Parse(msg string) string {
	message, err := parseMsg(msg)
	if err != nil {
		return ""
	}

	if p.isPrivateMessage(message) {
		return p.formatPrivateMsg(message)
	} else if p.isNickChange(message) {
		return p.formatNickChange(message)
	} else if p.isNames(message) {
		return p.formatNames(message)
	} else if p.isPing(message) {
		p.Client.Write("PONG " + p.formatPing(msg)) // keep the connection alive
		return ""
	}

	return message.printMessage()
}

func (p *MessageParser) NamesToList(msg string) []string {
	return strings.Split(msg, " ")
}

// format ping messages, get the message after the PING
func (p *MessageParser) formatPing(msg string) string {
	return msg[5:]
}

// FormatNames parse the names of the users in a channel
func (p *MessageParser) formatNames(msg *Message) string {
	// get names after 4th index
	names := msg.Args[3]
	p.Ui.UsersView.Clear()
	for _, name := range strings.Split(names, " ") {
		p.Ui.UsersView.AddItem(name, "", 0, nil)
	}
	p.Ui.App.Draw()

	return ""
}

func (p *MessageParser) isNames(msg *Message) bool {
	return msg.Command == "353"
}

func (p *MessageParser) isPrivateMessage(msg *Message) bool {
	return msg.Command == "PRIVMSG"
}

func (p *MessageParser) isNickChange(msg *Message) bool {
	return msg.Command == "NICK"
}

func (p *MessageParser) isPing(msg *Message) bool {
	return msg.Command == "PING"
}

func (p *MessageParser) formatNickChange(msg *Message) string {
	return msg.Source + " is now known as " + msg.Args[0]
}

func (p *MessageParser) formatPrivateMsg(msg *Message) string {
	// check if the message is a private message and directed to the user
	if msg.Args[0] == p.Client.Nick() {
		return "<" + msg.Source + ">(Private) " + msg.Args[1]
	} else {
		return "<" + msg.Source + "> " + msg.Args[1]
	}
}

func (l *Message) printMessage() string {
	return "<" + l.Source + ">" + strings.Join(l.Args, " ")
}

// parseMsg Breaks a message from an IRC server into its prefix, command, and arguments
// parsemsg(":test!~test@test.com PRIVMSG #channel :Hi!")
// Message('test!~test@test.com', 'PRIVMSG', ['#channel', 'Hi!'])
// https://stackoverflow.com/questions/930700/python-parsing-irc-messages
func parseMsg(s string) (*Message, error) {
	var prefix string
	var trailing string
	var args []string

	if s == "" {
		return nil, errors.New("empty string")
	}

	if s[0] == ':' {
		split := strings.SplitN(s[1:], " ", 2)
		prefix = split[0]
		s = split[1]
	}

	if strings.Contains(s, " :") {
		split := strings.SplitN(s, " :", 2)
		s = split[0]
		trailing = split[1]
		args = strings.Split(s, " ")
		args = append(args, trailing)
	} else {
		args = strings.Split(s, " ")
	}

	command := args[0]
	args = args[1:]

	return &Message{
		Source:  prefix,
		Command: command,
		Args:    args,
	}, nil
}
