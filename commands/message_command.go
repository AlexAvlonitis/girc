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
	cmd, err := m.Print()
	if err != nil {
		m.Client.PrintMessage(err.Error())
		return err
	}

	m.Client.Write(cmd)
	m.Client.PrintMessage("<" + m.Client.Nick() + "> " + m.Input)

	return nil
}

func (m *MessageCommand) Print() (string, error) {
	if m.Client.Channel() != "" {
		cmd := "PRIVMSG " + m.Client.Channel() + " :" + m.Input + "\r\n"
		return cmd, nil
	} else {
		return "", errors.New("you need to join a channel first")
	}
}
