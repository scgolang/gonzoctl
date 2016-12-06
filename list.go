package main

import (
	"fmt"
	"time"

	"github.com/pkg/errors"
	"github.com/scgolang/nsm"
	"github.com/scgolang/osc"
)

// ListProjects lists the projects managed by a gonzo server.
func (app *App) ListProjects(args []string) error {
	if err := app.Send(osc.Message{
		Address: nsm.AddressServerList,
	}); err != nil {
		return errors.Wrap(err, "sending message")
	}
	for {
		timeout := time.After(app.Timeout)

		app.debug("waiting for reply")

		select {
		case reply := <-app.replies:
			app.debug("got reply")

			if err := app.printProjectFrom(reply); err != nil {
				if err == ErrDone {
					return err
				}
				return errors.Wrap(err, "printing project")
			}
		case <-timeout:
			return errors.New("timeout")
		}
	}
	return nil
}

// printProjectFrom prints a project from an OSC reply to /nsm/server/list
func (app *App) printProjectFrom(msg osc.Message) error {
	if len(msg.Arguments) < 2 {
		return errors.New("expected two arguments")
	}
	addr, err := msg.Arguments[0].ReadString()
	if err != nil {
		return errors.Wrap(err, "reading reply address from osc message")
	}
	if addr != nsm.AddressServerList {
		// TODO: requeue message
	}
	project, err := msg.Arguments[1].ReadString()
	if err != nil {
		return errors.Wrap(err, "reading project from osc message")
	}
	if project == nsm.DoneString {
		return ErrDone
	}
	_, err = fmt.Println(project)
	return errors.Wrap(err, "printing project")
}
