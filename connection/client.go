package connection

import (
	"fmt"
	"log"
	"net"
	"strconv"
)

// Connection is the connection to the IRC server
type Connection struct {
	Conn net.Conn
}

// Client is the main struct for the IRC client
type Client struct {
	Server    string           // Server is the server to connect to
	Port      int              // Port is the port to connect to
	Nick      string           // Nick is the nickname to use
	User      string           // User is the username to use
	RealName  string           // RealName is the real name to use
	Conn      *Connection      // Conn is the connection to the server
	Channel   string           // joined channel
	DoneCh    chan interface{} // done channel
	ReadCh    chan []byte      // read channel
	Presenter *Presenter       // presenter formats the received messages
}

// NewClient creates a new IRC client
func NewClient(ch chan []byte, done chan interface{}) *Client {
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
		DoneCh:   done,
		ReadCh:   ch,
	}
}

// Connect connects to the IRC server, sends the NICK and USER commands
// and starts reading from the connection
func (c *Client) Connect() error {
	fmt.Println("Connecting to server...")
	conn, err := net.Dial("tcp", c.Server+":"+strconv.Itoa(c.Port))
	if err != nil {
		log.Printf("Error connecting to server: %s", err)
	}

	c.Conn = &Connection{
		Conn: conn,
	}

	// Send NICKname
	_, err = conn.Write([]byte("NICK " + c.Nick + "\r\n"))
	if err != nil {
		log.Printf("Error writing to connection: %s", err)
		return err
	}

	// Send USERname
	_, err = conn.Write([]byte("USER " + c.User + " 0 * :" + c.RealName + "\r\n"))
	if err != nil {
		log.Printf("Error writing to connection: %s", err)
		return err
	}

	c.Read()
	return nil
}

// Read reads from the connection
func (c *Client) Read() {
	buf := make([]byte, 4096)

	go func() {
		for {
			n, err := c.Conn.Conn.Read(buf)
			if err != nil {
				log.Fatalf("Error reading from connection: %s", err)
			}

			select {
			case <-c.DoneCh:
				return
			default:
				c.ReadCh <- buf[:n]
			}
		}
	}()
}

// SendCommand sends a message/command to the irc server
func (c *Client) SendCommand(msg string) {
	_, err := c.Conn.Conn.Write([]byte(msg))
	if err != nil {
		log.Printf("Error writing to connection: %s", err)
	}
}

// PrintMessage prints the message to the console via the channel
func (c *Client) PrintMessage(msg string) {
	c.ReadCh <- []byte(msg)
}

// Close closes the connection to the server
func (c *Client) Close() {
	c.Conn.Conn.Close()
}
