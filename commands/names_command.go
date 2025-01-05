package commands

import (
	"errors"
	"girc/interfaces"
	"strings"
)

type NamesCommand struct {
	Input  string
	Client interfaces.Client
}

func (n *NamesCommand) Execute() error {
	cmd, err := n.Print()
	if err != nil {
		n.Client.PrintMessage(err.Error())
		return err
	}

	n.Client.Write(cmd)

	return nil
}

func (n *NamesCommand) Print() (string, error) {
	parts := strings.Split(n.Input, " ")

	if len(parts) > 1 {
		cmd := "NAMES " + parts[1] + "\r\n"
		return cmd, nil
	} else if n.Client.Channel() != "" {
		return "NAMES " + n.Client.Channel() + "\r\n", nil
	} else {
		return "", errors.New("invalid command, use /names #channel")
	}
}
