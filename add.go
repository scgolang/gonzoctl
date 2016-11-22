package main

import (
	"github.com/pkg/errors"
	"github.com/scgolang/nsm"
	"github.com/scgolang/osc"
)

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
