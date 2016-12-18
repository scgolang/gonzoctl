package main

import (
	"fmt"
	"os"
	"time"

	"github.com/pkg/errors"
	"github.com/scgolang/nsm"
	"github.com/scgolang/osc"
)

// ListSessions lists the sessions managed by a gonzo server.
func (app *App) ListSessions(args []string) error {
	if err := app.Send(osc.Message{
		Address: nsm.AddressServerProjects,
	}); err != nil {
		return errors.Wrap(err, "sending message")
	}
	timeout := time.After(app.Timeout)

	app.debug("waiting for reply")

	select {
	case reply := <-app.replies:
		app.debug("got reply")

		if err := app.printProjectFrom(reply); err != nil {
			return errors.Wrap(err, "printing project")
		}
	case <-timeout:
		return errors.New("timeout")
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
	if addr != nsm.AddressServerProjects {
		// TODO: requeue message
	}
	numSessions, err := msg.Arguments[1].ReadInt32()
	if err != nil {
		return errors.Wrap(err, "reading number of sessions from osc message")
	}
	if expected, got := numSessions+2, int32(len(msg.Arguments)); expected != got {
		return errors.Errorf("expected %d arguments, got %d", expected, got)
	}
	for i := int32(0); i < numSessions; i++ {
		project, err := msg.Arguments[i+2].ReadString()
		if err != nil {
			return errors.Wrap(err, "reading project from osc message")
		}
		if _, err := fmt.Println(project); err != nil {
			return errors.Wrap(err, "printing project")
		}
	}
	return nil
}

func init() {
	commandUsage["ls"] = func() error {
		fmt.Fprintf(os.Stderr, "List sessions.\n")
		fmt.Fprintf(os.Stderr, "\n")
		fmt.Fprintf(os.Stderr, "Usage:\n")
		fmt.Fprintf(os.Stderr, "gonzoctl ls\n")
		fmt.Fprintf(os.Stderr, "\n")
		return ErrDone
	}
}
