package commands

import (
	"errors"
	"girc/interfaces"
	"strings"
)

type UserCommand struct {
	Input  string
	Client interfaces.Client
}

func (u *UserCommand) Execute() error {
	cmd, err := u.Print()
	if err != nil {
		u.Client.PrintMessage(err.Error())
		return err
	}

	u.Client.Write(cmd)
	u.Client.SetNick(strings.Split(u.Input, " ")[1])
	return nil
}

func (u *UserCommand) Print() (string, error) {
	parts := strings.Split(u.Input, " ")

	if len(parts) > 1 {
		cmd := "USER " + u.Client.User() + " 0 * :" + u.Client.RealName() + "\r\n"
		return cmd, nil
	} else if u.Client.Nick() != "" {
		return "NICK :" + u.Client.Nick() + "\r\n", nil
	} else {
		return "", errors.New("invalid command, use /nick newnick")
	}
}
