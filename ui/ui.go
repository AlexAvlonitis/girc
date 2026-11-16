// Package ui implements the terminal user interface: a list of buffers
// (status, channels and private conversations), the message log of the
// active buffer, its user list and an input line.
package ui

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

const (
	// StatusBuffer is the buffer for server messages. It always exists.
	StatusBuffer = "status"
	// ActiveBuffer can be passed as a buffer name to target whichever
	// buffer is currently displayed.
	ActiveBuffer = ""

	maxInputLength = 400
	maxScrollback  = 5000
)

// User is an entry of a buffer's user list.
type User struct {
	Nick   string
	Prefix string // highest channel membership prefix, e.g. "@" or "+"
}

type buffer struct {
	name      string
	page      string // tview.Pages page name; stable across renames
	view      *tview.TextView
	users     []User
	topic     string
	unread    int
	highlight bool
}

// UI is the terminal interface.
//
// The exported methods other than Run may be called from any goroutine except
// the UI goroutine itself (for example from the onSubmit callback): they run
// their work on the UI goroutine and wait for it, unless the UI has stopped.
type UI struct {
	app      *tview.Application
	pages    *tview.Pages
	list     *tview.List // buffer list
	users    *tview.List
	input    *tview.InputField
	onSubmit func(buffer, text string)
	now      func() time.Time
	stopped  chan struct{} // closed when Run returns

	// Owned by the UI goroutine.
	buffers  []*buffer
	active   *buffer
	nextPage int
}

// New builds the interface. onSubmit is called on the UI goroutine with the
// name of the active buffer and each line the user enters; it must not block
// or call back into the UI.
func New(onSubmit func(buffer, text string)) *UI {
	u := &UI{
		app:      tview.NewApplication(),
		pages:    tview.NewPages(),
		onSubmit: onSubmit,
		now:      time.Now,
		stopped:  make(chan struct{}),
	}

	u.list = tview.NewList().
		ShowSecondaryText(false).
		SetHighlightFullLine(true)
	u.list.SetBorder(true).SetTitle("Buffers")
	u.list.SetSelectedFunc(func(index int, _, _ string, _ rune) {
		u.switchTo(u.buffers[index])
		u.app.SetFocus(u.input)
	})

	u.users = tview.NewList().
		ShowSecondaryText(false).
		SetMainTextColor(tcell.ColorGreen)
	u.users.SetBorder(true).SetTitle("Users")
	// Selecting a user starts a private conversation with them.
	u.users.SetSelectedFunc(func(_ int, _, nick string, _ rune) {
		u.input.SetText("/query " + nick)
		u.app.SetFocus(u.input)
	})

	u.input = tview.NewInputField().
		SetFieldWidth(0).
		SetAcceptanceFunc(tview.InputFieldMaxLength(maxInputLength))
	u.input.SetDoneFunc(u.submit)

	grid := tview.NewGrid().
		SetRows(0, 1).
		SetColumns(20, 0, 24).
		AddItem(u.list, 0, 0, 1, 1, 0, 0, false).
		AddItem(u.pages, 0, 1, 1, 1, 0, 0, false).
		AddItem(u.users, 0, 2, 1, 1, 0, 0, false).
		AddItem(u.input, 1, 0, 1, 3, 0, 0, true)

	u.switchTo(u.open(StatusBuffer))

	u.app.SetRoot(grid, true).SetFocus(u.input)
	u.app.SetInputCapture(u.handleKey)

	return u
}

// Run starts the event loop and blocks until Stop is called or the user
// presses Ctrl+C.
func (u *UI) Run() error {
	defer close(u.stopped)
	return u.app.Run()
}

// update runs f on the UI goroutine and redraws the screen. It waits until f
// has run, unless the UI has stopped. It must not be called from the UI
// goroutine itself.
func (u *UI) update(f func()) {
	done := make(chan struct{})
	go func() {
		u.app.QueueUpdateDraw(f)
		close(done)
	}()
	select {
	case <-done:
	case <-u.stopped:
	}
}

