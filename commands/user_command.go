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
	cmd, err := u.BuildCommand()
	if err != nil {
		u.Client.PrintMessage(err.Error())
		return err
	}

	u.Client.Write(cmd)

	return nil
}

func (u *UserCommand) BuildCommand() (string, error) {
	parts := strings.Fields(u.Input)

	if len(parts) > 1 {
		args := strings.Join(parts[1:], " ")
		return "USER " + args + "\r\n", nil
	}

	return "", errors.New("invalid command, use /user user 0 * :realname")
}
