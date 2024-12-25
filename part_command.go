package main

import "fmt"

type PartCommand struct {
	Input  string
	Client *Client
}

func (c *PartCommand) Execute() {
	if c.Client.Channel == "" {
		fmt.Println("\nYou are not in a channel")
		return
	}

	cmd := "PART " + c.Client.Channel + "\r\n"
	c.Client.Write(cmd)
	c.Client.Channel = ""
}
