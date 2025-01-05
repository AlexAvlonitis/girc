package commands

import (
	"girc/interfaces"
)

type PongCommand struct {
	Input  string
	Client interfaces.Client
}

func (p *PongCommand) Execute() error {
	cmd, _ := p.BuildCommand()
	p.Client.Write(cmd)

	return nil
}

func (p *PongCommand) BuildCommand() (string, error) {
	cmd := "PONG " + p.Input
	return cmd, nil
}
