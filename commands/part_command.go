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
		c.Client.PrintMessage("\nYou have not joined a channel")
		return
	}

	cmd := "PART " + c.Client.Channel + "\r\n"
	c.Client.SendCommand(cmd)
	c.Client.Channel = ""
}
