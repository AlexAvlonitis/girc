package commands

import (
	"errors"
	"girc/interfaces"
	"strings"
)

type JoinCommand struct {
	Input  string
	Client interfaces.Client
}

func (j *JoinCommand) Execute() error {
	cmd, err := j.Print()
	if err != nil {
		j.Client.PrintMessage(err.Error())
		return err
	}

	j.Client.Write(cmd)
	j.Client.SetChannel(strings.Split(j.Input, " ")[1])

	return nil
}

func (j *JoinCommand) Print() (string, error) {
	parts := strings.Split(j.Input, " ")

	if len(parts) > 1 {
		channel := parts[1]
		return "JOIN :" + channel + "\r\n", nil
	} else {
		return "", errors.New("invalid command, use /join #channel")
	}
}
