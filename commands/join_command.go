package commands

import (
	"errors"
	"girc/app/connection_manager"
	"strings"
)

type JoinCommand struct {
	Input             string
	ConnectionManager *connection_manager.DefaultConnectionManager
}

func (c *JoinCommand) Execute() error {
	cmd, err := c.BuildCommand()
	if err != nil {
		return err
	}

	c.ConnectionManager.Client.Write(cmd)

	return nil
}

func (c *JoinCommand) BuildCommand() (string, error) {
	parts := strings.Split(c.Input, " ")

	if len(parts) > 1 {
		args := strings.Join(parts[1:], " ")
		return "JOIN " + args + "\r\n", nil
	} else {
		return "", errors.New("invalid command, use /join #channel")
	}
}
