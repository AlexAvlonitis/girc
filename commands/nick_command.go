package commands

import (
	"girc/connection"
	"strings"
)

type NickCommand struct {
	Input  string
	Client *connection.Client
}

func (c *NickCommand) Execute() {
	parts := strings.Split(c.Input, " ")

	if len(parts) > 1 {
		cmd := "NICK :" + parts[1] + "\r\n"
		c.Client.Write(cmd)
		c.Client.Nick = parts[1]
	} else {
		c.Client.PrintMessage("Invalid command, use /nick newnick")
	}
}
