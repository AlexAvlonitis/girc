package commands

import (
	"fmt"
	"girc/connection"
	"strings"
)

type NickCommand struct {
	Input  string
	Client *connection.Client
}

func (c *NickCommand) Execute() {
	parts := strings.Split(c.Input, " ")

	if len(parts) > 1 {
		cmd := "NICK :" + parts[1] + "\r\n"
		c.Client.Write(cmd)
	} else {
		fmt.Println("\nInvalid command, use /nick newnick")
	}
}
