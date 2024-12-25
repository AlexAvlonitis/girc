package main

import (
	"fmt"
	"strings"
)

type JoinCommand struct {
	Input  string
	Client *Client
}

func (c *JoinCommand) Execute() {
	parts := strings.Split(c.Input, " ")

	if len(parts) > 1 {
		channel := parts[1]
		cmd := "JOIN :" + channel + "\r\n"
		c.Client.Write(cmd)
		c.Client.Channel = channel
	} else {
		fmt.Println("\nInvalid command, use /join #channel")
	}
}
