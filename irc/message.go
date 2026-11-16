// Package irc implements the parts of the IRC protocol (RFC 1459/2812)
// that the client needs: message parsing/formatting and the connection.
package irc

import (
	"errors"
	"strings"
)

// Numeric replies handled by the client.
const (
	RplWelcome       = "001"
	RplNamReply      = "353"
	RplEndOfNames    = "366"
	ErrNicknameInUse = "433"
)

// Message is a single IRC protocol message.
//
//	:nick!user@host PRIVMSG #channel :Hello there
//	 ^ Prefix       ^ Command  ^ Params[0], Params[1]
type Message struct {
	Prefix  string
	Command string
	Params  []string
}

// ErrEmptyMessage is returned by ParseMessage for lines without a command.
var ErrEmptyMessage = errors.New("irc: message has no command")

// ParseMessage parses a raw line received from the server.
// IRCv3 message tags are accepted but discarded.
func ParseMessage(line string) (Message, error) {
	line = strings.TrimRight(line, "\r\n")

	if strings.HasPrefix(line, "@") {
		_, line, _ = strings.Cut(line, " ")
	}
	line = strings.TrimLeft(line, " ")

	var m Message
	if strings.HasPrefix(line, ":") {
		m.Prefix, line, _ = strings.Cut(line[1:], " ")
	}

	for {
		line = strings.TrimLeft(line, " ")
		if line == "" {
			break
		}
		if m.Command != "" && line[0] == ':' {
			m.Params = append(m.Params, line[1:])
			break
		}

		var token string
		token, line, _ = strings.Cut(line, " ")
		if m.Command == "" {
			m.Command = strings.ToUpper(token)
		} else {
			m.Params = append(m.Params, token)
		}
	}

	if m.Command == "" {
		return Message{}, ErrEmptyMessage
	}
	return m, nil
}

// Nick returns the nickname part of the prefix ("nick" in "nick!user@host").
// For server-originated messages it returns the server name.
func (m Message) Nick() string {
	nick, _, _ := strings.Cut(m.Prefix, "!")
	nick, _, _ = strings.Cut(nick, "@")
	return nick
}

// Param returns the i-th parameter, or "" if it does not exist.
func (m Message) Param(i int) string {
	if i < 0 || i >= len(m.Params) {
		return ""
	}
	return m.Params[i]
}

// IsNumeric reports whether the message is a three-digit numeric reply.
func (m Message) IsNumeric() bool {
	if len(m.Command) != 3 {
		return false
	}
	for _, r := range m.Command {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}

// lineSanitizer strips characters that would terminate or corrupt a line,
// preventing protocol injection from user input.
var lineSanitizer = strings.NewReplacer("\r", "", "\n", "", "\x00", "")

// String formats the message in wire format, without the trailing CRLF.
// The last parameter is sent as a trailing parameter when required.
func (m Message) String() string {
	var b strings.Builder
	if m.Prefix != "" {
		b.WriteByte(':')
		b.WriteString(m.Prefix)
		b.WriteByte(' ')
	}
	b.WriteString(m.Command)
	for i, p := range m.Params {
		b.WriteByte(' ')
		last := i == len(m.Params)-1
		if last && (p == "" || p[0] == ':' || strings.Contains(p, " ")) {
			b.WriteByte(':')
		}
		b.WriteString(p)
	}
	return lineSanitizer.Replace(b.String())
}

// IsChannel reports whether name looks like a channel name.
func IsChannel(name string) bool {
	return name != "" && strings.ContainsRune("#&+!", rune(name[0]))
}
