package commands

import (
	"girc/connection"
	"strings"
)

type QuitCommand struct {
	Input  string
	Client *connection.Client
}

func (c *QuitCommand) Execute() {
	cmd, _ := c.Print()
	c.Client.Write(cmd)
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
