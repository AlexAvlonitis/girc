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

	// Initialize the ui
	ui := ui.NewUI(client)

	// Run the main application loop
	go func() {
		for {
			select {
			case <-done:
				return
			case msg := <-ch:
				formatted := client.Presenter.FormatMessage(msg)
				fmt.Fprintf(ui.TextView, "%s\n", formatted)
			}
		}
	}()

	// Run the application
	if err := ui.App.SetRoot(ui.Flex, true).SetFocus(ui.InputField).Run(); err != nil {
		log.Fatalf("Error running application: %v", err)
	}
}
