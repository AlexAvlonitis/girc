package commands

import (
	"fmt"
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
		c.Client.Write(cmd)
	} else {
		fmt.Println("Join a channel /join #channel, or /help for more commands")
	}
}
