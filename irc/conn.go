package irc

import (
	"bufio"
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"sync"
	"time"
)

const (
	// maxLineLength covers the 512-byte RFC limit plus IRCv3 message tags.
	maxLineLength = 8191 + 512
	writeTimeout  = 30 * time.Second
)

// Conn is a connection to an IRC server. Send is safe for concurrent use.
type Conn struct {
	conn    net.Conn
	writeMu sync.Mutex
}

// Dial connects to addr. If tlsConfig is non-nil the connection uses TLS.
func Dial(ctx context.Context, addr string, tlsConfig *tls.Config) (*Conn, error) {
	var (
		dialer net.Dialer
		nc     net.Conn
		err    error
	)
	if tlsConfig != nil {
		td := tls.Dialer{NetDialer: &dialer, Config: tlsConfig}
		nc, err = td.DialContext(ctx, "tcp", addr)
	} else {
		nc, err = dialer.DialContext(ctx, "tcp", addr)
	}
	if err != nil {
		return nil, fmt.Errorf("dial %s: %w", addr, err)
	}
	return &Conn{conn: nc}, nil
}

// NewConn wraps an existing network connection.
func NewConn(nc net.Conn) *Conn {
	return &Conn{conn: nc}
}

// Send writes a single command to the server.
func (c *Conn) Send(command string, params ...string) error {
	line := Message{Command: command, Params: params}.String() + "\r\n"

	c.writeMu.Lock()
	defer c.writeMu.Unlock()

	if err := c.conn.SetWriteDeadline(time.Now().Add(writeTimeout)); err != nil {
		return fmt.Errorf("send %s: %w", command, err)
	}
	if _, err := io.WriteString(c.conn, line); err != nil {
		return fmt.Errorf("send %s: %w", command, err)
	}
	return nil
}

// ReadLoop reads messages from the server and calls handle for each one,
// until the connection is closed. It returns io.EOF when the server closes
// the connection cleanly. Malformed lines are skipped.
func (c *Conn) ReadLoop(handle func(Message)) error {
	sc := bufio.NewScanner(c.conn)
	sc.Buffer(make([]byte, 0, 4096), maxLineLength)
	for sc.Scan() {
		msg, err := ParseMessage(sc.Text())
		if err != nil {
			continue
		}
		handle(msg)
	}
	if err := sc.Err(); err != nil {
		return err
	}
	return io.EOF
}

// Close closes the underlying connection.
func (c *Conn) Close() error {
	return c.conn.Close()
}
