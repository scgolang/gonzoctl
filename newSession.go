package main

import (
	"fmt"
	"os"
	"time"

	"github.com/pkg/errors"
	"github.com/scgolang/nsm"
	"github.com/scgolang/osc"
)

// NewSession creates a new session.
func (app *App) NewSession(args []string) error {
	if expected, got := 1, len(args); expected != got {
		return errors.Errorf("expected %d arguments, got %d", expected, got)
	}
	name := args[0]
	if err := app.Send(osc.Message{
		Address: nsm.AddressServerNew,
		Arguments: osc.Arguments{
			osc.String(name),
		},
	}); err != nil {
		return errors.Wrap(err, "sending message")
	}
	timeout := time.After(app.Timeout)

	app.debug("waiting for reply")

	select {
	case errReply := <-app.errors:
		app.debug("got error " + errReply.Error())
	case reply := <-app.replies:
		app.debug("got reply for " + reply.Address)
	case <-timeout:
		return errors.New("timeout")
	}
	return nil
}

func init() {
	commandUsage["new"] = func() error {
		fmt.Fprintf(os.Stderr, "Create a new session.\n")
		fmt.Fprintf(os.Stderr, "\n")
		fmt.Fprintf(os.Stderr, "Usage:\n")
		fmt.Fprintf(os.Stderr, "gonzoctl new NAME\n")
		fmt.Fprintf(os.Stderr, "\n")
		return nil
	}
}
