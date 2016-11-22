package main

import (
	"fmt"
	"log"
	"time"

	"github.com/pkg/errors"
	"github.com/scgolang/nsm"
	"github.com/scgolang/osc"
)

// ListProjects lists the projects managed by a gonzo server.
func (app *App) ListProjects(args []string) error {
	msg, err := osc.NewMessage(nsm.AddressServerList)
	if err != nil {
		return errors.Wrap(err, "create osc message")
	}
	if err := app.Send(msg); err != nil {
		return errors.Wrap(err, "sending message")
	}
	for {
		timeout := time.After(app.Timeout)

		log.Println("waiting for reply")

		select {
		case reply := <-app.replies:
			log.Println("got reply")
			if err := app.printProjectFrom(reply); err != nil {
				if err == ErrDone {
					return nil
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
func (app *App) printProjectFrom(msg *osc.Message) error {
	addr, err := msg.ReadString()
	if err != nil {
		return errors.Wrap(err, "reading reply address from osc message")
	}
	if addr != nsm.AddressServerList {
		// TODO: requeue message
	}
	project, err := msg.ReadString()
	if err != nil {
		return errors.Wrap(err, "reading project from osc message")
	}
	if project == nsm.DoneString {
		return ErrDone
	}
	_, err = fmt.Println(project)
	return errors.Wrap(err, "printing project")
}