// Stop terminates the event loop.
func (u *UI) Stop() {
	u.app.Stop()
}

// Println appends a line to a buffer, prefixed with the current time. The line
// may contain tview style tags; untrusted text must be passed through Escape.
// Lines for buffers that do not exist go to the status buffer.
func (u *UI) Println(buffer, line string) {
	u.update(func() { u.println(buffer, line) })
}

// OpenBuffer creates a buffer if it does not exist yet, and switches to it
// if focus is true.
func (u *UI) OpenBuffer(name string, focus bool) {
	u.update(func() {
		b := u.open(name)
		if focus {
			u.switchTo(b)
		}
	})
}

// CloseBuffer removes a buffer. The status buffer cannot be closed.
func (u *UI) CloseBuffer(name string) {
	u.update(func() { u.close(name) })
}

// RenameBuffer renames a buffer, e.g. after a nick change.
func (u *UI) RenameBuffer(oldName, newName string) {
	u.update(func() { u.rename(oldName, newName) })
}

// ClearBuffer removes all lines from a buffer.
func (u *UI) ClearBuffer(name string) {
	u.update(func() {
		if b := u.find(name); b != nil {
			b.view.Clear()
		}
	})
}

// Highlight marks a buffer as having activity that concerns the user.
func (u *UI) Highlight(name string) {
	u.update(func() {
		if b := u.find(name); b != nil && b != u.active {
			b.highlight = true
			u.renderBufferList()
		}
	})
}

// SetUsers replaces a buffer's user list.
func (u *UI) SetUsers(name string, users []User) {
	u.update(func() {
		if b := u.find(name); b != nil {
			b.users = users
			if b == u.active {
				u.renderUsers()
			}
		}
	})
}

// SetTopic sets the topic shown in a buffer's title.
func (u *UI) SetTopic(name, topic string) {
	u.update(func() {
		if b := u.find(name); b != nil {
			b.topic = topic
			u.renderTitle(b)
		}
	})
}

// Escape escapes text so that it is displayed literally rather than being
// interpreted as style tags.
func Escape(text string) string {
	return tview.Escape(text)
}

// The methods below run on the UI goroutine.

func (u *UI) submit(key tcell.Key) {
	if key != tcell.KeyEnter {
		return
	}
	text := u.input.GetText()
	if strings.TrimSpace(text) == "" {
		return
	}
	u.input.SetText("")
	u.active.view.ScrollToEnd()
	u.onSubmit(u.active.name, text)
}

// handleKey implements the global shortcuts.
func (u *UI) handleKey(event *tcell.EventKey) *tcell.EventKey {
	switch event.Key() {
	case tcell.KeyTab:
		u.cycleFocus()
		return nil
	case tcell.KeyCtrlN:
		u.switchBy(1)
		return nil
	case tcell.KeyCtrlP:
		u.switchBy(-1)
		return nil
	case tcell.KeyPgUp, tcell.KeyPgDn:
		if u.input.HasFocus() {
			// Scroll the message log without leaving the input line.
			u.active.view.InputHandler()(event, func(tview.Primitive) {})
			return nil
		}
	case tcell.KeyRune:
		// Alt+1..9 and Alt+0 jump to buffers 1-10.
		if event.Modifiers()&tcell.ModAlt != 0 && event.Rune() >= '0' && event.Rune() <= '9' {
			i := int(event.Rune() - '1')
			if event.Rune() == '0' {
				i = 9
			}
			if i < len(u.buffers) {
				u.switchTo(u.buffers[i])
			}
			return nil
		}
	}
	return event
}

func (u *UI) cycleFocus() {
	switch {
	case u.input.HasFocus():
		u.app.SetFocus(u.pages)
	case u.pages.HasFocus():
		u.app.SetFocus(u.users)
	case u.users.HasFocus():
		u.app.SetFocus(u.list)
	default:
		u.app.SetFocus(u.input)
	}
}

