package main

import (
	"fmt"
	"net"
	"os"
	"strconv"

	"github.com/pkg/errors"
	"github.com/scgolang/osc"
)

// Sentinel errors.
var (
	ErrDone = errors.New("done")
)

// App holds the state for the application.
type App struct {
	Config
	osc.Conn

	replies chan *osc.Message
}

type cmdFunc func(args []string) error

// NewApp creates a new application.
func NewApp(config Config) (*App, error) {
	app := &App{
		Config:  config,
		replies: make(chan *osc.Message),
	}
	if err := app.initialize(); err != nil {
		return nil, errors.Wrap(err, "could not initialize app")
	}
	return app, nil
}

// initialize initializes the application.
func (app *App) initialize() error {
	// Initialize the OSC connection.
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

// commands returns a map from command names to the functions that handle the commands.
func (app *App) commands() map[string]cmdFunc {
	return map[string]cmdFunc{
		"add": app.Add,
		"ls":  app.ListProjects,
	}
}

// Run runs the application.
func (app *App) Run() error {
	defer close(app.replies)

	args := app.flags.Args()
	if len(args) == 0 {
		return fmt.Errorf("%s needs a command", os.Args[0])
	}
	var (
		command  = args[0]
		commands = app.commands()
	)
	run, ok := commands[command]
	if !ok {
		return errors.New("unrecognized command: " + command)
	}
	return errors.Wrapf(run(args[1:]), "running command %s", command)
}
