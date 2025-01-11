package commands

import (
	"errors"
	"girc/interfaces"
)

type MessageCommand struct {
	Input  string
	Client interfaces.Client
}

func (m *MessageCommand) Execute() error {
	cmd, err := m.BuildCommand()
	if err != nil {
		return err
	}

	m.Client.Write(cmd)
	// irc server does not echo messages
	m.Client.PrintMessage(":" + m.Client.Nick() + " " + cmd)

	return nil
}

func (m *MessageCommand) BuildCommand() (string, error) {
	if m.Client.Channel() == "" {
		return "", errors.New("you need to join a channel first")
	}

	cmd := "PRIVMSG " + m.Client.Channel() + " :" + m.Input + "\r\n"
	return cmd, nil
}
