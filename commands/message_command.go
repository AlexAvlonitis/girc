package commands

import (
	"girc/connection"
)

type MessageCommand struct {
	Input  string
	Client *connection.Client
}

func (c *MessageCommand) Execute() {
	// check if the channel is set and send the message, otherwise print an error
	if c.Client.Channel != "" {
		cmd := "PRIVMSG " + c.Client.Channel + " :" + c.Input + "\r\n"
		c.Client.SendCommand(cmd)
		c.Client.PrintMessage("\n<" + c.Client.Nick + "> " + c.Input)
	} else {
		c.Client.PrintMessage("\nYou need to join a channel first")
	}
}