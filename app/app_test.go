package app

import (
	"context"
	"io"
	"slices"
	"strings"
	"sync"
	"testing"
	"time"

	"girc/config"
	"girc/irc"
	"girc/ui"
)

type fakeSender struct {
	mu   sync.Mutex
	sent []string
}

func (f *fakeSender) Send(command string, params ...string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.sent = append(f.sent, irc.Message{Command: command, Params: params}.String())
	return nil
}

func (f *fakeSender) snapshot() []string {
	f.mu.Lock()
	defer f.mu.Unlock()
	return slices.Clone(f.sent)
}

// eventually fails the test if cond does not become true within a few seconds.
func eventually(t *testing.T, what string, cond func() bool) {
	t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	for !cond() {
		if time.Now().After(deadline) {
			t.Fatalf("timed out waiting for %s", what)
		}
		time.Sleep(time.Millisecond)
	}
}

// fakeView records what the app displays. It is safe for concurrent use so it
// can also be used with the event loop.
type fakeView struct {
	mu          sync.Mutex
	lines       map[string][]string // buffer -> lines
	open        []string
	focused     string
	users       map[string][]ui.User
	topics      map[string]string
	highlighted map[string]bool
}

func newFakeView() *fakeView {
	return &fakeView{
		lines:       make(map[string][]string),
		users:       make(map[string][]ui.User),
		topics:      make(map[string]string),
		highlighted: make(map[string]bool),
		open:        []string{ui.StatusBuffer},
	}
}

func (f *fakeView) Println(buffer, line string) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.lines[buffer] = append(f.lines[buffer], line)
}

func (f *fakeView) OpenBuffer(name string, focus bool) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if !slices.Contains(f.open, name) {
		f.open = append(f.open, name)
	}
	if focus {
		f.focused = name
	}
}

func (f *fakeView) CloseBuffer(name string) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.open = slices.DeleteFunc(f.open, func(s string) bool { return s == name })
}

func (f *fakeView) RenameBuffer(oldName, newName string) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if i := slices.Index(f.open, oldName); i >= 0 {
		f.open[i] = newName
	}
}

func (f *fakeView) ClearBuffer(name string) {
	f.mu.Lock()
	defer f.mu.Unlock()
	delete(f.lines, name)
}

func (f *fakeView) Highlight(name string) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.highlighted[name] = true
}

func (f *fakeView) SetUsers(name string, users []ui.User) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.users[name] = users
}

func (f *fakeView) SetTopic(name, topic string) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.topics[name] = topic
}

// last returns the last line printed to buffer.
func (f *fakeView) last(buffer string) string {
	f.mu.Lock()
	defer f.mu.Unlock()
	lines := f.lines[buffer]
	if len(lines) == 0 {
		return ""
	}
	return lines[len(lines)-1]
}

// userList renders a buffer's users as "@alice +bob carol".
func (f *fakeView) userList(buffer string) string {
	f.mu.Lock()
	defer f.mu.Unlock()
	var names []string
	for _, u := range f.users[buffer] {
		names = append(names, u.Prefix+u.Nick)
	}
	return strings.Join(names, " ")
}

