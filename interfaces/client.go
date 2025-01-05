package interfaces

// Client interface defines the methods that a client should implement
type Client interface {
	Connect() error
	Read()
	SendPong(string)
	Register(string) error
	Write(string)
	PrintMessage(string)
	Close()
	Server() string
	Port() int
	Nick() string
	User() string
	RealName() string
	Ssl() bool
	Connection() Connection
	Channel() string
	Users() []string
	SetChannel(string)
	SetNick(string)
	SetConn(Connection)
	SetUsers([]string)
}
