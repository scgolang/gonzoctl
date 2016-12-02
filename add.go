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
	if err := app.Send(osc.Message{
		Address: nsm.AddressServerAdd,
		Arguments: osc.Arguments{
			osc.String(args[0]),
		},
	}); err != nil {
		return errors.Wrap(err, "could not send add message")
	}
	return nil
}
