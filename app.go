package main

import (
	"context"
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

// ErrDone is an error returned by a goroutine to say that we should exit the program.
var ErrDone = errors.New("done")

// App holds the state for the application.
type App struct {
	Config
	osc.Conn

	cancel context.CancelFunc
	ctx    context.Context
	group  *errgroup.Group

	errors  chan Error
	replies chan osc.Message
}

type cmdFunc func(args []string) error

// NewApp creates a new application.
func NewApp(ctx context.Context, config Config) (*App, error) {
	cctx, cancel := context.WithCancel(ctx)
	g, gctx := errgroup.WithContext(cctx)

	app := &App{
		Config: config,

		cancel: cancel,
		ctx:    gctx,
		group:  g,

		errors:  make(chan Error),
		replies: make(chan osc.Message),
	}
	if err := app.initialize(); err != nil {
		return nil, errors.Wrap(err, "could not initialize app")
	}
	return app, nil
}

// Close closes the app.
func (app *App) Close() error {
	close(app.errors)
	close(app.replies)
	return app.Conn.Close()
}

// Error handles error replies from gonzo.
func (app *App) Error(msg osc.Message) error {
	if len(msg.Arguments) != 3 {
		return errors.New("expected 3 arguments for error message")
	}
	address, err := msg.Arguments[0].ReadString()
	if err != nil {
		return errors.Wrap(err, "reading address in error message")
	}
	code, err := msg.Arguments[1].ReadInt32()
	if err != nil {
		return errors.Wrap(err, "reading code in error message")
	}
	errmsg, err := msg.Arguments[2].ReadString()
	if err != nil {
		return errors.Wrap(err, "reading errmsg in error message")
	}
	app.errors <- NewError(nsm.NewError(nsm.Code(code), errmsg), address)
	app.debugf("received error: address=%s code=%d message=%s", address, code, errmsg)
	return nil
}

// Go runs a new goroutine as part of an errgroup.Group
func (app *App) Go(f func() error) {
	app.group.Go(f)
}

// Ping sends a ping message.
func (app *App) Ping(args []string) error {
	return errors.Wrap(app.Send(osc.Message{Address: "/ping"}), "sending ping")
}

// Pong handles ping responses from gonzo.
func (app *App) Pong(msg osc.Message) error {
	fmt.Println("pong")
	return nil
}

// Reply handles replies from gonzo.
func (app *App) Reply(msg osc.Message) error {
	addr, err := msg.Arguments[0].ReadString()
	if err != nil {
		return errors.Wrap(err, "reading first argument of reply")
	}
	app.debugf("received reply for %s", addr)
	app.replies <- msg
	return nil
}

// Run runs the application.
func (app *App) Run() error {
	app.Go(app.ServeOSC)
	app.Go(app.run)

	app.debugf("initialized connection laddr=%s raddr=%s\n", app.LocalAddr(), app.RemoteAddr())

	return app.Wait()
}

// ServeOSC listens for osc methods to be invoked.
func (app *App) ServeOSC() error {
	if err := app.Serve(app.dispatcher()); err != nil {
		app.debugf("ServeOSC error %s", err)
		return err
	}
	return nil
}

// Wait waits for all the goroutines in an errgroup.Group
func (app *App) Wait() error {
	err := app.group.Wait()
	if err == ErrDone {
		return nil
	}
	return err
}

// WithCancel returns an osc method that calls the provided osc method and then cancels the app.
func (app *App) WithCancel(m osc.Method) osc.Method {
	return func(msg osc.Message) error {
		err := m(msg)
		app.cancel()
		return err
	}
}

// commands returns a map from command names to the functions that handle the commands.
func (app *App) commands() map[string]cmdFunc {
	return map[string]cmdFunc{
		"add":  withDone(app.Add),
		"help": withDone(usageCmd),
		"lc":   withDone(app.ListClients),
		"logs": withDone(app.ClientLogs),
		"ls":   withDone(app.ListSessions),
		"new":  withDone(app.NewSession),
		"rm":   withDone(app.RemoveSession),
		"ping": app.Ping,
	}
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

// dispatcher returns an osc dispatcher that handles replies from gonzo.
func (app *App) dispatcher() osc.Dispatcher {
	return osc.Dispatcher{
		nsm.AddressError: app.Error,
		"/pong":          app.WithCancel(app.Pong),
		nsm.AddressReply: app.Reply,
	}
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
	conn, err := osc.DialUDPContext(app.ctx, "udp", laddr, raddr)
	if err != nil {
		return errors.Wrap(err, "could not listen on udp")
	}
	app.Conn = conn

	return nil
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
	return run(args[1:])
}

// withDone returns a cmdFunc that returns ErrDone if f returns
// nil, and otherwise returns the error that f returns.
func withDone(f cmdFunc) cmdFunc {
	return func(args []string) error {
		if err := f(args); err != nil {
			return err
		}
		return ErrDone
	}
}

func init() {
	commandUsage["ping"] = func() error {
		fmt.Fprintf(os.Stderr, "Ping a gonzo server.\n")
		fmt.Fprintf(os.Stderr, "\n")
		fmt.Fprintf(os.Stderr, "Usage:\n")
		fmt.Fprintf(os.Stderr, "gonzoctl ping\n")
		fmt.Fprintf(os.Stderr, "\n")
		return nil
	}
}
