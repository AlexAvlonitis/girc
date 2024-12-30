package commands

import (
	"errors"
	"girc/connection"
	"strings"
)

type JoinCommand struct {
	Input  string
	Client *connection.Client
}

func (c *JoinCommand) Execute() {
	cmd, err := c.Print()
	if err != nil {
		c.Client.PrintMessage(err.Error())
		return
	}

	c.Client.Write(cmd)
	c.Client.Channel = strings.Split(c.Input, " ")[1]
}

func (c *JoinCommand) Print() (string, error) {
	parts := strings.Split(c.Input, " ")

	if len(parts) > 1 {
		channel := parts[1]
		return "JOIN :" + channel + "\r\n", nil
	} else {
		return "", errors.New("invalid command, use /join #channel")
	}
}
