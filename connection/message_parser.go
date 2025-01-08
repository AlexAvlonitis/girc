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
		privMsg:  p.privMsgCallback,
		nickMsg:  p.nickCallback,
		namesMsg: p.namesCallback,
		pingMsg:  p.pingCallback,
		joinMsg:  p.joinCallback,
		partMsg:  p.partCallback,
	}

	if handler, exists := handlers[message.Command]; exists {
		return handler(message)
	}

	return p.genericMsgCallback(message)
}

// pingCallback handles PING messages from the server, and sends a PONG message back
// to keep the connection alive
func (p *MessageParser) pingCallback(msg *Message) string {
	p.Client.SendPong(msg.Args[0])

	return ""
}

// namesCallback parse the names of the users in a channel
func (p *MessageParser) namesCallback(msg *Message) string {
	names := msg.Args[3]
	for _, name := range strings.Split(names, " ") {
		p.Client.SetUsers(append(p.Client.Users(), name))
		p.Ui.UsersView.AddItem(name, "", 0, nil)
	}
	p.Ui.App.Draw()

	return ""
}

func (p *MessageParser) nickCallback(msg *Message) string {
	newNick := msg.Args[0]
	p.Client.SetNick(newNick)

	return "[green]" + msg.Source + "[-] is now known as " + newNick
}

func (p *MessageParser) privMsgCallback(msg *Message) string {
	// check if the message is a private message and directed to the user
	if msg.Args[0] == p.Client.Nick() {
		return "[green]<" + msg.Source + ">[-][yellow](Private)[-] " + msg.Args[1]
	}

	return "[green]<" + msg.Source + ">[-] " + msg.Args[1]
}

func (p *MessageParser) joinCallback(msg *Message) string {
	p.Ui.UsersView.Clear()
	p.Client.SetChannel(msg.Args[0])
	p.Client.SetUsers([]string{})

	return "[green]" + msg.Source + "[-] has joined the channel " + msg.Args[0]
}

func (p *MessageParser) partCallback(msg *Message) string {
	p.Ui.UsersView.Clear()
	p.Client.SetUsers([]string{})
	p.Client.SetChannel("")

	return "[green]" + msg.Source + "[-] has left the channel " + msg.Args[0]
}

func (p *MessageParser) genericMsgCallback(msg *Message) string {
	return "[green]<" + p.Client.Nick() + ">[-] " + strings.Join(msg.Args, " ")
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
