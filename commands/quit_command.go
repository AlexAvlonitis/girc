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
	parts := strings.Split(c.Input, " ")

	if len(parts) > 1 {
		cmd := "QUIT :" + parts[1] + "\r\n"
		c.Client.SendCommand(cmd)
	} else {
		cmd := "QUIT :Bye bye\r\n"
		c.Client.SendCommand(cmd)
	}
}
