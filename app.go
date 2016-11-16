package main

import (
	"fmt"
	"net"
	"os"
	"strconv"

	"github.com/pkg/errors"
	"github.com/scgolang/nsm"
	"github.com/scgolang/osc"
)

// App holds the state for the application.
type App struct {
	Config
	osc.Conn
}

// NewApp creates a new application.
func NewApp(config Config) (*App, error) {
	app := &App{Config: config}
	if err := app.initialize(); err != nil {
		return nil, errors.Wrap(err, "could not initialize app")
	}
	return app, nil
}

// initialize initializes the application.
func (app *App) initialize() error {
	a := net.JoinHostPort(app.Host, strconv.Itoa(app.Port))
	addr, err := net.ResolveUDPAddr("udp", a)
	if err != nil {
		return errors.Wrap(err, "could not resolve udp address")
	}
	conn, err := osc.DialUDP("udp", nil, addr)
	if err != nil {
		return errors.Wrap(err, "could not listen on udp")
	}
	app.Conn = conn
	return nil
}

// Add tells gonzo to add a client.
func (app *App) Add(args []string) error {
	if len(args) != 1 {
		return errors.New("add takes exactly one arg")
	}

	msg, err := osc.NewMessage(nsm.AddressServerAdd)
	if err != nil {
		return errors.Wrap(err, "could not create osc message")
	}

	progname := args[0]
	if err := msg.WriteString(progname); err != nil {
		return errors.Wrap(err, "could not add progname to message")
	}
	if err := app.Send(msg); err != nil {
		return errors.Wrap(err, "could not send add message")
	}
	return nil
}

// Run runs the application.
func (app *App) Run() error {
	args := app.flags.Args()
	if len(args) == 0 {
		return fmt.Errorf("%s needs a command", os.Args[0])
	}
	command := args[0]
	switch command {
	case "add":
		return app.Add(args[1:])
	}
	return nil
}
