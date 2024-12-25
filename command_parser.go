package main

import (
	"fmt"
	"strings"
)

// ParseCommand parses a command and sends the appropriate IRC command
// REFACTOR: create a command factory to handle the commands
func SendCommand(input string, client *Client) {
	if strings.HasPrefix(input, "/") {
		parts := strings.Split(input, " ")

		switch parts[0] {
		case "/join":
			if len(parts) > 1 {
				c := parts[1]
				cmd := "JOIN :" + c + "\r\n"
				client.Write(cmd)
				client.Channel = c
			} else {
				fmt.Println("\nInvalid command, use /join #channel")
			}
		case "/part":
			if client.Channel == "" {
				fmt.Println("\nYou are not in a channel")
				return
			}
			cmd := "PART " + client.Channel + "\r\n"
			client.Write(cmd)
			client.Channel = ""
		case "/nick":
			if len(parts) > 1 {
				cmd := "NICK :" + parts[1] + "\r\n"
				client.Write(cmd)
			} else {
				fmt.Println("\nInvalid command, use /nick newnick")
			}
		case "/quit":
			if len(parts) > 1 {
				cmd := "QUIT :" + parts[1] + "\r\n"
				client.Write(cmd)
			} else {
				cmd := "QUIT :Bye bye\r\n"
				client.Write(cmd)
			}
		case "/help":
			fmt.Println("Commands:")
			fmt.Println("/join #channel - join a channel")
			fmt.Println("/part #channel- leave a channel")
			fmt.Println("/nick newnick - change your nickname")
			fmt.Println("/quit - quit the server")
		default:
			fmt.Println("\nInvalid command, use /help for more commands")
		}
		return
	}

	// check if the channel is set and send the message, otherwise print an error
	if client.Channel != "" {
		cmd := "PRIVMSG " + client.Channel + " :" + input + "\r\n"
		client.Write(cmd)
	} else {
		fmt.Println("Join a channel /join #channel, or /help for more commands")
	}
}
