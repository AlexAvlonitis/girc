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
	cmd, _ := c.Print()
	c.Client.Write(cmd)

	return nil
}

func (c *QuitCommand) Print() (string, error) {
	parts := strings.Split(c.Input, " ")

	var cmd string
	if len(parts) > 1 {
		cmd = "QUIT :" + parts[1] + "\r\n"
	} else {
		cmd = "QUIT :Bye bye\r\n"
	}

	return cmd, nil
}
