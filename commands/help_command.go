package commands

import (
	"girc/connection"
)

type HelpCommand struct {
	Client *connection.Client
}

func (c *HelpCommand) Execute() {
	cmd, _ := c.Print()
	c.Client.PrintMessage(cmd)
}

func (c *HelpCommand) Print() (string, error) {
	msg := "Commands:\n"
	msg += "/join #channel - join a channel\n"
	msg += "/part #channel - leave a channel\n"
	msg += "/nick newnick - change your nickname\n"
	msg += "/msg nickname message - send a private message\n"
	msg += "/quit - quit the server\n"

	return msg, nil
}
