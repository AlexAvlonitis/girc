package main

import (
	"fmt"
	"strings"
)

type Command interface {
	Execute()
}

func CommandExecute(cmd Command) {
	cmd.Execute()
}

// ParseCommand parses a command and sends the appropriate IRC command
func SendCommand(input string, client *Client) {
	if strings.HasPrefix(input, "/") {
		parts := strings.Split(input, " ")

		switch parts[0] {
		case "/join":
			cmd := &JoinCommand{Input: input, Client: client}
			CommandExecute(cmd)
		case "/part":
			cmd := &PartCommand{Input: input, Client: client}
			CommandExecute(cmd)
		case "/nick":
			cmd := &NickCommand{Input: input, Client: client}
			CommandExecute(cmd)
		case "/quit":
			cmd := &QuitCommand{Input: input, Client: client}
			CommandExecute(cmd)
		case "/help":
			cmd := &HelpCommand{}
			CommandExecute(cmd)
		default:
			fmt.Println("\nInvalid command, use /help for more commands")
		}
		return
	}

	cmd := &MessageCommand{Input: input, Client: client}
	CommandExecute(cmd)
}
