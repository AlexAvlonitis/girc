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
func NewConnection(client interfaces.Client) interfaces.Connection {
	address := client.Server() + ":" + strconv.Itoa(client.Port())
	var conn net.Conn
	var err error

	if client.Ssl() {
		conn, err = createSSLConnection(address)
	} else {
		conn, err = createTCPConnection(address)
	}

	if err != nil {
		log.Fatalf("Error connecting to server: %s", err)
	}

	return &DefaultConnection{conn: conn}
}

func createSSLConnection(address string) (net.Conn, error) {
	fmt.Println("Connecting to server over SSL...")

	tlsConfig := &tls.Config{
		InsecureSkipVerify: true, // Skip certificate verification (unsafe for production)
	}

	return tls.Dial("tcp4", address, tlsConfig)
}

func createTCPConnection(address string) (net.Conn, error) {
	fmt.Println("Connecting to server...")

	return net.Dial("tcp4", address)
}
