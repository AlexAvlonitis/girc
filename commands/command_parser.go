package commands

import (
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
	// Check if the input is a command
	if strings.HasPrefix(input, "/") {
		parts := strings.Split(input, " ")

		switch parts[0] {
		case "/join":
			cmd = &JoinCommand{Input: input, Client: client}
		case "/part":
			cmd = &PartCommand{Input: input, Client: client}
		case "/nick":
			cmd = &NickCommand{Input: input, Client: client}
		case "/quit", "/exit", "/bye", "/q":
			cmd = &QuitCommand{Input: input, Client: client}
		case "/help":
			cmd = &HelpCommand{Client: client}
		default:
			client.PrintMessage("Invalid command, use /help for more commands")
			return
		}

		CommandExecute(cmd)
		return
	}

	cmd = &MessageCommand{Input: input, Client: client}
	CommandExecute(cmd)
}
