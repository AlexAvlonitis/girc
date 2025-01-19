package commands

import (
	"errors"
	"girc/app/connection_manager"
	"strings"
)

type NickCommand struct {
	Input             string
	ConnectionManager *connection_manager.DefaultConnectionManager
}

func (c *NickCommand) Execute() error {
	cmd, err := c.BuildCommand()
	if err != nil {
		return err
	}

	c.ConnectionManager.Client.Write(cmd)

	return nil
}

func (n *NickCommand) BuildCommand() (string, error) {
	parts := strings.Fields(n.Input)

	if len(parts) > 1 {
		args := strings.Join(parts[1:], " ")
		return "NICK " + args + "\r\n", nil
	}

	return "", errors.New("invalid command, use /nick newnick")
}