func newTestApp(t *testing.T) (*App, *fakeSender, *fakeView) {
	t.Helper()
	s, v := &fakeSender{}, newFakeView()
	return newApp(config.Config{Nick: "me", User: "me", RealName: "Real Me"}, s, v), s, v
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

func run(t *testing.T, a *App, buffer, text string) error {
	t.Helper()
	return a.execute(input{buffer: buffer, text: text})
}

// joined returns an app that has joined #go with alice (op), bob (voice) and carol.
func joined(t *testing.T) (*App, *fakeSender, *fakeView) {
	t.Helper()
	a, s, v := newTestApp(t)
	feed(t, a,
		":srv 001 me :Welcome",
		":me!u@h JOIN #go",
		":srv 353 me = #go :@alice +bob",
		":srv 353 me = #go :carol me",
		":srv 366 me #go :End of /NAMES list.",
	)
	s.sent = nil
	return a, s, v
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

func TestJoinOpensBufferWithUsers(t *testing.T) {
	_, _, v := joined(t)

	if v.focused != "#go" {
		t.Errorf("focused buffer = %q, want #go", v.focused)
	}
	if got, want := v.userList("#go"), "@alice +bob carol me"; got != want {
		t.Errorf("users = %q, want %q", got, want)
	}
}

func TestMultipleChannels(t *testing.T) {
	a, _, v := joined(t)
	feed(t, a,
		":me!u@h JOIN #rust",
		":srv 353 me = #rust :me dave",
		":srv 366 me #rust :End",
		":alice!u@h PRIVMSG #go :hello go",
		":dave!u@h PRIVMSG #rust :hello rust",
	)

	if !strings.Contains(v.last("#go"), "hello go") {
		t.Errorf("#go last line = %q", v.last("#go"))
	}
	if !strings.Contains(v.last("#rust"), "hello rust") {
		t.Errorf("#rust last line = %q", v.last("#rust"))
	}
	if got := v.userList("#go"); got != "@alice +bob carol me" {
		t.Errorf("#go users changed by joining #rust: %q", got)
	}

	// Other users' joins and parts only affect their channel.
	feed(t, a, ":eve!u@h JOIN #rust", ":alice!u@h PART #go :bye")
	if got := v.userList("#rust"); got != "dave eve me" {
		t.Errorf("#rust users = %q", got)
	}
	if got := v.userList("#go"); got != "+bob carol me" {
		t.Errorf("#go users = %q", got)
	}

	// A quit is shown in, and removes the user from, every shared channel.
	feed(t, a, ":me!u@h JOIN #c", ":srv 353 me = #c :me carol", ":srv 366 me #c :End", ":carol!u@h QUIT :gone")
	for _, ch := range []string{"#go", "#c"} {
		if strings.Contains(v.userList(ch), "carol") || !strings.Contains(v.last(ch), "carol has quit") {
			t.Errorf("%s: users %q, last line %q", ch, v.userList(ch), v.last(ch))
		}
	}
	if strings.Contains(v.last("#rust"), "carol") {
		t.Errorf("quit shown in #rust where carol was not present")
	}

	// Leaving a channel closes its buffer and keeps the others.
	feed(t, a, ":me!u@h PART #go")
	if slices.Contains(v.open, "#go") || !slices.Contains(v.open, "#rust") {
		t.Errorf("open buffers = %q", v.open)
	}
	if a.s.channel("#go") != nil {
		t.Error("#go still joined after PART")
	}
}

func TestPrivateMessagesOpenQueryBuffer(t *testing.T) {
	a, s, v := joined(t)
	feed(t, a, ":alice!u@h PRIVMSG me :psst")

	if !slices.Contains(v.open, "alice") {
		t.Fatalf("no query buffer opened, open = %q", v.open)
	}
	if v.focused == "alice" {
		t.Error("incoming PM should not steal focus")
	}
	if !strings.Contains(v.last("alice"), "psst") || !v.highlighted["alice"] {
		t.Errorf("PM not shown/highlighted: %q", v.last("alice"))
	}

	// Typing in the query buffer replies to alice.
	if err := run(t, a, "alice", "hi back"); err != nil {
		t.Fatal(err)
	}
	if want := []string{"PRIVMSG alice :hi back"}; !slices.Equal(s.sent, want) {
		t.Errorf("sent %q, want %q", s.sent, want)
	}

	// The query follows nick changes.
	feed(t, a, ":alice!u@h NICK alicia")
	if !slices.Contains(v.open, "alicia") || slices.Contains(v.open, "alice") {
		t.Errorf("query not renamed, open = %q", v.open)
	}
	if got := v.userList("#go"); got != "@alicia +bob carol me" {
		t.Errorf("users after nick change = %q", got)
	}

	if err := run(t, a, "alicia", "/close"); err != nil {
		t.Fatal(err)
	}
	if slices.Contains(v.open, "alicia") {
		t.Errorf("query not closed, open = %q", v.open)
	}
}

func TestQueryCommand(t *testing.T) {
	a, s, v := joined(t)
	if err := run(t, a, "#go", "/query bob hey there"); err != nil {
		t.Fatal(err)
	}
	if v.focused != "bob" {
		t.Errorf("focused = %q, want bob", v.focused)
	}
	if want := []string{"PRIVMSG bob :hey there"}; !slices.Equal(s.sent, want) {
		t.Errorf("sent %q, want %q", s.sent, want)
	}
	if !strings.Contains(v.last("bob"), "hey there") {
		t.Errorf("own message not echoed in query: %q", v.last("bob"))
	}
}

func TestMentionHighlights(t *testing.T) {
	a, _, v := joined(t)
	feed(t, a, ":alice!u@h PRIVMSG #go :memes are fun")
	if v.highlighted["#go"] {
		t.Error("'memes' should not count as a mention of 'me'")
	}
	feed(t, a, ":alice!u@h PRIVMSG #go :hey ME, look")
	if !v.highlighted["#go"] {
		t.Error("mention not highlighted")
	}
}

func TestModeChangesUpdateUserList(t *testing.T) {
	a, _, v := joined(t)
	feed(t, a, ":alice!u@h MODE #go +o-v+l carol bob 10")
	if got, want := v.userList("#go"), "@alice @carol bob me"; got != want {
		t.Errorf("users = %q, want %q", got, want)
	}

	feed(t, a, ":srv 005 me PREFIX=(qov)~@+ :are supported")
	feed(t, a, ":alice!u@h MODE #go +q me")
	if got, want := v.userList("#go"), "~me @alice @carol bob"; got != want {
		t.Errorf("users = %q, want %q", got, want)
	}
}

func TestTopic(t *testing.T) {
	a, _, v := joined(t)
	feed(t, a, ":srv 332 me #go :Go talk", ":srv 333 me #go alice 1700000000")
	if v.topics["#go"] != "Go talk" {
		t.Errorf("topic = %q", v.topics["#go"])
	}
	feed(t, a, ":bob!u@h TOPIC #go :New [topic]")
	if v.topics["#go"] != "New [topic]" {
		t.Errorf("topic = %q", v.topics["#go"])
	}
}

func TestKickKeepsBufferOpen(t *testing.T) {
	a, s, v := joined(t)
	feed(t, a, ":alice!u@h KICK #go me :behave")

	if !slices.Contains(v.open, "#go") {
		t.Error("buffer closed after kick")
	}
	if !strings.Contains(v.last("#go"), "kicked") {
		t.Errorf("last line = %q", v.last("#go"))
	}
	if err := run(t, a, "#go", "hello?"); err == nil || !strings.Contains(err.Error(), "not in #go") {
		t.Errorf("say after kick: err = %v", err)
	}
	if len(s.sent) != 0 {
		t.Errorf("sent %q after kick", s.sent)
	}
}

func TestCTCP(t *testing.T) {
	a, s, v := joined(t)
	feed(t, a, ":alice!u@h PRIVMSG me :\x01VERSION\x01")
	if len(s.sent) != 1 || !strings.HasPrefix(s.sent[0], "NOTICE alice :\x01VERSION girc") {
		t.Errorf("VERSION reply = %q", s.sent)
	}
	if slices.Contains(v.open, "alice") {
		t.Error("CTCP request opened a query buffer")
	}

	feed(t, a, ":alice!u@h PRIVMSG #go :\x01ACTION waves\x01")
	if !strings.Contains(v.last("#go"), "* alice") {
		t.Errorf("action shown as %q", v.last("#go"))
	}
}

func TestNumericRouting(t *testing.T) {
	a, _, v := joined(t)
	feed(t, a,
		":srv 324 me #go +nt",
		":srv 311 me bob ~b host * :Bob",
		":srv 401 me nobody :No such nick",
		":srv 372 me :- message of the day",
	)
	if !strings.Contains(v.last("#go"), "+nt") {
		t.Errorf("channel numeric not in channel buffer: %q", v.last("#go"))
	}
	if lines := v.lines[ui.ActiveBuffer]; len(lines) != 2 || !strings.Contains(lines[1], "[red]") {
		t.Errorf("replies to user = %q", lines)
	}
	if !strings.Contains(v.last(ui.StatusBuffer), "message of the day") {
		t.Errorf("status = %q", v.last(ui.StatusBuffer))
	}
}

func TestNickInUseDuringRegistration(t *testing.T) {
	a, s, _ := newTestApp(t)
	feed(t, a, ":srv 433 * me :Nickname is already in use")

	if want := []string{"NICK me_"}; !slices.Equal(s.sent, want) {
		t.Errorf("sent %q, want %q", s.sent, want)
	}
	feed(t, a, ":srv 001 me_ :Welcome")
	if a.s.nick != "me_" {
		t.Errorf("nick = %q, want me_", a.s.nick)
	}
}

func TestMessagesAreEscaped(t *testing.T) {
	a, _, v := joined(t)
	feed(t, a, ":alice!u@h PRIVMSG #go :look [red]here[-]")
	if !strings.Contains(v.last("#go"), "[red[]") {
		t.Errorf("message text was not escaped: %q", v.last("#go"))
	}
}

func TestMalformedMessagesDoNotPanic(t *testing.T) {
	a, _, _ := joined(t)
	cmds := []string{"PING", "PRIVMSG", "NOTICE", "JOIN", "PART", "KICK", "QUIT", "NICK", "TOPIC", "MODE",
		"INVITE", "ERROR", "001", "005", "331", "332", "333", "353", "366", "433", "999"}
	for _, cmd := range cmds {
		feed(t, a, cmd, ":x!u@h "+cmd, ":x!u@h "+cmd+" #go", ":x!u@h "+cmd+" me")
	}
}

func TestCommands(t *testing.T) {
	tests := []struct {
		buffer  string
		input   string
		want    []string
		wantErr string
	}{
		{buffer: ui.StatusBuffer, input: "hello", wantErr: "no channel or conversation"},
		{input: "hello", want: []string{"PRIVMSG #go hello"}},
		{input: "//not a command", want: []string{"PRIVMSG #go :/not a command"}},
		{input: "/join #a,#b", want: []string{"JOIN #a,#b"}},
		{input: "/join   #go   key", want: []string{"JOIN #go key"}},
		{input: "/join", wantErr: "usage: /join"},
		{input: "/part", want: []string{"PART #go"}},
		{input: "/part see you", want: []string{"PART #go :see you"}},
		{input: "/part #other later", want: []string{"PART #other later"}},
		{buffer: ui.StatusBuffer, input: "/part", wantErr: "not a channel"},
		{input: "/close", want: []string{"PART #go"}},
		{buffer: ui.StatusBuffer, input: "/close", wantErr: "cannot be closed"},
		{input: "/msg bob hi  there", want: []string{"PRIVMSG bob :hi  there"}},
		{input: "/msg bob", wantErr: "usage: /msg"},
		{input: "/me waves", want: []string{"PRIVMSG #go :\x01ACTION waves\x01"}},
		{input: "/notice bob hi", want: []string{"NOTICE bob hi"}},
		{input: "/ctcp bob version", want: []string{"PRIVMSG bob \x01VERSION\x01"}},
		{input: "/nick new", want: []string{"NICK new"}},
		{input: "/nick", wantErr: "usage: /nick"},
		{input: "/topic", want: []string{"TOPIC #go"}},
		{input: "/topic Go is fun", want: []string{"TOPIC #go :Go is fun"}},
		{input: "/mode", want: []string{"MODE #go"}},
		{input: "/mode +m", want: []string{"MODE #go +m"}},
		{input: "/mode me +i", want: []string{"MODE me +i"}},
		{input: "/op a b c d", want: []string{"MODE #go +ooo a b c", "MODE #go +o d"}},
		{input: "/devoice bob", want: []string{"MODE #go -v bob"}},
		{input: "/ban spammer", want: []string{"MODE #go +b spammer!*@*"}},
		{input: "/kick bob be nice", want: []string{"KICK #go bob :be nice"}},
		{buffer: ui.StatusBuffer, input: "/kick bob", wantErr: "not a channel"},
		{input: "/invite dave", want: []string{"INVITE dave #go"}},
		{input: "/whois bob", want: []string{"WHOIS bob"}},
		{input: "/who", want: []string{"WHO #go"}},
		{input: "/away", want: []string{"AWAY"}},
		{input: "/away lunch time", want: []string{"AWAY :lunch time"}},
		{input: "/list", want: []string{"LIST"}},
		{input: "/names", want: []string{"NAMES #go"}},
		{input: "/quote PRIVMSG #x :raw text", want: []string{"PRIVMSG #x :raw text"}},
		{input: "/quit", want: []string{"QUIT :Bye bye"}},
		{input: "/QUIT see you", want: []string{"QUIT :see you"}},
		{input: "/bogus", wantErr: "unknown command"},
	}

	for _, tt := range tests {
		buffer := tt.buffer
		if buffer == "" {
			buffer = "#go"
		}
		t.Run(buffer+" "+tt.input, func(t *testing.T) {
			a, s, _ := joined(t)
			err := run(t, a, buffer, tt.input)
			if tt.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("error = %v, want containing %q", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if !slices.Equal(s.sent, tt.want) {
				t.Errorf("sent %q, want %q", s.sent, tt.want)
			}
		})
	}
}

func TestHelpListsEveryCommand(t *testing.T) {
	a, _, v := newTestApp(t)
	if err := run(t, a, ui.StatusBuffer, "/help"); err != nil {
		t.Fatal(err)
	}
	help := v.last(ui.ActiveBuffer)
	for name := range a.commands {
		if !strings.Contains(help, "/"+name) {
			t.Errorf("/help does not mention /%s", name)
		}
	}
}

func TestEventLoop(t *testing.T) {
	a, s, v := newTestApp(t)
	messages := make(chan irc.Message)
	readErr := make(chan error, 1)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go a.loop(ctx, messages, readErr)

	messages <- irc.Message{Prefix: "me!u@h", Command: "JOIN", Params: []string{"#go"}}
	a.submit("#go", "hello")
	a.submit("#go", "/quit")
	eventually(t, "QUIT to be sent", func() bool { return slices.Contains(s.snapshot(), "QUIT :Bye bye") })

	// The server closes the connection after QUIT; the loop must then exit.
	readErr <- io.EOF
	select {
	case <-a.done:
	case <-time.After(5 * time.Second):
		t.Fatal("event loop did not exit after quitting")
	}

	want := []string{"NICK me", "USER me 0 * :Real Me", "PRIVMSG #go hello", "QUIT :Bye bye"}
	if got := s.snapshot(); !slices.Equal(got, want) {
		t.Errorf("sent %q, want %q", got, want)
	}
	if !strings.Contains(v.last("#go"), "hello") {
		t.Errorf("#go last line = %q", v.last("#go"))
	}
}

func TestQuitAfterDisconnect(t *testing.T) {
	a, s, v := newTestApp(t)
	messages := make(chan irc.Message)
	readErr := make(chan error, 1)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go a.loop(ctx, messages, readErr)
	readErr <- io.EOF
	eventually(t, "disconnect to be reported", func() bool {
		return strings.Contains(v.last(ui.StatusBuffer), "Ctrl+C")
	})

	a.submit(ui.StatusBuffer, "/join #go")
	eventually(t, "not connected error", func() bool {
		return strings.Contains(v.last(ui.StatusBuffer), "not connected")
	})

	a.submit(ui.StatusBuffer, "/quit")
	select {
	case <-a.done:
	case <-time.After(5 * time.Second):
		t.Fatal("event loop did not exit on /quit after disconnect")
	}
	if got := s.snapshot(); len(got) != 2 { // only NICK and USER
		t.Errorf("sent %q while disconnected", got)
	}
}
