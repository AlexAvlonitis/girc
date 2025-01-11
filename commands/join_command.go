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
	cmd, err := j.BuildCommand()
	if err != nil {
		return err
	}

	j.Client.Write(cmd)

	return nil
}

func (j *JoinCommand) BuildCommand() (string, error) {
	parts := strings.Split(j.Input, " ")

	if len(parts) > 1 {
		args := strings.Join(parts[1:], " ")
		return "JOIN " + args + "\r\n", nil
	} else {
		return "", errors.New("invalid command, use /join #channel")
	}
}
