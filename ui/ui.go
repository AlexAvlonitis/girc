package ui

import (
	"girc/interfaces"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type UI struct {
	App          *tview.Application
	MessageView  *tview.TextView
	UsersView    *tview.List
	Grid         *tview.Grid
	MessageInput *tview.InputField
	Client       interfaces.Client
}

// NewUI creates a new UI instance and initializes the components
func NewUI(c interfaces.Client) *UI {
	ui := &UI{Client: c}
	ui.App = tview.NewApplication()

	ui.initMessageView()
	ui.initUsersView()
	ui.setupUsersViewInputCapture()
	ui.initMessageInput()
	ui.initGrid()
	ui.setupInputCapture()

	return ui
}

// initMessageView initializes the TextView to display messages
func (ui *UI) initMessageView() {
	ui.MessageView = tview.NewTextView().
		SetDynamicColors(true).
		SetRegions(true).
		SetWordWrap(true).
		SetChangedFunc(func() {
			ui.App.Draw() // Redraw the application when the text changes
			ui.MessageView.ScrollToEnd()
		})

	ui.MessageView.SetBorder(true).SetTitle("GIRC Client")
}

// initUsersView initializes the ListView to display a list of users
func (ui *UI) initUsersView() {
	ui.UsersView = tview.NewList().
		ShowSecondaryText(false).
		SetMainTextColor(tcell.ColorGreen)

	ui.UsersView.SetBorder(true).SetTitle("Users")
}

// setupUsersViewInputCapture sets up keyboard event handling for the UsersView
// When Enter is pressed, the selected user's name is inserted into the MessageInput
func (ui *UI) setupUsersViewInputCapture() {
	ui.UsersView.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEnter {
			username, _ := ui.UsersView.GetItemText(ui.UsersView.GetCurrentItem())
			ui.MessageInput.SetText("/msg " + username + " ")
			ui.App.SetFocus(ui.MessageInput)
		}
		return event
	})
}

// initMessageInput initializes the InputField for user input
func (ui *UI) initMessageInput() {
	ui.MessageInput = tview.NewInputField().
		SetLabel("Input: ").
		SetFieldWidth(0).
		SetAcceptanceFunc(tview.InputFieldMaxLength(170))
}

// initGrid initializes the Grid layout to arrange the components
func (ui *UI) initGrid() {
	ui.Grid = tview.NewGrid().
		SetRows(0, 3).
		SetColumns(0, 30).
		AddItem(ui.MessageView, 0, 0, 1, 1, 0, 0, true).
		AddItem(ui.UsersView, 0, 1, 1, 1, 0, 0, false).
		AddItem(ui.MessageInput, 1, 0, 1, 2, 0, 0, false)
}

// setupInputCapture sets up keyboard event handling
func (ui *UI) setupInputCapture() {
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
}
