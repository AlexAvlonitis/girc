package app

import (
	"cmp"
	"slices"
	"strconv"
	"strings"

	"girc/irc"
	"girc/ui"
)

// member is a user in a channel.
type member struct {
	nick  string
	modes string // prefix modes held, e.g. "ov"
}

// channel is a joined channel.
type channel struct {
	name    string
	topic   string
	members map[string]*member // keyed by irc.Fold(nick)
	// pending collects RPL_NAMREPLY entries until RPL_ENDOFNAMES.
	pending map[string]*member
}

func (c *channel) member(nick string) *member {
	return c.members[irc.Fold(nick)]
}

// session is the client's view of the connection. It is owned by the event
// loop goroutine and must not be accessed from anywhere else.
type session struct {
	nick       string
	registered bool
	channels   map[string]*channel // joined channels, keyed by irc.Fold(name)
	queries    map[string]string   // open private conversations: folded nick -> nick

	// Server capabilities from RPL_ISUPPORT, with RFC 2812 defaults.
	prefixModes   string    // channel membership modes, highest first
	prefixSymbols string    // matching nick prefixes in RPL_NAMREPLY
	chanModes     [4]string // CHANMODES types A, B, C and D
	maxModes      int       // MODES: max parameterized modes per MODE command
}

func newSession(nick string) *session {
	return &session{
		nick:          nick,
		channels:      make(map[string]*channel),
		queries:       make(map[string]string),
		prefixModes:   "ov",
		prefixSymbols: "@+",
		chanModes:     [4]string{"beI", "k", "l", "imnpst"},
		maxModes:      3,
	}
}

func (s *session) isMe(nick string) bool {
	return irc.EqualFold(s.nick, nick)
}

// channel returns the joined channel with the given name, or nil.
func (s *session) channel(name string) *channel {
	return s.channels[irc.Fold(name)]
}

func (s *session) join(name string) *channel {
	c := &channel{name: name, members: make(map[string]*member)}
	s.channels[irc.Fold(name)] = c
	return c
}

func (s *session) part(name string) {
	delete(s.channels, irc.Fold(name))
}

// channelsWith returns the joined channels nick is a member of.
func (s *session) channelsWith(nick string) []*channel {
	var chans []*channel
	for _, c := range s.channels {
		if c.member(nick) != nil {
			chans = append(chans, c)
		}
	}
	return chans
}

// query returns the name of the open conversation with nick, if any.
func (s *session) query(nick string) (string, bool) {
	name, ok := s.queries[irc.Fold(nick)]
	return name, ok
}

// openQuery records an open conversation with nick and returns its name.
func (s *session) openQuery(nick string) string {
	if name, ok := s.query(nick); ok {
		return name
	}
	s.queries[irc.Fold(nick)] = nick
	return nick
}

func (s *session) closeQuery(nick string) {
	delete(s.queries, irc.Fold(nick))
}

// renameQuery renames an open conversation and reports whether one existed.
func (s *session) renameQuery(oldNick, newNick string) (string, bool) {
	name, ok := s.query(oldNick)
	if !ok {
		return "", false
	}
	delete(s.queries, irc.Fold(oldNick))
	s.queries[irc.Fold(newNick)] = newNick
	return name, true
}

// parseMember parses a RPL_NAMREPLY entry such as "@+nick".
func (s *session) parseMember(entry string) *member {
	m := &member{}
	for entry != "" {
		i := strings.IndexByte(s.prefixSymbols, entry[0])
		if i < 0 {
			break
		}
		m.modes += string(s.prefixModes[i])
		entry = entry[1:]
	}
	m.nick, _, _ = strings.Cut(entry, "!") // userhost-in-names
	return m
}

// addNames buffers a RPL_NAMREPLY batch for c.
func (s *session) addNames(c *channel, entries []string) {
	if c.pending == nil {
		c.pending = make(map[string]*member)
	}
	for _, entry := range entries {
		if m := s.parseMember(entry); m.nick != "" {
			c.pending[irc.Fold(m.nick)] = m
		}
	}
}

// commitNames replaces c's members with the buffered names.
func (s *session) commitNames(c *channel) {
	if c.pending != nil {
		c.members, c.pending = c.pending, nil
	}
}

// applyModes applies a channel MODE change to c's members.
func (s *session) applyModes(c *channel, modes string, params []string) {
	nextParam := func() string {
		if len(params) == 0 {
			return ""
		}
		p := params[0]
		params = params[1:]
		return p
	}

	add := true
	for _, mode := range modes {
		switch {
		case mode == '+':
			add = true
		case mode == '-':
			add = false
		case strings.ContainsRune(s.prefixModes, mode):
			m := c.member(nextParam())
			if m == nil {
				continue
			}
			if add && !strings.ContainsRune(m.modes, mode) {
				m.modes += string(mode)
			} else if !add {
				m.modes = strings.ReplaceAll(m.modes, string(mode), "")
			}
		case strings.ContainsRune(s.chanModes[0]+s.chanModes[1], mode),
			add && strings.ContainsRune(s.chanModes[2], mode):
			nextParam() // the mode takes a parameter we do not track
		}
	}
}

// applyISupport applies RPL_ISUPPORT tokens such as "PREFIX=(ov)@+".
func (s *session) applyISupport(tokens []string) {
	for _, token := range tokens {
		key, value, _ := strings.Cut(token, "=")
		switch key {
		case "PREFIX":
			modes, symbols, ok := strings.Cut(strings.TrimPrefix(value, "("), ")")
			if ok && len(modes) == len(symbols) {
				s.prefixModes, s.prefixSymbols = modes, symbols
			}
		case "CHANMODES":
			types := strings.Split(value, ",")
			if len(types) >= 4 {
				copy(s.chanModes[:], types)
			}
		case "MODES":
			if n, err := strconv.Atoi(value); err == nil && n > 0 {
				s.maxModes = n
			}
		}
	}
}

// rank returns the position of m's highest mode in the prefix order, or
// len(prefixModes) if m has none.
func (s *session) rank(m *member) int {
	for i, mode := range s.prefixModes {
		if strings.ContainsRune(m.modes, mode) {
			return i
		}
	}
	return len(s.prefixModes)
}

// users returns c's members for display, operators first.
func (s *session) users(c *channel) []ui.User {
	members := make([]*member, 0, len(c.members))
	for _, m := range c.members {
		members = append(members, m)
	}
	slices.SortFunc(members, func(a, b *member) int {
		return cmp.Or(cmp.Compare(s.rank(a), s.rank(b)), cmp.Compare(irc.Fold(a.nick), irc.Fold(b.nick)))
	})

	users := make([]ui.User, len(members))
	for i, m := range members {
		users[i].Nick = m.nick
		if r := s.rank(m); r < len(s.prefixSymbols) {
			users[i].Prefix = string(s.prefixSymbols[r])
		}
	}
	return users
}
