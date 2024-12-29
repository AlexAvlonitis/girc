package ui

import (
	"girc/connection"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type UI struct {
	App          *tview.Application
	MessageView  *tview.TextView
	UsersView    *tview.List
	Grid         *tview.Grid
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

	// Create the ListView to display a list of strings
	ui.UsersView = tview.NewList().
		ShowSecondaryText(false)

	// Set up a border and title for the list view
	ui.UsersView.SetBorder(true).SetTitle("Users")

	// Create the InputField for user input
	ui.MessageInput = tview.NewInputField().
		SetLabel("Input: ").
		SetFieldWidth(0).
		SetAcceptanceFunc(tview.InputFieldMaxLength(170))

	// Create a grid layout to arrange the text view, list view, and input field
	ui.Grid = tview.NewGrid().
		SetRows(0, 3).
		SetColumns(0, 30).
		AddItem(ui.MessageView, 0, 0, 1, 1, 0, 0, true).
		AddItem(ui.UsersView, 0, 1, 1, 1, 0, 0, false).
		AddItem(ui.MessageInput, 1, 0, 1, 2, 0, 0, false)

	// Capture keyboard events and navigate between the widgets
	ui.App.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyTab: // Tab key
			// When Tab is pressed, move the focus to the next widget
			if ui.App.GetFocus() == ui.MessageView {
				ui.App.SetFocus(ui.UsersView)
			} else if ui.App.GetFocus() == ui.UsersView {
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
