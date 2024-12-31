package connection

import (
	"crypto/tls"
	"fmt"
	"girc/interfaces"
	"log"
	"net"
	"strconv"
)

// Connection is the connection to the IRC server
type DefaultConnection struct {
	conn net.Conn
}

func (c *DefaultConnection) Conn() net.Conn { return c.conn }

// NewConnection creates a new connection to the server
func NewConnection(c interfaces.Client) interfaces.Connection {
	var conn net.Conn
	var err error

	if c.Ssl() {
		fmt.Println("Connecting to server over SSL...")

		tlsConfig := &tls.Config{
			InsecureSkipVerify: true, // Skip certificate verification (unsafe for production)
		}

		conn, err = tls.Dial("tcp4", c.Server()+":"+strconv.Itoa(c.Port()), tlsConfig)
		if err != nil {
			log.Fatalf("Error connecting to server: %s", err)
		}
	} else {
		fmt.Println("Connecting to server...")
		conn, err = net.Dial("tcp4", c.Server()+":"+strconv.Itoa(c.Port()))
		if err != nil {
			log.Fatalf("Error connecting to server: %s", err)
		}
	}

	return &DefaultConnection{conn: conn}
}
