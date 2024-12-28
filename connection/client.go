package connection

import (
	"fmt"
	"log"
	"net"
	"strconv"
)

// Connection is the connection to the IRC server
type Connection struct {
	// Conn is the connection to the server
	Conn net.Conn
}

// Client is the main struct for the IRC client
type Client struct {
	// Server is the server to connect to
	Server string
	// Port is the port to connect to
	Port int
	// Nick is the nickname to use
	Nick string
	// User is the username to use
	User string
	// RealName is the real name to use
	RealName string
	// Conn is the connection to the server
	Conn *Connection
	// joined channel
	Channel string
}

// NewClient creates a new IRC client
func NewClient() *Client {
	cfg, err := NewConfiguration()
	if err != nil {
		log.Fatalf("Error reading configuration: %s", err)
		return nil
	}

	return &Client{
		Server:   cfg.Server,
		Port:     cfg.Port,
		Nick:     cfg.Nick,
		User:     cfg.User,
		RealName: cfg.RealName,
	}
}

// Connect connects to the IRC server
func (c *Client) Connect() error {
	fmt.Println("Connecting to server...")
	conn, err := net.Dial("tcp", c.Server+":"+strconv.Itoa(c.Port))
	if err != nil {
		log.Printf("Error connecting to server: %s", err)
	}

	c.Conn = &Connection{
		Conn: conn,
	}

	_, err = conn.Write([]byte("NICK " + c.Nick + "\r\n"))
	if err != nil {
		log.Printf("Error writing to connection: %s", err)
		return err
	}

	_, err = conn.Write([]byte("USER " + c.User + " 0 * :" + c.RealName + "\r\n"))
	if err != nil {
		log.Printf("Error writing to connection: %s", err)
		return err
	}

	return nil
}

// reads from the connection
func (c *Client) Read(ch chan []byte, done chan interface{}) {
	buf := make([]byte, 512)
	for {
		n, err := c.Conn.Conn.Read(buf)
		if err != nil {
			log.Fatalf("Error reading from connection: %s", err)
		}
		fmt.Printf("%s", buf[:n])

		select {
		case <-done:
			return
		default:
			ch <- buf[:n]
		}
	}
}

// write to connection
func (c *Client) Write(msg string) {
	_, err := c.Conn.Conn.Write([]byte(msg))
	if err != nil {
		log.Printf("Error writing to connection: %s", err)
	}
}

// Close closes the connection to the server
func (c *Client) Close() {
	c.Conn.Conn.Close()
}
