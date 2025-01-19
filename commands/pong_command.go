package commands

import "girc/app/connection_manager"

type PongCommand struct {
	Input             string
	ConnectionManager *connection_manager.DefaultConnectionManager
}

func (c *PongCommand) Execute() error {
	cmd, _ := c.BuildCommand()
	c.ConnectionManager.Client.Write(cmd)

	return nil
}

func (c *PongCommand) BuildCommand() (string, error) {
	cmd := "PONG " + c.Input

	return cmd, nil
}
