package commands

import (
	"errors"
	"girc/connection"
	"strings"
)

type NickCommand struct {
	Input  string
	Client *connection.Client
}

func (c *NickCommand) Execute() {
	cmd, err := c.Print()
	if err != nil {
		c.Client.PrintMessage(err.Error())
		return
	}

	c.Client.Write(cmd)
	c.Client.Nick = strings.Split(c.Input, " ")[1]
}

func (c *NickCommand) Print() (string, error) {
	parts := strings.Split(c.Input, " ")

	if len(parts) > 1 {
		cmd := "NICK :" + parts[1] + "\r\n"
		return cmd, nil
	} else {
		return "", errors.New("invalid command, use /nick newnick")
	}
}
