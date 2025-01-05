package connection

import (
	"errors"
	"girc/interfaces"
	"girc/ui"
	"strings"
)

const (
	privMsg  = "PRIVMSG"
	nickMsg  = "NICK"
	namesMsg = "353"
	pingMsg  = "PING"
	joinMsg  = "JOIN"
	partMsg  = "PART"
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

	handlers := map[string]func(*Message) string{
		privMsg:  p.formatPrivateMsg,
		nickMsg:  p.formatNickChange,
		namesMsg: p.formatNames,
		pingMsg:  p.handlePing,
		joinMsg:  p.formatJoin,
		partMsg:  p.formatPart,
	}

	if handler, exists := handlers[message.Command]; exists {
		return handler(message)
	}

	return message.printMessage()
}

// handlePing handles PING messages from the server, and sends a PONG message back
// to keep the connection alive
func (p *MessageParser) handlePing(msg *Message) string {
	p.Client.SendPong(msg.Args[0])
	return ""
}

// FormatNames parse the names of the users in a channel
func (p *MessageParser) formatNames(msg *Message) string {
	names := msg.Args[3]
	for _, name := range strings.Split(names, " ") {
		p.Client.SetUsers(append(p.Client.Users(), name))
		p.Ui.UsersView.AddItem(name, "", 0, nil)
	}
	p.Ui.App.Draw()

	return ""
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

func (p *MessageParser) formatJoin(msg *Message) string {
	p.Ui.UsersView.Clear()
	return msg.Source + " has joined the channel " + msg.Args[0]
}

func (p *MessageParser) formatPart(msg *Message) string {
	p.Ui.UsersView.Clear()
	return msg.Source + " has left the channel " + msg.Args[0]
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

	// Check if the input string is empty
	if s == "" {
		return nil, errors.New("empty string")
	}

	// If the message starts with a ':', it has a prefix
	if s[0] == ':' {
		// Split the string into prefix and the rest of the message
		split := strings.SplitN(s[1:], " ", 2)
		prefix = split[0]
		s = split[1]
	}

	// Check if the message contains a trailing part after ' :'
	if strings.Contains(s, " :") {
		// Split the string into the main part and the trailing part
		split := strings.SplitN(s, " :", 2)
		s = split[0]
		trailing = split[1]
		// Split the main part into arguments and append the trailing part as the last argument
		args = strings.Split(s, " ")
		args = append(args, trailing)
	} else {
		// If there is no trailing part, split the entire string into arguments
		args = strings.Split(s, " ")
	}

	// The first argument is the command
	command := args[0]
	// The rest are the arguments
	args = args[1:]

	// Return the parsed message
	return &Message{
		Source:  prefix,
		Command: command,
		Args:    args,
	}, nil
}
