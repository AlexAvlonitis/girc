package commands

import (
	"errors"
	"girc/app/connection_manager"
)

type MessageCommand struct {
	Input             string
	ConnectionManager *connection_manager.DefaultConnectionManager
}

func (c *MessageCommand) Execute() error {
	cmd, err := c.BuildCommand()
	if err != nil {
		return err
	}

	client := c.ConnectionManager.Client
	client.Write(cmd)
	// irc server does not echo messages
	client.PrintMessage(":" + client.Nick() + " " + cmd)

	return nil
}

func (c *MessageCommand) BuildCommand() (string, error) {
	client := c.ConnectionManager.Client
	if client.Channel() == "" {
		return "", errors.New("you need to join a channel first")
	}

	cmd := "PRIVMSG " + client.Channel() + " :" + c.Input + "\r\n"
	return cmd, nil
}
