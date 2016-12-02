package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"strconv"

	"github.com/pkg/errors"
	"github.com/scgolang/nsm"
	"github.com/scgolang/osc"
	"golang.org/x/sync/errgroup"
)

// Sentinel errors.
var (
	ErrDone = errors.New("done")
)

// App holds the state for the application.
type App struct {
	Config
	osc.Conn

	replies chan osc.Message
}

type cmdFunc func(args []string) error

// NewApp creates a new application.
func NewApp(config Config) (*App, error) {
	app := &App{
		Config:  config,
		replies: make(chan osc.Message),
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
	raddr, err := net.ResolveUDPAddr("udp", a)
	if err != nil {
		return errors.Wrap(err, "could not resolve remote udp address")
	}
	laddr, err := net.ResolveUDPAddr("udp", "0.0.0.0:0")
	if err != nil {
		return errors.Wrap(err, "could not resolve local udp address")
	}
	conn, err := osc.DialUDP("udp", laddr, raddr)
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

// dispatcher returns an osc dispatcher that handles replies from gonzo.
func (app *App) dispatcher() osc.Dispatcher {
	return osc.Dispatcher{
		nsm.AddressReply: func(msg osc.Message) error {
			app.debug("received reply")
			app.replies <- msg
			return nil
		},
	}
}

// ServeOSC listens for osc methods to be invoked.
func (app *App) ServeOSC() error {
	return app.Serve(app.dispatcher())
}

// Run runs the application.
func (app *App) Run() error {
	defer close(app.replies)

	var eg errgroup.Group

	eg.Go(app.ServeOSC)
	eg.Go(app.run)

	app.debugf("initialized connection laddr=%s raddr=%s\n", app.LocalAddr(), app.RemoteAddr())

	return eg.Wait()
}

// run runs the command we have invoked.
func (app *App) run() error {
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

// debug prints a debug message.
func (app *App) debug(msg string) {
	if app.Debug {
		log.Println(msg)
	}
}

// debugf prints a debug message with printf semantics.
func (app *App) debugf(format string, args ...interface{}) {
	if app.Debug {
		log.Printf(format, args...)
	}
}
