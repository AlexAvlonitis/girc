package irc

import (
	"bufio"
	"errors"
	"io"
	"net"
	"testing"
)

func TestConnSendAndReadLoop(t *testing.T) {
	client, server := net.Pipe()
	c := NewConn(client)

	go func() {
		w := bufio.NewWriter(server)
		w.WriteString(":srv 001 me :Welcome\r\n")
		w.WriteString("\r\n") // malformed, should be skipped
		w.WriteString("PING :srv\r\n")
		w.Flush()

		line, _ := bufio.NewReader(server).ReadString('\n')
		if line != "PONG srv\r\n" {
			t.Errorf("server received %q", line)
		}
		server.Close()
	}()

	var got []string
	err := c.ReadLoop(func(m Message) {
		got = append(got, m.Command)
		if m.Command == "PING" {
			if err := c.Send("PONG", m.Params...); err != nil {
				t.Errorf("Send: %v", err)
			}
		}
	})
	if !errors.Is(err, io.EOF) {
		t.Errorf("ReadLoop error = %v, want io.EOF", err)
	}
	if len(got) != 2 || got[0] != "001" || got[1] != "PING" {
		t.Errorf("handled %q, want [001 PING]", got)
	}
}
