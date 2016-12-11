package main

import (
	"fmt"
	"time"

	"github.com/pkg/errors"
	"github.com/scgolang/nsm"
	"github.com/scgolang/osc"
)

// ListClients lists the clients currently being managed by a gonzo server.
func (app *App) ListClients(args []string) error {
	if err := app.Send(osc.Message{
		Address: nsm.AddressServerClients,
	}); err != nil {
		return errors.Wrap(err, "sending message")
	}
	timeout := time.After(app.Timeout)

	app.debug("waiting for reply")

	select {
	case reply := <-app.replies:
		app.debug("got reply")

		if err := app.printClientFrom(reply); err != nil {
			return errors.Wrap(err, "printing client")
		}
		return ErrDone
	case <-timeout:
		return errors.New("timeout")
	}
	return nil
}

// printClientsFrom prints a client from an OSC reply to /nsm/server/clients
func (app *App) printClientFrom(msg osc.Message) error {
	const numClientFields = 6

	if len(msg.Arguments) < 2 {
		return errors.New("expected two arguments")
	}
	addr, err := msg.Arguments[0].ReadString()
	if err != nil {
		return errors.Wrap(err, "reading reply address from osc message")
	}
	if addr != nsm.AddressServerClients {
		// TODO: requeue message
	}
	numClients, err := msg.Arguments[1].ReadInt32()
	if err != nil {
		return errors.Wrap(err, "reading number of clients from osc message")
	}
	if expected, got := (numClients*numClientFields)+2, int32(len(msg.Arguments)); expected != got {
		return errors.Errorf("expected %d arguments, got %d", expected, got)
	}
	for i := int32(0); i < numClients; i++ {
		j := i * numClientFields
		clientName, err := msg.Arguments[j+2].ReadString()
		if err != nil {
			return errors.Wrap(err, "reading client from osc message")
		}
		if _, err := fmt.Println(clientName); err != nil {
			return errors.Wrap(err, "printing client")
		}
	}
	return nil
}
