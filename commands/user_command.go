package commands

import (
	"girc/interfaces"
)

type UserCommand struct {
	Client interfaces.Client
}

func (u *UserCommand) Execute() error {
	cmd, err := u.Print()
	if err != nil {
		u.Client.PrintMessage(err.Error())
		return err
	}

	u.Client.Write(cmd)
	return nil
}

func (u *UserCommand) Print() (string, error) {
	cmd := "USER " + u.Client.User() + " 0 * :" + u.Client.RealName() + "\r\n"
	return cmd, nil
}
