package commands

import (
	"errors"
	"girc/connection"
)

type MessageCommand struct {
	Input  string
	Client *connection.Client
}

func (c *MessageCommand) Execute() {
	cmd, err := c.Print()
	if err != nil {
		c.Client.PrintMessage(err.Error())
		return
	}

	c.Client.Write(cmd)
	c.Client.PrintMessage("<" + c.Client.Nick + "> " + c.Input)
}

func (c *MessageCommand) Print() (string, error) {
	if c.Client.Channel != "" {
		cmd := "PRIVMSG " + c.Client.Channel + " :" + c.Input + "\r\n"
		return cmd, nil
	} else {
		return "", errors.New("you need to join a channel first")
	}
}
