package connection

import (
	"fmt"
	"girc/commands"
	"girc/interfaces"
	"io"
	"log"
)

// DefaultClient is the main struct for the IRC client
type DefaultClient struct {
	server     string                // Server is the server to connect to
	port       int                   // Port is the port to connect to
	nick       string                // Nick is the nickname to use
	user       string                // User is the username to use
	realName   string                // RealName is the real name to use
	ssl        bool                  // Ssl is true if the connection is over SSL
	channel    string                // joined channel
	doneCh     chan interface{}      // done channel
	readCh     chan []byte           // read channel
	connection interfaces.Connection // Connection is the connection to the server
}

// Implement the methods to satisfy the Client interface
func (c *DefaultClient) Server() string                    { return c.server }
func (c *DefaultClient) Port() int                         { return c.port }
func (c *DefaultClient) Nick() string                      { return c.nick }
func (c *DefaultClient) User() string                      { return c.user }
func (c *DefaultClient) RealName() string                  { return c.realName }
func (c *DefaultClient) Ssl() bool                         { return c.ssl }
func (c *DefaultClient) Channel() string                   { return c.channel }
func (c *DefaultClient) Connection() interfaces.Connection { return c.connection }

// Setters
func (c *DefaultClient) SetChannel(channel string)          { c.channel = channel }
func (c *DefaultClient) SetNick(nick string)                { c.nick = nick }
func (c *DefaultClient) SetConn(conn interfaces.Connection) { c.connection = conn }

// NewClient creates a new IRC client
func NewClient(ch chan []byte, done chan interface{}) *DefaultClient {
	cfg, err := NewConfiguration()
	if err != nil {
		log.Fatalf("Error reading configuration: %s", err)
		return nil
	}

	return &DefaultClient{
		server:   cfg.Server,
		port:     cfg.Port,
		nick:     cfg.Nick,
		user:     cfg.User,
		realName: cfg.RealName,
		ssl:      cfg.Ssl,
		doneCh:   done,
		readCh:   ch,
	}
}

// Connect connects to the IRC server, sends the NICK and USER commands
// and starts reading from the connection
func (c *DefaultClient) Connect() error {
	conn := NewConnection(c)
	fmt.Println("1 Connection ID: ", conn)
	c.SetConn(conn)

	err := c.Register(c.Channel())
	if err != nil {
		log.Fatalf("Error registering with server: %s", err)
		return err
	}

	c.Read()
	return nil
}

// Read reads from the connection
func (c *DefaultClient) Read() {
	buf := make([]byte, 4096)

	go func() {
		for {
			n, err := c.Connection().Conn().Read(buf)
			if err != nil {
				if err == io.EOF {
					c.PrintMessage("Connection closed by server")
					close(c.doneCh)
					return
				}

				log.Fatalf("Error reading from connection: %s", err)
				close(c.doneCh)
				return
			}

			select {
			case <-c.doneCh:
				return
			default:
				c.readCh <- buf[:n]
			}
		}
	}()
}

// SendPong sends a PONG message to the server, to keep the connection alive
func (c *DefaultClient) SendPong(msg string) {
	pongCmd := commands.PongCommand{Input: msg, Client: c}
	err := pongCmd.Execute()
	if err != nil {
		fmt.Printf("Error sending PONG: %s", err)
	}
}

// Register sends the NICK and USER commands to the server, to register the client
func (c *DefaultClient) Register(channel string) error {
	// Send NICKname
	nickCmd := commands.NickCommand{Client: c}
	err := nickCmd.Execute()
	if err != nil {
		return err
	}

	// Send USERname
	userCmd := commands.UserCommand{Client: c}
	err = userCmd.Execute()
	if err != nil {
		return err
	}

	return nil
}

// Write sends a message/command to the IRC server
func (c *DefaultClient) Write(msg string) {
	fmt.Println(msg)
	_, err := c.Connection().Conn().Write([]byte(msg))
	if err != nil {
		log.Printf("Error writing to connection: %s", err)
	}
}

// PrintMessage prints the message to the console via the channel
func (c *DefaultClient) PrintMessage(msg string) {
	c.readCh <- []byte(msg)
}

// Close closes the connection to the server
func (c *DefaultClient) Close() {
	c.Connection().Conn().Close()
}
