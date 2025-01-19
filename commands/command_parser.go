package commands

import (
	"girc/app/connection_manager"
	"girc/interfaces"
	"strings"
)

// SendCommand parses a command and sends the appropriate IRC command
func SendCommand(input string, cm *connection_manager.DefaultConnectionManager) {
	var cmd interfaces.Command

	commandMap := map[string]interfaces.Command{
		"/join":  &JoinCommand{Input: input, ConnectionManager: cm},
		"/part":  &PartCommand{Input: input, ConnectionManager: cm},
		"/nick":  &NickCommand{Input: input, ConnectionManager: cm},
		"/msg":   &PersonalMessageCommand{Input: input, ConnectionManager: cm},
		"/names": &NamesCommand{Input: input, ConnectionManager: cm},
		"/user":  &UserCommand{Input: input, ConnectionManager: cm},
		"/pong":  &PongCommand{Input: input, ConnectionManager: cm},
		"/quit":  &QuitCommand{Input: input, ConnectionManager: cm},
		"/help":  &HelpCommand{ConnectionManager: cm},
	}

	// Check if the input is a command
	if strings.HasPrefix(input, "/") {
		parts := strings.Split(input, " ")
		if command, exists := commandMap[parts[0]]; exists {
			cmd = command
		} else {
			cm.Client.PrintMessage("Invalid command, use /help for more commands")
			return
		}
	} else {
		cmd = &MessageCommand{Input: input, ConnectionManager: cm}
	}

	err := cmd.Execute()
	if err != nil {
		// Custom error that can be parsed by msg parser
		cm.Client.PrintMessage(":error 999 e :" + err.Error())
	}
}
