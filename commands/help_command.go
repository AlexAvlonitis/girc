package commands

import "fmt"

type HelpCommand struct{}

func (c *HelpCommand) Execute() {
	fmt.Println("Commands:")
	fmt.Println("/join #channel - join a channel")
	fmt.Println("/part #channel- leave a channel")
	fmt.Println("/nick newnick - change your nickname")
	fmt.Println("/quit - quit the server")
}
