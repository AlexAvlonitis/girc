package connection

import "strings"

type Presenter struct{}

func NewPresenter() *Presenter {
	return &Presenter{}
}

// FormatMessage formats the message received from the server
func (p *Presenter) FormatMessage(msg []byte) string {
	s := string(msg)

	if strings.Contains(s, "PRIVMSG") {
		return p.formatPrivateMessage(s)
	}

	return s
}

// format private messages
func (p *Presenter) formatPrivateMessage(msg string) string {
	parts := strings.Split(msg, " ")
	nick := strings.Split(parts[0], "!")[0][1:]
	msg = strings.Join(parts[3:], " ")[1:]
	return "<" + nick + "> " + msg
}
