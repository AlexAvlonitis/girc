package commands

import (
	"fmt"
	"girc/connection"
	"strings"
)

type Command interface {
	Execute()
}

func CommandExecute(cmd Command) {
	cmd.Execute()
}

// ParseCommand parses a command and sends the appropriate IRC command
func SendCommand(input string, client *connection.Client) {
	var cmd Command

	if strings.HasPrefix(input, "/") {
		parts := strings.Split(input, " ")

		switch parts[0] {
		case "/join":
			cmd = &JoinCommand{Input: input, Client: client}
		case "/part":
			cmd = &PartCommand{Input: input, Client: client}
		case "/nick":
			cmd = &NickCommand{Input: input, Client: client}
		case "/quit":
			cmd = &QuitCommand{Input: input, Client: client}
		case "/help":
			cmd = &HelpCommand{}
		default:
			fmt.Println("\nInvalid command, use /help for more commands")
			return
		}

		CommandExecute(cmd)
		return
	}

	cmd = &MessageCommand{Input: input, Client: client}
	CommandExecute(cmd)
}
