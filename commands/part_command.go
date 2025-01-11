package commands

import (
	"errors"
	"girc/interfaces"
	"strings"
)

type PartCommand struct {
	Input  string
	Client interfaces.Client
}

func (c *PartCommand) Execute() error {
	cmd, err := c.BuildCommand()
	if err != nil {
		return err
	}

	c.Client.Write(cmd)

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
