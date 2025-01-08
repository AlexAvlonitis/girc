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
	cmd, err := n.BuildCommand()
	if err != nil {
		n.Client.PrintMessage(err.Error())
		return err
	}

	n.Client.Write(cmd)

	return nil
}

func (n *NamesCommand) BuildCommand() (string, error) {
	parts := strings.Fields(n.Input)

	if len(parts) > 1 {
		args := strings.Join(parts[1:], " ")
		return "NAMES " + args + "\r\n", nil
	}

	return "", errors.New("invalid command, use /names #channel")
}
