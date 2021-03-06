package main

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/pkg/errors"
	"github.com/scgolang/nsm"
	"github.com/scgolang/osc"
)

// ListSessions lists the sessions managed by a gonzo server.
func (app *App) ListSessions(args []string) error {
	if err := app.Send(osc.Message{
		Address: nsm.AddressServerSessions,
	}); err != nil {
		return errors.Wrap(err, "sending message")
	}
	timeout := time.After(app.Timeout)

	app.debug("waiting for reply")

	select {
	case <-timeout:
		return errors.New("timeout")
	case reply := <-app.replies:
		app.debug("got reply")

		if err := app.printSessionFrom(reply); err != nil {
			return errors.Wrap(err, "printing project")
		}
	}
	return nil
}

// printSessionFrom prints a session from an OSC reply to /nsm/server/list
func (app *App) printSessionFrom(msg osc.Message) error {
	const minNumArgs = 3

	if len(msg.Arguments) < minNumArgs {
		return errors.New("expected two arguments")
	}
	addr, err := msg.Arguments[0].ReadString()
	if err != nil {
		return errors.Wrap(err, "reading reply address from osc message")
	}
	if addr != nsm.AddressServerSessions {
		// TODO: requeue message
	}
	numSessions, err := msg.Arguments[1].ReadInt32()
	if err != nil {
		return errors.Wrap(err, "reading number of sessions from osc message")
	}
	curridx, err := msg.Arguments[2].ReadInt32()
	if err != nil {
		return errors.Wrap(err, "reading current session index from osc message")
	}
	if expected, got := numSessions+minNumArgs, int32(len(msg.Arguments)); expected != got {
		return errors.Errorf("expected %d arguments, got %d", expected, got)
	}
	for i := int32(0); i < numSessions; i++ {
		project, err := msg.Arguments[i+minNumArgs].ReadString()
		if err != nil {
			return errors.Wrap(err, "reading project from osc message")
		}
		if i == curridx {
			fmt.Printf(" * ")
		} else {
			fmt.Printf("   ")
		}
		if _, err := fmt.Println(filepath.Base(project)); err != nil {
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
		return nil
	}
}
