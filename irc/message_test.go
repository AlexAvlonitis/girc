package irc

import (
	"reflect"
	"testing"
)

func TestParseMessage(t *testing.T) {
	tests := []struct {
		line string
		want Message
	}{
		{
			line: ":test!~test@test.com PRIVMSG #channel :Hi there!\r\n",
			want: Message{Prefix: "test!~test@test.com", Command: "PRIVMSG", Params: []string{"#channel", "Hi there!"}},
		},
		{
			line: "PING :irc.example.com",
			want: Message{Command: "PING", Params: []string{"irc.example.com"}},
		},
		{
			line: "PING",
			want: Message{Command: "PING"},
		},
		{
			line: ":server 353 me = #go :@alice +bob carol",
			want: Message{Prefix: "server", Command: "353", Params: []string{"me", "=", "#go", "@alice +bob carol"}},
		},
		{
			line: ":nick!u@h JOIN #go",
			want: Message{Prefix: "nick!u@h", Command: "JOIN", Params: []string{"#go"}},
		},
		{
			line: ":nick!u@h PRIVMSG #go :",
			want: Message{Prefix: "nick!u@h", Command: "PRIVMSG", Params: []string{"#go", ""}},
		},
		{
			line: ":nick!u@h PRIVMSG #go ::-) a :colon",
			want: Message{Prefix: "nick!u@h", Command: "PRIVMSG", Params: []string{"#go", ":-) a :colon"}},
		},
		{
			line: "@time=2024-01-01T00:00:00Z :n!u@h privmsg  #go   :x",
			want: Message{Prefix: "n!u@h", Command: "PRIVMSG", Params: []string{"#go", "x"}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.line, func(t *testing.T) {
			got, err := ParseMessage(tt.line)
			if err != nil {
				t.Fatalf("ParseMessage(%q) error: %v", tt.line, err)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseMessage(%q)\n got  %#v\n want %#v", tt.line, got, tt.want)
			}
		})
	}
}

func TestParseMessageEmpty(t *testing.T) {
	for _, line := range []string{"", "\r\n", ":prefix-only", "   "} {
		if _, err := ParseMessage(line); err == nil {
			t.Errorf("ParseMessage(%q): expected error", line)
		}
	}
}

func TestMessageNick(t *testing.T) {
	tests := map[string]string{
		"nick!user@host": "nick",
		"nick@host":      "nick",
		"irc.server.net": "irc.server.net",
		"":               "",
	}
	for prefix, want := range tests {
		if got := (Message{Prefix: prefix}).Nick(); got != want {
			t.Errorf("Nick() for prefix %q = %q, want %q", prefix, got, want)
		}
	}
}

func TestMessageParam(t *testing.T) {
	m := Message{Params: []string{"a"}}
	if got := m.Param(0); got != "a" {
		t.Errorf("Param(0) = %q, want a", got)
	}
	if got := m.Param(5); got != "" {
		t.Errorf("Param(5) = %q, want empty", got)
	}
}

func TestMessageString(t *testing.T) {
	tests := []struct {
		msg  Message
		want string
	}{
		{Message{Command: "JOIN", Params: []string{"#go"}}, "JOIN #go"},
		{Message{Command: "PRIVMSG", Params: []string{"#go", "hello world"}}, "PRIVMSG #go :hello world"},
		{Message{Command: "PRIVMSG", Params: []string{"#go", ":)"}}, "PRIVMSG #go ::)"},
		{Message{Command: "QUIT", Params: []string{""}}, "QUIT :"},
		{Message{Command: "QUIT"}, "QUIT"},
		// CR/LF must never reach the wire, or users could inject commands.
		{Message{Command: "PRIVMSG", Params: []string{"#go", "hi\r\nQUIT"}}, "PRIVMSG #go hiQUIT"},
	}
	for _, tt := range tests {
		if got := tt.msg.String(); got != tt.want {
			t.Errorf("%#v.String() = %q, want %q", tt.msg, got, tt.want)
		}
	}
}

func TestIsNumeric(t *testing.T) {
	for cmd, want := range map[string]bool{"001": true, "353": true, "PING": false, "12": false, "1a3": false} {
		if got := (Message{Command: cmd}).IsNumeric(); got != want {
			t.Errorf("IsNumeric(%q) = %v, want %v", cmd, got, want)
		}
	}
}
