package main

import (
	"fmt"
	"os"
	"time"

	"github.com/pkg/errors"
	"github.com/scgolang/nsm"
	"github.com/scgolang/osc"
)

// Remove removes a session.
func (app *App) RemoveSession(args []string) error {
	if len(args) < 1 {
		return errors.New("add takes exactly one argument")
	}
	name := args[0]
	if err := app.Send(osc.Message{
		Address: nsm.AddressServerRemove,
		Arguments: osc.Arguments{
			osc.String(name),
		},
	}); err != nil {
		return errors.Wrap(err, "sending remove message")
	}

	select {
	case <-time.After(2 * time.Second):
		return errors.New("timeout")
	case err := <-app.errors:
		return err
	case reply := <-app.replies:
		app.debugf("got reply %s", reply)
	}
	return nil
}

func init() {
	commandUsage["rm"] = func() error {
		fmt.Fprintf(os.Stderr, "Remove a session.\n")
		fmt.Fprintf(os.Stderr, "\n")
		fmt.Fprintf(os.Stderr, "Usage:\n")
		fmt.Fprintf(os.Stderr, "gonzoctl rm NAME\n")
		fmt.Fprintf(os.Stderr, "\n")
		fmt.Fprintf(os.Stderr, "NAME      The name of the session.\n")
		fmt.Fprintf(os.Stderr, "\n")
		fmt.Fprintf(os.Stderr, "Example:\n")
		fmt.Fprintf(os.Stderr, "gonzoctl rm session1\n")
		return nil
	}
}
