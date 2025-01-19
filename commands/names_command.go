package commands

import (
	"errors"
	"girc/app/connection_manager"
	"strings"
)

type NamesCommand struct {
	Input             string
	ConnectionManager *connection_manager.DefaultConnectionManager
}

func (c *NamesCommand) Execute() error {
	cmd, err := c.BuildCommand()
	if err != nil {
		return err
	}

	c.ConnectionManager.Client.Write(cmd)

	return nil
}

func (n *NamesCommand) BuildCommand() (string, error) {
	parts := strings.Fields(n.Input)

	if len(parts) > 1 {
		args := strings.Join(parts[1:], " ")
		return "NAMES " + args + "\r\n", nil
	}

	return "", errors.New("invalid command, use /names #channel")
}
