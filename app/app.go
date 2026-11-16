// Package app wires the IRC connection, the session state and the terminal
// UI together.
//
// All session state is owned by a single event loop goroutine. The network
// reader and the UI communicate with it over channels: incoming server
// messages and user input are events, and the loop's only output is calls to
// the connection and to the (goroutine-safe) UI.
package app

import (
	"context"
	"crypto/tls"
	"errors"
	"io"
	"sync/atomic"
	"time"

	"girc/config"
	"girc/irc"
	"girc/ui"
)

const (
	dialTimeout        = 30 * time.Second
	defaultQuitMessage = "Bye bye"
	inputQueueSize     = 64
)

var errNotConnected = errors.New("not connected to the server")

// sender sends commands to the IRC server.
type sender interface {
	Send(command string, params ...string) error
}

// view displays output to the user. See ui.UI.
type view interface {
	Println(buffer, line string)
	OpenBuffer(name string, focus bool)
	CloseBuffer(name string)
	RenameBuffer(oldName, newName string)
	ClearBuffer(name string)
	Highlight(name string)
	SetUsers(name string, users []ui.User)
	SetTopic(name, topic string)
}

// input is a line typed by the user in a buffer.
type input struct {
	buffer string
	text   string
}

// App is an IRC client session.
type App struct {
	cfg      config.Config
	conn     sender
	view     view
	commands map[string]command
	inputs   chan input
	done     chan struct{} // closed when the event loop exits
	quitting atomic.Bool   // set once the user asked to quit

	// Owned by the event loop goroutine.
	s         *session
	connected bool
}

func newApp(cfg config.Config, conn sender, v view) *App {
	a := &App{
		cfg:       cfg,
		conn:      conn,
		view:      v,
		inputs:    make(chan input, inputQueueSize),
		done:      make(chan struct{}),
		s:         newSession(cfg.Nick),
		connected: true,
	}
	a.commands = a.newCommands()
	return a
}

// Run connects to the configured server and runs the UI until the user quits.
func Run(ctx context.Context, cfg config.Config) error {
	var tlsConfig *tls.Config
	if cfg.TLS {
		tlsConfig = &tls.Config{
			ServerName:         cfg.Server,
			InsecureSkipVerify: cfg.InsecureSkipVerify, //nolint:gosec // explicit opt-in via config
		}
	}

	dialCtx, cancelDial := context.WithTimeout(ctx, dialTimeout)
	defer cancelDial()
	conn, err := irc.Dial(dialCtx, cfg.Addr(), tlsConfig)
	if err != nil {
		return err
	}
	defer conn.Close()

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	a := newApp(cfg, conn, nil)
	u := ui.New(a.submit)
	a.view = u

	messages := make(chan irc.Message)
	readErr := make(chan error, 1)
	go func() {
		readErr <- conn.ReadLoop(func(m irc.Message) {
			select {
			case messages <- m:
			case <-ctx.Done():
			}
		})
	}()

	go func() {
		a.loop(ctx, messages, readErr)
		u.Stop()
	}()

	err = u.Run()
	cancel()
	<-a.done

	// The UI was closed (e.g. Ctrl+C) without /quit: say goodbye to the server.
	if a.quitting.CompareAndSwap(false, true) {
		_ = conn.Send("QUIT", defaultQuitMessage)
	}
	return err
}

// submit queues a line typed by the user. It is called on the UI goroutine
// and never blocks it.
func (a *App) submit(buffer, text string) {
	in := input{buffer: buffer, text: text}
	select {
	case a.inputs <- in:
	default:
		go func() {
			select {
			case a.inputs <- in:
			case <-a.done:
			}
		}()
	}
}

// loop is the event loop. It returns when ctx is cancelled, or when the
// server closes the connection after the user quit.
func (a *App) loop(ctx context.Context, messages <-chan irc.Message, readErr <-chan error) {
	defer close(a.done)

	a.printInfo(ui.StatusBuffer, "Connected to "+a.cfg.Addr())
	if err := a.register(); err != nil {
		a.printError(ui.StatusBuffer, err)
	}

	for {
		select {
		case <-ctx.Done():
			return
		case m := <-messages:
			a.handleMessage(m)
		case in := <-a.inputs:
			a.handleInput(in)
			if a.quitting.Load() && !a.connected {
				return // /quit after the connection was lost
			}
		case err := <-readErr:
			a.connected = false
			if a.quitting.Load() {
				return
			}
			if errors.Is(err, io.EOF) {
				err = errors.New("disconnected by the server")
			}
			a.printError(ui.StatusBuffer, err)
			a.printInfo(ui.StatusBuffer, "Press Ctrl+C to exit")
			a.view.Highlight(ui.StatusBuffer)
		}
	}
}

// send sends a command to the server.
func (a *App) send(command string, params ...string) error {
	if !a.connected {
		return errNotConnected
	}
	return a.conn.Send(command, params...)
}

// register sends the NICK and USER commands that open the session.
func (a *App) register() error {
	if err := a.send("NICK", a.cfg.Nick); err != nil {
		return err
	}
	return a.send("USER", a.cfg.User, "0", "*", a.cfg.RealName)
}

// refreshUsers pushes c's user list to the UI.
func (a *App) refreshUsers(c *channel) {
	a.view.SetUsers(c.name, a.s.users(c))
}
