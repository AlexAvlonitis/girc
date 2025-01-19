package commands

import (
	"errors"
	"girc/app/connection_manager"
	"strings"
)

type PersonalMessageCommand struct {
	Input             string
	ConnectionManager *connection_manager.DefaultConnectionManager
}

func (c *PersonalMessageCommand) Execute() error {
	cmd, err := c.BuildCommand()
	if err != nil {
		return err
	}

	c.ConnectionManager.Client.Write(cmd)
	// irc does not echo personal messages
	c.ConnectionManager.Client.PrintMessage(":" + c.ConnectionManager.Client.Nick() + " " + cmd)

	return nil
}

func (c *PersonalMessageCommand) BuildCommand() (string, error) {
	if c.ConnectionManager.Client.Channel() == "" {
		return "", errors.New("you need to join a channel first")
	}

	parts := strings.Split(c.Input, " ")
	if len(parts) > 2 {
		cmd := "PRIVMSG " + parts[1] + " :" + strings.Join(parts[2:], " ") + "\r\n"
		return cmd, nil
	} else {
		return "", errors.New("invalid command, use /msg nickname message")
	}
}
