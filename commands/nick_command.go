package commands

import (
	"errors"
	"girc/interfaces"
	"strings"
)

type NickCommand struct {
	Input  string
	Client interfaces.Client
}

func (n *NickCommand) Execute() error {
	cmd, err := n.BuildCommand()
	if err != nil {
		return err
	}

	n.Client.Write(cmd)

	return nil
}

func (n *NickCommand) BuildCommand() (string, error) {
	parts := strings.Fields(n.Input)

	if len(parts) > 1 {
		args := strings.Join(parts[1:], " ")
		return "NICK " + args + "\r\n", nil
	}

	return "", errors.New("invalid command, use /nick newnick")
}
