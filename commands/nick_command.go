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
		n.Client.PrintMessage(err.Error())
		return err
	}

	n.Client.Write(cmd)
	return nil
}

func (n *NickCommand) BuildCommand() (string, error) {
	parts := strings.Fields(n.Input)

	if len(parts) > 1 {
		newNick := parts[1]
		n.Client.SetNick(newNick)
		return "NICK " + newNick + "\r\n", nil
	}

	currentNick := n.Client.Nick()
	if currentNick != "" {
		return "NICK " + currentNick + "\r\n", nil
	}

	return "", errors.New("invalid command, use /nick newnick")
}
