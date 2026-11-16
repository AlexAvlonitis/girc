// Package app wires the IRC connection, the session state and the terminal
// UI together.
package app

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
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
)

// sender sends commands to the IRC server.
type sender interface {
	Send(command string, params ...string) error
}

// view displays output to the user.
type view interface {
	Println(line string)
	SetUsers(nicks []string)
}

// App is a running IRC client session.
type App struct {
	cfg      config.Config
	conn     sender
	view     view
	state    *state
	commands map[string]command
	quitting atomic.Bool
}

func newApp(cfg config.Config, conn sender, v view) *App {
	a := &App{
		cfg:   cfg,
		conn:  conn,
		view:  v,
		state: newState(cfg.Nick),
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

	dialCtx, cancel := context.WithTimeout(ctx, dialTimeout)
	defer cancel()
	conn, err := irc.Dial(dialCtx, cfg.Addr(), tlsConfig)
	if err != nil {
		return err
	}
	defer conn.Close()

	a := newApp(cfg, conn, nil)
	u := ui.New("GIRC - "+cfg.Server, a.handleInput)
	a.view = u

	go func() {
		a.serve(conn)
		if a.quitting.Load() {
			u.Stop()
		}
	}()

	err = u.Run()

	// The UI was closed (e.g. Ctrl+C) without /quit: say goodbye to the server.
	if a.quitting.CompareAndSwap(false, true) {
		_ = conn.Send("QUIT", defaultQuitMessage)
	}
	return err
}

// serve registers with the server and processes incoming messages until the
// connection is closed.
func (a *App) serve(conn *irc.Conn) {
	a.printInfo(fmt.Sprintf("Connected to %s", a.cfg.Addr()))

	if err := a.register(); err != nil {
		a.printError(err)
		return
	}

	err := conn.ReadLoop(a.handleMessage)
	if a.quitting.Load() {
		return
	}
	if errors.Is(err, io.EOF) {
		a.printError(errors.New("disconnected by server, press Ctrl+C to exit"))
	} else {
		a.printError(fmt.Errorf("connection lost: %w, press Ctrl+C to exit", err))
	}
}

// register sends the NICK and USER commands that open the session.
func (a *App) register() error {
	if err := a.conn.Send("NICK", a.cfg.Nick); err != nil {
		return err
	}
	return a.conn.Send("USER", a.cfg.User, "0", "*", a.cfg.RealName)
}

// printInfo prints an informational line. text is escaped.
func (a *App) printInfo(text string) {
	a.view.Println("[gray]-!-[-] " + ui.Escape(text))
}

// printError prints an error line.
func (a *App) printError(err error) {
	a.view.Println("[red]error:[-] " + ui.Escape(err.Error()))
}

// refreshUsers pushes the current user list to the UI.
func (a *App) refreshUsers() {
	a.view.SetUsers(a.state.Users())
}
