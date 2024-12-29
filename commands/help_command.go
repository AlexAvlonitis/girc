package commands

import (
	"girc/connection"
)

type HelpCommand struct {
	Client *connection.Client
}

func (c *HelpCommand) Execute() {
	msg := "Commands:\n"
	msg += "/join #channel - join a channel\n"
	msg += "/part #channel - leave a channel\n"
	msg += "/nick newnick - change your nickname\n"
	msg += "/quit - quit the server\n"

	c.Client.PrintMessage(msg)
}
