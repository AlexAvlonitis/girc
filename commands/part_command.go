package commands

import (
	"errors"
	"girc/interfaces"
)

type PartCommand struct {
	Input  string
	Client interfaces.Client
}

func (c *PartCommand) Execute() error {
	cmd, err := c.Print()
	if err != nil {
		c.Client.PrintMessage(err.Error())
		return err
	}

	c.Client.Write(cmd)
	c.Client.SetChannel("")

	return nil
}

func (c *PartCommand) Print() (string, error) {
	if c.Client.Channel() == "" {
		return "", errors.New("you have not joined that channel")
	}

	return "PART " + c.Client.Channel() + "\r\n", nil
}
