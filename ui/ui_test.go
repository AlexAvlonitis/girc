package ui

import (
	"strings"
	"testing"
	"time"

	"github.com/gdamore/tcell/v2"
)

// These tests call the UI-goroutine methods directly; no terminal is needed.

func newTestUI(t *testing.T) (*UI, *[]string) {
	t.Helper()
	var submitted []string
	u := New(func(buffer, text string) { submitted = append(submitted, buffer+": "+text) })
	u.now = func() time.Time { return time.Date(2024, 1, 1, 12, 34, 0, 0, time.UTC) }
	return u, &submitted
}

func names(u *UI) string {
	var s []string
	for _, b := range u.buffers {
		s = append(s, b.name)
	}
	return strings.Join(s, ",")
}

func TestBuffers(t *testing.T) {
	u, _ := newTestUI(t)
	if got := names(u); got != StatusBuffer || u.active.name != StatusBuffer {
		t.Fatalf("initial buffers = %q, active %q", got, u.active.name)
	}

	u.switchTo(u.open("#go"))
	u.open("alice")
	u.open("#GO") // existing buffers are matched case-insensitively
	if got := names(u); got != "status,#go,alice" {
		t.Errorf("buffers = %q", got)
	}

	u.println("alice", "hi")
	u.println("alice", "there")
	u.println("#go", "seen")
	if alice := u.find("alice"); alice.unread != 2 {
		t.Errorf("alice unread = %d, want 2", alice.unread)
	}
	if u.active.unread != 0 {
		t.Errorf("active buffer unread = %d, want 0", u.active.unread)
	}
	if got := u.find("#go").view.GetText(true); got != "12:34 seen\n" {
		t.Errorf("#go text = %q", got)
	}

	u.println("#unknown", "lost?")
	if got := u.find(StatusBuffer).view.GetText(true); !strings.Contains(got, "lost?") {
		t.Errorf("line for unknown buffer not in status: %q", got)
	}

	u.switchTo(u.find("alice"))
	if alice := u.find("alice"); alice.unread != 0 {
		t.Errorf("unread not reset on switch: %d", alice.unread)
	}

	u.rename("alice", "alicia")
	if got := names(u); got != "status,#go,alicia" {
		t.Errorf("buffers after rename = %q", got)
	}

	// Closing the active buffer switches to its neighbor.
	u.close("alicia")
	if got := names(u); got != "status,#go" || u.active.name != "#go" {
		t.Errorf("after close: buffers %q, active %q", got, u.active.name)
	}

	u.close(StatusBuffer)
	if got := names(u); got != "status,#go" {
		t.Errorf("status buffer was closed: %q", got)
	}
}

func TestKeys(t *testing.T) {
	u, _ := newTestUI(t)
	u.open("#a")
	u.open("#b")

	u.handleKey(tcell.NewEventKey(tcell.KeyRune, '3', tcell.ModAlt))
	if u.active.name != "#b" {
		t.Errorf("Alt+3 -> %q, want #b", u.active.name)
	}
	u.handleKey(tcell.NewEventKey(tcell.KeyCtrlN, 0, tcell.ModCtrl))
	if u.active.name != StatusBuffer {
		t.Errorf("Ctrl+N wraps -> %q, want status", u.active.name)
	}
	u.handleKey(tcell.NewEventKey(tcell.KeyCtrlP, 0, tcell.ModCtrl))
	if u.active.name != "#b" {
		t.Errorf("Ctrl+P wraps -> %q, want #b", u.active.name)
	}
}

func TestSubmitUsesActiveBuffer(t *testing.T) {
	u, submitted := newTestUI(t)
	u.switchTo(u.open("#go"))

	u.input.SetText("hello")
	u.submit(tcell.KeyEnter)
	u.input.SetText("   ")
	u.submit(tcell.KeyEnter)

	if len(*submitted) != 1 || (*submitted)[0] != "#go: hello" {
		t.Errorf("submitted = %q", *submitted)
	}
	if u.input.GetText() != "   " {
		t.Error("blank input should be left alone")
	}
}

func TestUsers(t *testing.T) {
	u, _ := newTestUI(t)
	u.switchTo(u.open("#go"))
	u.find("#go").users = []User{{Nick: "alice", Prefix: "@"}, {Nick: "bob"}}
	u.renderUsers()

	if n := u.users.GetItemCount(); n != 2 {
		t.Fatalf("user list has %d items", n)
	}
	// The secondary text holds the bare nick used by /query.
	if _, nick := u.users.GetItemText(0); nick != "alice" {
		t.Errorf("secondary text = %q, want alice", nick)
	}
}
