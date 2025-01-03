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
	cmd, err := n.Print()
	if err != nil {
		n.Client.PrintMessage(err.Error())
		return err
	}

	n.Client.Write(cmd)
	return nil
}

func (n *NickCommand) Print() (string, error) {
	parts := strings.Split(n.Input, " ")

	if len(parts) > 1 {
		cmd := "NICK " + parts[1] + "\r\n"
		n.Client.SetNick(parts[1])
		return cmd, nil
	} else if n.Client.Nick() != "" {
		return "NICK " + n.Client.Nick() + "\r\n", nil
	} else {
		return "", errors.New("invalid command, use /nick newnick")
	}
}
