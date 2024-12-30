package commands

import (
	"errors"
	"girc/connection"
)

type PartCommand struct {
	Input  string
	Client *connection.Client
}

func (c *PartCommand) Execute() {
	cmd, err := c.Print()
	if err != nil {
		c.Client.PrintMessage(err.Error())
		return
	}

	c.Client.Write(cmd)
	c.Client.Channel = ""
}

func (c *PartCommand) Print() (string, error) {
	if c.Client.Channel == "" {
		return "", errors.New("you have not joined that channel")
	}

	return "PART " + c.Client.Channel + "\r\n", nil
}
