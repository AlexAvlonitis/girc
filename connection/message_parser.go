package connection

import (
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
	message := parseMsg(msg)

	handlers := map[string]func(*Message) string{
		privMsg:  p.privMsgCallback,
		nickMsg:  p.nickCallback,
		joinMsg:  p.joinCallback,
		namesMsg: p.namesCallback,
		pingMsg:  p.pingCallback,
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

	return "[green]<" + msg.Source + ">[-] " + strings.Join(msg.Args[1:], " ")
}

func (p *MessageParser) joinCallback(msg *Message) string {
	channel := strings.TrimPrefix(msg.Args[0], ":")

	p.Ui.UsersView.Clear()
	p.Client.SetChannel(channel)
	p.Client.SetUsers([]string{})

	return "[green]<" + msg.Source + ">[-] has joined the channel " + channel
}

func (p *MessageParser) partCallback(msg *Message) string {
	p.Ui.UsersView.Clear()
	p.Client.SetUsers([]string{})
	p.Client.SetChannel("")

	return "[green]<" + msg.Source + ">[-] has left the channel " + msg.Args[0]
}

func (p *MessageParser) genericMsgCallback(msg *Message) string {
	if len(msg.Args) > 0 {
		return "[green]<" + msg.Source + ">[-] " + strings.Join(msg.Args[1:], " ")
	}

	return "[green]<" + p.Client.Nick() + ">[-] " + strings.Join(msg.Args, " ")
}

// parseMsg Breaks a message from an IRC server into its prefix, command, and arguments
// parsemsg(":test!~test@test.com PRIVMSG #channel :Hi!")
// Message('test!~test@test.com', 'PRIVMSG', ['#channel', 'Hi!'])
func parseMsg(s string) *Message {
	// Trim any extraneous spaces or newlines
	s = strings.TrimSpace(s)

	// Initialize the result
	message := &Message{}

	// Check for a prefix (starts with ":")
	if strings.HasPrefix(s, ":") {
		// Split off the prefix
		split := strings.SplitN(s[1:], " ", 2)
		message.Source = split[0]
		if len(split) > 1 {
			s = split[1]
		} else {
			s = ""
		}
	}

	// Extract the command
	parts := strings.SplitN(s, " ", 2)
	message.Command = parts[0]
	if len(parts) > 1 {
		s = parts[1]
	} else {
		s = ""
	}

	// Handle parameters and trailing arguments
	if strings.Contains(s, " :") {
		// Split middle parameters and trailing
		middleAndTrailing := strings.SplitN(s, " :", 2)
		if len(middleAndTrailing[0]) > 0 {
			message.Args = append(message.Args, strings.Fields(middleAndTrailing[0])...)
		}
		message.Args = append(message.Args, middleAndTrailing[1])
	} else if len(s) > 0 {
		// Only middle parameters (no trailing)
		message.Args = strings.Fields(s)
	}

	return message
}
