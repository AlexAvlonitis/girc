package main

import (
	"strings"
)

type QuitCommand struct {
	Input  string
	Client *Client
}

func (c *QuitCommand) Execute() {
	parts := strings.Split(c.Input, " ")

	if len(parts) > 1 {
		cmd := "QUIT :" + parts[1] + "\r\n"
		c.Client.Write(cmd)
	} else {
		cmd := "QUIT :Bye bye\r\n"
		c.Client.Write(cmd)
	}
}
