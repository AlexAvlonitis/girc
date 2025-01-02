package commands

import (
	"girc/interfaces"
)

type PongCommand struct {
	Input  string
	Client interfaces.Client
}

func (p *PongCommand) Execute() error {
	cmd, _ := p.Print()
	p.Client.Write(cmd)

	return nil
}

func (p *PongCommand) Print() (string, error) {
	cmd := "PONG " + p.Input
	return cmd, nil
}
