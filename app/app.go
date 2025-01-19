package app

import (
	"fmt"
	"girc/app/connection_manager"
	"girc/commands"
	"girc/connection"
	"girc/ui"
	"log"

	"github.com/gdamore/tcell/v2"
)

func Init() {
	// Create done channel to signal when the application is done
	done := make(chan interface{})
	defer close(done)

	// Create a channel to read messages from the IRC server
	ch := make(chan string)
	defer close(ch)

	// Create a client and establish a connection
	client := connection.NewClient(ch, done)
	err := client.Connect()
	if err != nil {
		log.Fatalf("Error connecting to server: %s", err)
		return
	}

	// Create a connection manager, to manage the connection details and channels
	cm := connection_manager.NewConnectionManager(client)

	// Initialize the ui
	ui := ui.NewUI()

	// Set what happens when the user presses Enter on the input
	ui.MessageInput.SetDoneFunc(func(key tcell.Key) {
		if key == tcell.KeyEnter {
			text := ui.MessageInput.GetText()
			if len(text) > 0 {
				commands.SendCommand(text, cm)
				// Clear the input field after sending
				ui.MessageInput.SetText("") // Clear the input field
			}
		}
	})

	// create a message parser, to parse incoming messages from the server
	msgParser := commands.NewMessageParser(cm, ui)

	// Run the main application loop
	go func() {
		for {
			select {
			case <-done:
				return
			case msg := <-ch:
				parsed := msgParser.Parse(msg)
				fmt.Fprintf(ui.MessageView, "%s\n", parsed)
			}
		}
	}()

	// Run the application
	if err := ui.App.SetRoot(ui.Grid, true).SetFocus(ui.MessageInput).Run(); err != nil {
		log.Fatalf("Error running application: %v", err)
	}
}
