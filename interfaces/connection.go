package interfaces

import "net"

// Connection interface defines the methods that a connection should implement
type Connection interface {
	Conn() net.Conn
}
