package main

import (
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
	if err := app.Send(osc.Message{
		Address: nsm.AddressServerAdd,
		Arguments: osc.Arguments{
			osc.String(args[0]),
			osc.String(args[1]),
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
