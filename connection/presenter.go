package connection

import "strings"

type Presenter struct {
	Client *Client
}

func NewPresenter(c *Client) *Presenter {
	return &Presenter{Client: c}
}

// FormatMessage formats the message received from the server
func (p *Presenter) FormatMessage(msg []byte) string {
	s := string(msg)

	// check if the message is a private message and directed to the user
	if strings.Contains(s, "PRIVMSG") && strings.Contains(s, p.Client.Nick) {
		return p.formatPrivateMessage(s)
	} else if strings.Contains(s, "PRIVMSG") {
		return p.FormatMessageToChannel(s)
	} else if strings.Contains(s, "NICK") {
		return p.formatNickChange(s)
	}

	return s
}

// format private messages directed to the user
func (p *Presenter) formatPrivateMessage(msg string) string {
	parts := strings.Split(msg, " ")
	nick := strings.Split(parts[0], "!")[0][1:]
	msg = strings.Join(parts[3:], " ")[1:]
	return "<" + nick + ">" + "(Private) " + msg
}

func (p *Presenter) FormatMessageToChannel(msg string) string {
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
