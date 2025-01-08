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
		m.Client.PrintMessage(err.Error())
		return err
	}

	m.Client.Write(cmd)
	// m.Client.PrintMessage("<" + m.Client.Nick() + "> " + m.Input)

	return nil
}

func (m *MessageCommand) BuildCommand() (string, error) {
	if m.Client.Channel() == "" {
		return "", errors.New("you need to join a channel first")
	}

	cmd := "PRIVMSG " + m.Client.Channel() + " :" + m.Input + "\r\n"
	return cmd, nil
}