// find returns the buffer with the given name, the active buffer for
// ActiveBuffer, or nil.
func (u *UI) find(name string) *buffer {
	if name == ActiveBuffer {
		return u.active
	}
	for _, b := range u.buffers {
		if strings.EqualFold(b.name, name) {
			return b
		}
	}
	return nil
}

func (u *UI) index(b *buffer) int {
	for i, other := range u.buffers {
		if other == b {
			return i
		}
	}
	return -1
}

func (u *UI) open(name string) *buffer {
	if b := u.find(name); b != nil {
		return b
	}

	b := &buffer{
		name: name,
		page: strconv.Itoa(u.nextPage),
		view: tview.NewTextView().
			SetDynamicColors(true).
			SetWordWrap(true).
			SetMaxLines(maxScrollback),
	}
	u.nextPage++
	b.view.SetBorder(true)
	u.renderTitle(b)

	u.buffers = append(u.buffers, b)
	u.pages.AddPage(b.page, b.view, true, false)
	u.renderBufferList()
	return b
}

func (u *UI) close(name string) {
	b := u.find(name)
	if b == nil || b.name == StatusBuffer {
		return
	}
	i := u.index(b)
	u.buffers = append(u.buffers[:i], u.buffers[i+1:]...)
	u.pages.RemovePage(b.page)
	if b == u.active {
		u.switchTo(u.buffers[max(i-1, 0)])
	} else {
		u.renderBufferList()
	}
}

func (u *UI) rename(oldName, newName string) {
	b := u.find(oldName)
	if b == nil || b.name == StatusBuffer || (u.find(newName) != nil && !strings.EqualFold(oldName, newName)) {
		return
	}
	b.name = newName
	u.renderTitle(b)
	u.renderBufferList()
	if b == u.active {
		u.renderInputLabel()
	}
}

func (u *UI) println(name, line string) {
	b := u.find(name)
	if b == nil {
		b = u.find(StatusBuffer)
	}
	fmt.Fprintf(b.view, "[gray]%s[-] %s\n", u.now().Format("15:04"), line)
	if b != u.active {
		b.unread++
		u.renderBufferList()
	}
}

func (u *UI) switchBy(delta int) {
	n := len(u.buffers)
	u.switchTo(u.buffers[(u.index(u.active)+delta+n)%n])
}

func (u *UI) switchTo(b *buffer) {
	u.active = b
	b.unread, b.highlight = 0, false
	u.pages.SwitchToPage(b.page)
	u.renderBufferList()
	u.renderUsers()
	u.renderInputLabel()
}

func (u *UI) renderTitle(b *buffer) {
	title := " " + Escape(b.name) + " "
	if b.topic != "" {
		title = " " + Escape(b.name) + " - " + Escape(b.topic) + " "
	}
	b.view.SetTitle(title)
}

func (u *UI) renderInputLabel() {
	u.input.SetLabel("[" + Escape(u.active.name) + "] ")
}

func (u *UI) renderBufferList() {
	u.list.Clear()
	for i, b := range u.buffers {
		name := fmt.Sprintf("%d %s", i+1, Escape(b.name))
		switch {
		case b == u.active:
		case b.highlight:
			name = fmt.Sprintf("[yellow::b]%s (%d)[-::-]", name, b.unread)
		case b.unread > 0:
			name = fmt.Sprintf("[white::b]%s (%d)[-::-]", name, b.unread)
		default:
			name = "[gray]" + name + "[-]"
		}
		u.list.AddItem(name, "", 0, nil)
	}
	if i := u.index(u.active); i >= 0 {
		u.list.SetCurrentItem(i)
	}
}

func (u *UI) renderUsers() {
	u.users.Clear()
	for _, user := range u.active.users {
		label := Escape(user.Nick)
		if user.Prefix != "" {
			label = "[yellow]" + Escape(user.Prefix) + "[-]" + label
		}
		u.users.AddItem(label, user.Nick, 0, nil)
	}
	u.users.SetTitle(fmt.Sprintf("Users (%d)", len(u.active.users)))
}
