package connection

import (
	"girc/interfaces"
	"strings"
)

type Presenter struct {
	Client interfaces.Client
}

func NewPresenter(c interfaces.Client) *Presenter {
	return &Presenter{Client: c}
}

type Message struct {
	Content string
	Type    string
}

// FormatMessage formats the messages received from the server, and categorizes them
func (p *Presenter) FormatMessage(msg []byte) *Message {
	s := string(msg)

	// check if the message is a private message and directed to the user
	if strings.Contains(s, "PRIVMSG") && strings.Contains(s, p.Client.Nick()) {
		return &Message{Content: p.formatPrivateMessage(s), Type: "private"}
	} else if strings.Contains(s, "PRIVMSG") {
		return &Message{Content: p.formatMessageToChannel(s), Type: "channel"}
	} else if strings.Contains(s, "NICK") {
		return &Message{Content: p.formatNickChange(s), Type: "nick"}
	} else if strings.Contains(s, "JOIN") {
		return &Message{Content: s, Type: "join"}
	} else if strings.Contains(s, "PING") {
		return &Message{Content: p.formatPing(s), Type: "ping"}
	}

	return &Message{Content: s, Type: "unknown"}
}

func (p *Presenter) NamesToList(msg string) []string {
	return strings.Split(msg, " ")
}

// format private messages directed to the user
func (p *Presenter) formatPrivateMessage(msg string) string {
	parts := strings.Split(msg, " ")
	nick := strings.Split(parts[0], "!")[0][1:]
	msg = strings.Join(parts[3:], " ")[1:]
	return "<" + nick + ">" + "(Private) " + msg
}

func (p *Presenter) formatMessageToChannel(msg string) string {
	parts := strings.Split(msg, " ")
	nick := strings.Split(parts[0], "!")[0][1:]
	msg = strings.Join(parts[3:], " ")[1:]
	return "<" + nick + "> " + msg
}

// format nick changes
func (p *Presenter) formatNickChange(msg string) string {
	parts := strings.Split(msg, " ")
	oldNick := strings.Split(parts[0], "!")[0][1:]
	newNick := parts[2][1:]
	return oldNick + " is now known as " + newNick
}

// format ping messages, get the message after the PING
func (p *Presenter) formatPing(msg string) string {
	return msg[5:]
}
