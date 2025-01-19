package commands

import (
	"errors"
	"girc/app/connection_manager"
	"strings"
)

type PartCommand struct {
	Input             string
	ConnectionManager *connection_manager.DefaultConnectionManager
}

func (c *PartCommand) Execute() error {
	cmd, err := c.BuildCommand()
	if err != nil {
		return err
	}

	c.ConnectionManager.Client.Write(cmd)

	return nil
}

func (c *PartCommand) BuildCommand() (string, error) {
	parts := strings.Fields(c.Input)

	if len(parts) > 1 {
		args := strings.Join(parts[1:], " ")
		return "PART " + args + "\r\n", nil
	}

	return "", errors.New("invalid command, use /part #channel")
}
