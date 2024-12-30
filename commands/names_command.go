package commands

import (
	"girc/connection"
	"strings"
)

type NamesCommand struct {
	Input  string
	Client *connection.Client
}

func (c *NamesCommand) Execute() {
	parts := strings.Split(c.Input, " ")

	if len(parts) > 1 {
		cmd := "NAMES " + parts[1] + "\r\n"
		c.Client.Write(cmd)
	} else {
		c.Client.PrintMessage("Invalid command, use /names #channel")
	}
}
