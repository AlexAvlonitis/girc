// Package ui implements the terminal user interface.
package ui

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

const maxInputLength = 400

// UI is the terminal interface: a message log, a user list and an input line.
// Println and SetUsers are safe to call from any goroutine.
type UI struct {
	app      *tview.Application
	messages *tview.TextView
	users    *tview.List
	input    *tview.InputField
	onSubmit func(string)
}

// New builds the interface. onSubmit is called on the UI goroutine with
// each line the user enters.
func New(title string, onSubmit func(string)) *UI {
	u := &UI{
		app:      tview.NewApplication(),
		onSubmit: onSubmit,
	}

	u.messages = tview.NewTextView().
		SetDynamicColors(true).
		SetWordWrap(true)
	u.messages.SetBorder(true).SetTitle(title)

	u.users = tview.NewList().
		ShowSecondaryText(false).
		SetMainTextColor(tcell.ColorGreen)
	u.users.SetBorder(true).SetTitle("Users")
	// Selecting a user starts a private message to them.
	u.users.SetSelectedFunc(func(_ int, nick, _ string, _ rune) {
		u.input.SetText("/msg " + nick + " ")
		u.app.SetFocus(u.input)
	})

	u.input = tview.NewInputField().
		SetLabel("Input: ").
		SetFieldWidth(0).
		SetAcceptanceFunc(tview.InputFieldMaxLength(maxInputLength))
	u.input.SetDoneFunc(func(key tcell.Key) {
		if key != tcell.KeyEnter {
			return
		}
		text := u.input.GetText()
		if text == "" {
			return
		}
		u.input.SetText("")
		u.onSubmit(text)
	})

	grid := tview.NewGrid().
		SetRows(0, 3).
		SetColumns(0, 30).
		AddItem(u.messages, 0, 0, 1, 1, 0, 0, false).
		AddItem(u.users, 0, 1, 1, 1, 0, 0, false).
		AddItem(u.input, 1, 0, 1, 2, 0, 0, true)

	u.app.SetRoot(grid, true).SetFocus(u.input)
	u.app.SetInputCapture(u.handleKey)

	return u
}

// handleKey implements global shortcuts. Tab cycles focus between widgets.
func (u *UI) handleKey(event *tcell.EventKey) *tcell.EventKey {
	if event.Key() != tcell.KeyTab {
		return event
	}
	switch u.app.GetFocus() {
	case u.input:
		u.app.SetFocus(u.messages)
	case u.messages:
		u.app.SetFocus(u.users)
	default:
		u.app.SetFocus(u.input)
	}
	return nil
}

// Run starts the event loop and blocks until Stop is called or the user
// presses Ctrl+C.
func (u *UI) Run() error {
	return u.app.Run()
}

// Stop terminates the event loop.
func (u *UI) Stop() {
	u.app.Stop()
}

// Println appends a line to the message log. The line may contain tview
// color tags; untrusted text must be passed through Escape first.
func (u *UI) Println(line string) {
	u.app.QueueUpdateDraw(func() {
		fmt.Fprintln(u.messages, line)
		u.messages.ScrollToEnd()
	})
}

// SetUsers replaces the contents of the user list.
func (u *UI) SetUsers(nicks []string) {
	u.app.QueueUpdateDraw(func() {
		u.users.Clear()
		for _, nick := range nicks {
			u.users.AddItem(nick, "", 0, nil)
		}
	})
}

// Escape escapes text so that it is displayed literally rather than being
// interpreted as color or region tags.
func Escape(text string) string {
	return tview.Escape(text)
}
