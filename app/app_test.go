package app

import (
	"slices"
	"strings"
	"testing"

	"girc/config"
	"girc/irc"
)

type fakeSender struct{ sent []string }

func (f *fakeSender) Send(command string, params ...string) error {
	f.sent = append(f.sent, irc.Message{Command: command, Params: params}.String())
	return nil
}

type fakeView struct {
	lines []string
	users []string
}

func (f *fakeView) Println(line string)     { f.lines = append(f.lines, line) }
func (f *fakeView) SetUsers(nicks []string) { f.users = nicks }

func newTestApp(t *testing.T) (*App, *fakeSender, *fakeView) {
	t.Helper()
	s, v := &fakeSender{}, &fakeView{}
	return newApp(config.Config{Nick: "me"}, s, v), s, v
}

func feed(t *testing.T, a *App, lines ...string) {
	t.Helper()
	for _, line := range lines {
		m, err := irc.ParseMessage(line)
		if err != nil {
			t.Fatalf("ParseMessage(%q): %v", line, err)
		}
		a.handleMessage(m)
	}
}

func TestPingIsAnsweredWithPong(t *testing.T) {
	a, s, v := newTestApp(t)
	feed(t, a, "PING :irc.example.com")

	if want := []string{"PONG irc.example.com"}; !slices.Equal(s.sent, want) {
		t.Errorf("sent %q, want %q", s.sent, want)
	}
	if len(v.lines) != 0 {
		t.Errorf("PING should not be displayed, got %q", v.lines)
	}
}

func TestJoinAndNames(t *testing.T) {
	a, _, v := newTestApp(t)
	feed(t, a,
		":me!u@h JOIN #go",
		":srv 353 me = #go :@alice +bob me",
		":srv 353 me = #go :carol",
		":srv 366 me #go :End of /NAMES list.",
	)

	if got := a.state.Channel(); got != "#go" {
		t.Errorf("channel = %q, want #go", got)
	}
	if want := []string{"alice", "bob", "carol", "me"}; !slices.Equal(v.users, want) {
		t.Errorf("users = %q, want %q", v.users, want)
	}

	// Another user joining must not reset our channel or user list.
	feed(t, a, ":dave!u@h JOIN #go")
	if got := a.state.Channel(); got != "#go" {
		t.Errorf("channel after other JOIN = %q, want #go", got)
	}
	if want := []string{"alice", "bob", "carol", "dave", "me"}; !slices.Equal(v.users, want) {
		t.Errorf("users after other JOIN = %q, want %q", v.users, want)
	}

	// Another user leaving must not make us leave.
	feed(t, a, ":alice!u@h PART #go :bye", ":bob!u@h QUIT :gone")
	if got := a.state.Channel(); got != "#go" {
		t.Errorf("channel after other PART = %q, want #go", got)
	}
	if want := []string{"carol", "dave", "me"}; !slices.Equal(v.users, want) {
		t.Errorf("users after PART/QUIT = %q, want %q", v.users, want)
	}

	feed(t, a, ":me!u@h PART #go")
	if got := a.state.Channel(); got != "" {
		t.Errorf("channel after own PART = %q, want empty", got)
	}
	if len(v.users) != 0 {
		t.Errorf("users after own PART = %q, want empty", v.users)
	}
}

func TestNickChange(t *testing.T) {
	a, _, v := newTestApp(t)
	feed(t, a, ":me!u@h JOIN #go", ":srv 353 me = #go :me alice", ":srv 366 me #go :end")

	feed(t, a, ":alice!u@h NICK alicia")
	if got := a.state.Nick(); got != "me" {
		t.Errorf("own nick changed by other user's NICK: %q", got)
	}
	if want := []string{"alicia", "me"}; !slices.Equal(v.users, want) {
		t.Errorf("users = %q, want %q", v.users, want)
	}

	feed(t, a, ":me!u@h NICK :myself")
	if got := a.state.Nick(); got != "myself" {
		t.Errorf("own nick = %q, want myself", got)
	}
}

func TestNickInUseDuringRegistration(t *testing.T) {
	a, s, _ := newTestApp(t)
	feed(t, a, ":srv 433 * me :Nickname is already in use")

	if want := []string{"NICK me_"}; !slices.Equal(s.sent, want) {
		t.Errorf("sent %q, want %q", s.sent, want)
	}
	feed(t, a, ":srv 001 me_ :Welcome")
	if got := a.state.Nick(); got != "me_" {
		t.Errorf("nick = %q, want me_", got)
	}
}

func TestPrivmsgIsEscaped(t *testing.T) {
	a, _, v := newTestApp(t)
	feed(t, a, ":me!u@h JOIN #go", ":alice!u@h PRIVMSG #go :look [red]here[-]")

	last := v.lines[len(v.lines)-1]
	if !strings.Contains(last, "[red[]") {
		t.Errorf("message text was not escaped: %q", last)
	}
}

func TestMalformedMessagesDoNotPanic(t *testing.T) {
	a, _, _ := newTestApp(t)
	for _, cmd := range []string{"PING", "PRIVMSG", "NOTICE", "JOIN", "PART", "KICK", "QUIT", "NICK", "ERROR", "001", "353", "366", "433", "999"} {
		feed(t, a, cmd)
		feed(t, a, ":x!u@h "+cmd)
	}
}

func TestCommands(t *testing.T) {
	tests := []struct {
		input   string
		want    []string
		wantErr string
	}{
		{input: "hello", wantErr: "join a channel first"},
		{input: "/join #go", want: []string{"JOIN #go"}},
		{input: "/join   #go   key", want: []string{"JOIN #go key"}},
		{input: "/join", wantErr: "usage: /join"},
		{input: "/msg bob hi  there", want: []string{"PRIVMSG bob :hi  there"}},
		{input: "/msg bob", wantErr: "usage: /msg"},
		{input: "/nick new", want: []string{"NICK new"}},
		{input: "/nick", wantErr: "usage: /nick"},
		{input: "/quit", want: []string{"QUIT :Bye bye"}},
		{input: "/QUIT see you", want: []string{"QUIT :see you"}},
		{input: "/part", wantErr: "join a channel first"},
		{input: "/part #go later", want: []string{"PART #go later"}},
		{input: "/bogus", wantErr: "unknown command"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			a, s, _ := newTestApp(t)
			err := a.execute(tt.input)
			if tt.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("execute(%q) error = %v, want containing %q", tt.input, err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("execute(%q): %v", tt.input, err)
			}
			if !slices.Equal(s.sent, tt.want) {
				t.Errorf("execute(%q) sent %q, want %q", tt.input, s.sent, tt.want)
			}
		})
	}
}

func TestSayAndPartInChannel(t *testing.T) {
	a, s, v := newTestApp(t)
	feed(t, a, ":me!u@h JOIN #go")

	if err := a.execute("hello world"); err != nil {
		t.Fatal(err)
	}
	if err := a.execute("/part"); err != nil {
		t.Fatal(err)
	}
	if want := []string{"PRIVMSG #go :hello world", "PART #go"}; !slices.Equal(s.sent, want) {
		t.Errorf("sent %q, want %q", s.sent, want)
	}
	if last := v.lines[len(v.lines)-1]; !strings.Contains(last, "hello world") {
		t.Errorf("own message not echoed, last line %q", last)
	}
}
