package connection_manager

import "girc/interfaces"

type User struct {
	Nick string
}

type Channel struct {
	Name  string
	Users []User
}

type DefaultConnectionManager struct {
	Client         interfaces.Client
	joinedChannels map[string]Channel
	currentChannel Channel
}

// JoinChannel adds a channel to the connection manager list of joinedChannels
// and sets it as the current channel
func (cm *DefaultConnectionManager) JoinChannel(name string) {
	cm.joinedChannels[name] = Channel{Name: name}
	cm.currentChannel = cm.joinedChannels[name]
}

// AddUsersToCurrentChannel adds a list of users to the current channel
func (cm *DefaultConnectionManager) AddUsersToCurrentChannel(users []User) {
	cm.currentChannel.Users = users
}

func NewConnectionManager(c interfaces.Client) *DefaultConnectionManager {
	return &DefaultConnectionManager{
		joinedChannels: make(map[string]Channel),
		Client:         c,
	}
}
