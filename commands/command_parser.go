package commands

import (
	"girc/interfaces"
	"strings"
)

// SendCommand parses a command and sends the appropriate IRC command
func SendCommand(input string, client interfaces.Client) {
	var cmd interfaces.Command

	commandMap := map[string]interfaces.Command{
		"/join":  &JoinCommand{Input: input, Client: client},
		"/part":  &PartCommand{Input: input, Client: client},
		"/nick":  &NickCommand{Input: input, Client: client},
		"/msg":   &PersonalMessageCommand{Input: input, Client: client},
		"/names": &NamesCommand{Input: input, Client: client},
		"/user":  &UserCommand{Input: input, Client: client},
		"/quit":  &QuitCommand{Input: input, Client: client},
		"/help":  &HelpCommand{Client: client},
	}

	// Check if the input is a command
	if strings.HasPrefix(input, "/") {
		parts := strings.Split(input, " ")
		if command, exists := commandMap[parts[0]]; exists {
			cmd = command
		} else {
			client.PrintMessage("Invalid command, use /help for more commands")
			return
		}
	} else {
		cmd = &MessageCommand{Input: input, Client: client}
	}

	err := cmd.Execute()
	if err != nil {
		// Custom error that can be parsed by msg parser
		client.PrintMessage(":error 999 e :" + err.Error())
	}
}
