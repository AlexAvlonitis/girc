package commands

import (
	"girc/connection"
)

type PartCommand struct {
	Input  string
	Client *connection.Client
}

func (c *PartCommand) Execute() {
	if c.Client.Channel == "" {
		c.Client.PrintMessage("You have not joined that channel")
		return
	}

	cmd := "PART " + c.Client.Channel + "\r\n"
	c.Client.Write(cmd)
	c.Client.Channel = ""
}
