package main

import (
	"fmt"
	"os"
	"time"

	"github.com/pkg/errors"
	"github.com/scgolang/nsm"
	"github.com/scgolang/osc"
)

// Add tells gonzo to add a client.
func (app *App) Add(args []string) error {
	if len(args) < 2 {
		return errors.New("add takes exactly two arguments")
	}
	var (
		name       = args[0]
		executable = args[1]
	)
	if err := app.Send(osc.Message{
		Address: nsm.AddressServerAdd,
		Arguments: osc.Arguments{
			osc.String(name),
			osc.String(executable),
		},
	}); err != nil {
		return errors.Wrap(err, "could not send add message")
	}

	select {
	case <-time.After(2 * time.Second):
		return errors.New("timeout")
	case reply := <-app.replies:
		app.debugf("got reply %s", reply)
	case err := <-app.errors:
		return err
	}
	return ErrDone
}

func init() {
	commandUsage["add"] = func() error {
		fmt.Fprintf(os.Stderr, "Add a new client to the current session.\n")
		fmt.Fprintf(os.Stderr, "\n")
		fmt.Fprintf(os.Stderr, "Usage:\n")
		fmt.Fprintf(os.Stderr, "gonzoctl add NAME PROGRAM\n")
		fmt.Fprintf(os.Stderr, "\n")
		fmt.Fprintf(os.Stderr, "NAME      is the name of the new client.\n")
		fmt.Fprintf(os.Stderr, "PROGRAM   is the path to the executable for the client.\n")
		fmt.Fprintf(os.Stderr, "\n")
		fmt.Fprintf(os.Stderr, "Example:\n")
		fmt.Fprintf(os.Stderr, "gonzoctl add sc-servers1 sc-servers\n")
		return ErrDone
	}
}
