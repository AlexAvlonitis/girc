package app

import (
	"fmt"
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
	ch := make(chan []byte)
	defer close(ch)

	// Create a client and establish a connection
	client := connection.NewClient(ch, done)
	err := client.Connect()
	if err != nil {
		log.Fatalf("Error connecting to server: %s", err)
		return
	}

	// Initialize the ui
	ui := ui.NewUI(client)

	// Set what happens when the user presses Enter on the input
	ui.MessageInput.SetDoneFunc(func(key tcell.Key) {
		if key == tcell.KeyEnter {
			text := ui.MessageInput.GetText()
			if len(text) > 0 {
				commands.SendCommand(text, client)
				// Clear the input field after sending
				ui.MessageInput.SetText("") // Clear the input field
			}
		}
	})

	// Create a presenter to format incoming messages
	presenter := connection.NewPresenter(client)

	// Run the main application loop
	go func() {
		for {
			select {
			case <-done:
				return
			case msg := <-ch:
				message := presenter.FormatMessage(msg)

				switch message.Type {
				case "ping":
					client.SendPong(message.Content)
				// TODO: get the names of a channel
				case "join":
					fmt.Fprintf(ui.MessageView, "%s\n", message.Content)
					// names := presenter.NamesToList(message.Content)
					// ui.UsersView.Clear()
					// for _, name := range names {
					// 	ui.UsersView.AddItem(name, "", 0, nil)
					// }
					// ui.App.Draw()
				default:
					fmt.Fprintf(ui.MessageView, "%s\n", message.Content)
				}
			}
		}
	}()

	// Run the application
	if err := ui.App.SetRoot(ui.Grid, true).SetFocus(ui.MessageInput).Run(); err != nil {
		log.Fatalf("Error running application: %v", err)
	}
}
