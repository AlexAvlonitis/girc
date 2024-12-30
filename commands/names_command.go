package commands

import (
	"errors"
	"girc/connection"
	"strings"
)

type NamesCommand struct {
	Input  string
	Client *connection.Client
}

func (c *NamesCommand) Execute() {
	cmd, err := c.Print()
	if err != nil {
		c.Client.PrintMessage(err.Error())
		return
	}

	c.Client.Write(cmd)
}

func (c *NamesCommand) Print() (string, error) {
	parts := strings.Split(c.Input, " ")

	if len(parts) > 1 {
		cmd := "NAMES " + parts[1] + "\r\n"
		return cmd, nil
	} else {
		return "", errors.New("invalid command, use /names #channel")
	}
}
