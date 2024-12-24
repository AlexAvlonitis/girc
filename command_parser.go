package main

import (
	"strings"
)

// ParseCommand parses a command from a string and returs an appropriate IRC command
func ParseCommand(input string) string {
	// Split the string by the first space
	parts := strings.Split(input, " ")

	// Check if the split was successful
	if len(parts) == 2 {
		// Check if the first part is a valid IRC command
		switch parts[0] {
		case "/join":
			return "JOIN :#" + parts[1] + "\r\n"
		case "/nick":
			return "NICK :" + parts[1] + "\r\n"
		case "/msg":
			return "PRIVMSG #testing000 :" + parts[1] + "\r\n"
		}

	} else {
		return ""
	}

	return input
}
