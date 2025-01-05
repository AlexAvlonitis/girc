package commands

import (
	"errors"
	"girc/interfaces"
	"strings"
)

type PersonalMessageCommand struct {
	Input  string
	Client interfaces.Client
}

func (c *PersonalMessageCommand) Execute() error {
	cmd, err := c.BuildCommand()
	if err != nil {
		c.Client.PrintMessage(err.Error())
		return err
	}

	c.Client.Write(cmd)
	c.Client.PrintMessage(":" + c.Client.Nick() + " " + cmd)

	return nil
}

func (c *PersonalMessageCommand) BuildCommand() (string, error) {
	if c.Client.Channel() != "" {
		parts := strings.Split(c.Input, " ")
		if len(parts) > 2 {
			cmd := "PRIVMSG " + parts[1] + " :" + strings.Join(parts[2:], " ") + "\r\n"
			return cmd, nil
		} else {
			return "", errors.New("invalid command, use /msg nickname message")
		}
	} else {
		return "", errors.New("you need to join a channel first")
	}
}
