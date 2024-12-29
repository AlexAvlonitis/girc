package commands

import (
	"girc/connection"
	"strings"
)

type JoinCommand struct {
	Input  string
	Client *connection.Client
}

func (c *JoinCommand) Execute() {
	parts := strings.Split(c.Input, " ")

	if len(parts) > 1 {
		channel := parts[1]
		cmd := "JOIN :" + channel + "\r\n"
		c.Client.Write(cmd)
		c.Client.Channel = channel
	} else {
		c.Client.PrintMessage("Invalid command, use /join #channel")
	}
}
