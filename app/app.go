package app

import (
	"fmt"
	"girc/commands"
	"girc/connection"
	"log"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

func Init() {
	// Initialize the tview application
	app := tview.NewApplication()

	// Create the TextView to display messages
	textView := tview.NewTextView().
		SetDynamicColors(true).
		SetRegions(true).
		SetWordWrap(true).
		SetChangedFunc(func() {
			app.Draw() // Redraw the application when the text changes
		})

	// Set up a border and title for the text view
	textView.SetBorder(true).SetTitle("G-IRC Client")

	// Create channels for communication
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

	// Create the InputField for user input
	var inputField *tview.InputField
	inputField = tview.NewInputField().
		SetLabel("Input: ").
		SetFieldWidth(0).
		SetAcceptanceFunc(tview.InputFieldMaxLength(170)).
		SetDoneFunc(func(key tcell.Key) {
			if key == tcell.KeyEnter {
				text := inputField.GetText()
				if len(text) > 0 {
					commands.SendCommand(text, client)
					// Clear the input field after sending
					inputField.SetText("") // Clear the input field
				}
			}
		})

	// Create a flex layout to arrange the text view and input field
	flex := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(textView, 0, 1, true).
		AddItem(inputField, 3, 0, false)

	// Capture keyboard events and navigate between the widgets
	app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyTab: // Tab key
			// When Tab is pressed, move the focus to the InputField
			if app.GetFocus() == textView {
				app.SetFocus(inputField)
			} else {
				app.SetFocus(textView)
			}
		case tcell.KeyCtrlC: // Ctrl+C to quit the application
			// Stop the application on Ctrl+C
			app.Stop()
		}
		// Return the event to allow tview to handle other keys
		return event
	})

	// Set the focus to the inputField when the application starts
	app.SetFocus(inputField)

	textView.ScrollToEnd()

	// Run the main application loop
	go func() {
		for {
			select {
			case <-done:
				return
			case msg := <-ch:
				// When a message is received, update the text view
				fmt.Fprintf(textView, "%s\n", string(msg))
			}
		}
	}()

	// Run the application
	if err := app.SetRoot(flex, true).Run(); err != nil {
		log.Fatalf("Error running application: %v", err)
	}
}
