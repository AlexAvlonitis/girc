package commands

import (
	"girc/interfaces"
	"strings"
)

type QuitCommand struct {
	Input  string
	Client interfaces.Client
}

func (c *QuitCommand) Execute() error {
	cmd, _ := c.BuildCommand()
	c.Client.Write(cmd)

	return nil
}

func (c *QuitCommand) BuildCommand() (string, error) {
	parts := strings.Split(c.Input, " ")

	if len(parts) > 1 {
		args := strings.Join(parts[1:], " ")
		return "QUIT " + args + "\r\n", nil
	}

	return "QUIT Bye bye\r\n", nil
}
