package app

import (
	"fmt"
	"girc/connection"
	"girc/ui"
	"log"
)

func Init() {
	// Create done channel to signal when the application is done
	done := make(chan interface{})
	defer close(done)

	// Create a channel to read messages from the IRC server
	ch := make(chan []byte)

	// Create a client and establish a connection
	client := connection.NewClient(ch, done)
	err := client.Connect()
	if err != nil {
		log.Fatalf("Error connecting to server: %s", err)
		return
	}

	// Create a presenter to format incoming messages
	presenter := connection.NewPresenter(client)

	// Initialize the ui
	ui := ui.NewUI(client)

	// Run the main application loop
	go func() {
		for {
			select {
			case <-done:
				return
			case msg := <-ch:
				formatted := presenter.FormatMessage(msg)
				fmt.Fprintf(ui.MessageView, "%s\n", formatted)
			}
		}
	}()

	// Run the application
	if err := ui.App.SetRoot(ui.Flex, true).SetFocus(ui.MessageInput).Run(); err != nil {
		log.Fatalf("Error running application: %v", err)
	}
}
