package commands

import "girc/app/connection_manager"

type HelpCommand struct {
	ConnectionManager *connection_manager.DefaultConnectionManager
}

func (c *HelpCommand) Execute() error {
	cmd, _ := c.BuildCommand()
	c.ConnectionManager.Client.PrintMessage(cmd)

	return nil
}

func (h *HelpCommand) BuildCommand() (string, error) {
	msg := "Commands:\n"
	msg += "/join #channel - join a channel\n"
	msg += "/part #channel - leave a channel\n"
	msg += "/nick newnick - change your nickname\n"
	msg += "/msg nickname message - send a private message\n"
	msg += "/quit - quit the server\n"

	return msg, nil
}
