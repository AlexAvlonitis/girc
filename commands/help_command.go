package commands

import "girc/interfaces"

type HelpCommand struct {
	Client interfaces.Client
}

func (h *HelpCommand) Execute() error {
	cmd, _ := h.Print()
	h.Client.PrintMessage(cmd)

	return nil
}

func (h *HelpCommand) Print() (string, error) {
	msg := "Commands:\n"
	msg += "/join #channel - join a channel\n"
	msg += "/part #channel - leave a channel\n"
	msg += "/nick newnick - change your nickname\n"
	msg += "/msg nickname message - send a private message\n"
	msg += "/quit - quit the server\n"

	return msg, nil
}
