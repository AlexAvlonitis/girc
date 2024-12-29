package ui

import (
	"girc/commands"
	"girc/connection"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type UI struct {
	App          *tview.Application
	MessageView  *tview.TextView
	Flex         *tview.Flex
	MessageInput *tview.InputField
	Client       *connection.Client
}

func NewUI(c *connection.Client) *UI {
	ui := &UI{Client: c}
	ui.App = tview.NewApplication()

	// Create the TextView to display messages
	ui.MessageView = tview.NewTextView().
		SetDynamicColors(true).
		SetRegions(true).
		SetWordWrap(true).
		SetChangedFunc(func() {
			ui.App.Draw() // Redraw the application when the text changes
		})

	// Set up a border and title for the text view
	ui.MessageView.SetBorder(true).SetTitle("GIRC Client")

	// Create the InputField for user input
	ui.MessageInput = tview.NewInputField().
		SetLabel("Input: ").
		SetFieldWidth(0).
		SetAcceptanceFunc(tview.InputFieldMaxLength(170)).
		SetDoneFunc(func(key tcell.Key) {
			if key == tcell.KeyEnter {
				text := ui.MessageInput.GetText()
				if len(text) > 0 {
					commands.SendCommand(text, ui.Client)
					// Clear the input field after sending
					ui.MessageInput.SetText("") // Clear the input field
				}
			}
		})

	// Create a flex layout to arrange the text view and input field
	ui.Flex = tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(ui.MessageView, 0, 1, true).
		AddItem(ui.MessageInput, 3, 0, false)

	// Capture keyboard events and navigate between the widgets
	ui.App.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyTab: // Tab key
			// When Tab is pressed, move the focus to the InputField
			if ui.App.GetFocus() == ui.MessageView {
				ui.App.SetFocus(ui.MessageInput)
			} else {
				ui.App.SetFocus(ui.MessageView)
			}
		case tcell.KeyCtrlC: // Ctrl+C to quit the application
			// Stop the application on Ctrl+C
			ui.App.Stop()
		}
		// Return the event to allow tview to handle other keys
		return event
	})

	ui.MessageView.ScrollToEnd()

	return ui
}
