package commands

import (
	"girc/connection"
	"strings"
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
		c.Client.PrintMessage("<" + c.Client.Nick + "> " + c.Input)
	} else {
		c.Client.PrintMessage("You need to join a channel first")
	}
}

type PersonalMessageCommand struct {
	Input  string
	Client *connection.Client
}

func (c *PersonalMessageCommand) Execute() {
	// check if the channel is set and send the message, otherwise print an error
	if c.Client.Channel != "" {
		parts := strings.Split(c.Input, " ")
		if len(parts) > 2 {
			cmd := "PRIVMSG " + parts[1] + " :" + strings.Join(parts[2:], " ") + "\r\n"
			c.Client.Write(cmd)
			c.Client.PrintMessage("<" + c.Client.Nick + ">(Private) " + strings.Join(parts[2:], " "))
		} else {
			c.Client.PrintMessage("Invalid command, use /msg nickname message")
		}
	} else {
		c.Client.PrintMessage("You need to join a channel first")
	}
}
