package app

import (
	"slices"
	"strings"
	"sync"
)

// state is the client's view of the session. It is shared between the UI
// goroutine (commands) and the network goroutine (incoming messages).
type state struct {
	mu         sync.Mutex
	nick       string
	channel    string            // current channel, "" if none
	users      map[string]string // users in the current channel, keyed by folded nick
	names      []string          // RPL_NAMREPLY entries collected until RPL_ENDOFNAMES
	registered bool
}

func newState(nick string) *state {
	return &state{nick: nick, users: make(map[string]string)}
}

func (s *state) Nick() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.nick
}

func (s *state) SetNick(nick string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.nick = nick
}

// IsMe reports whether nick is the client's own nickname.
func (s *state) IsMe(nick string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return strings.EqualFold(s.nick, nick)
}

func (s *state) Channel() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.channel
}

// IsCurrent reports whether channel is the current channel.
func (s *state) IsCurrent(channel string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.channel != "" && strings.EqualFold(s.channel, channel)
}

func (s *state) Registered() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.registered
}

func (s *state) SetRegistered() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.registered = true
}

// Join makes channel the current channel and clears the user list.
func (s *state) Join(channel string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.channel = channel
	clear(s.users)
	s.names = nil
}

// Leave clears the current channel and its user list.
func (s *state) Leave() {
	s.Join("")
}

func (s *state) AddUser(nick string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.users[strings.ToLower(nick)] = nick
}

// RemoveUser removes nick from the user list and reports whether it was present.
func (s *state) RemoveUser(nick string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	key := strings.ToLower(nick)
	_, ok := s.users[key]
	delete(s.users, key)
	return ok
}

// RenameUser renames a user and reports whether they were in the list.
func (s *state) RenameUser(oldNick, newNick string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	key := strings.ToLower(oldNick)
	if _, ok := s.users[key]; !ok {
		return false
	}
	delete(s.users, key)
	s.users[strings.ToLower(newNick)] = newNick
	return true
}

// AddNames buffers a batch of nicknames from RPL_NAMREPLY.
func (s *state) AddNames(nicks []string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.names = append(s.names, nicks...)
}

// CommitNames replaces the user list with the buffered names.
func (s *state) CommitNames() {
	s.mu.Lock()
	defer s.mu.Unlock()
	clear(s.users)
	for _, nick := range s.names {
		s.users[strings.ToLower(nick)] = nick
	}
	s.names = nil
}

// Users returns the current channel's users, sorted case-insensitively.
func (s *state) Users() []string {
	s.mu.Lock()
	defer s.mu.Unlock()
	users := make([]string, 0, len(s.users))
	for _, nick := range s.users {
		users = append(users, nick)
	}
	slices.SortFunc(users, func(a, b string) int {
		return strings.Compare(strings.ToLower(a), strings.ToLower(b))
	})
	return users
}
