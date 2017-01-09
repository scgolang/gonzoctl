package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/pkg/errors"
	"github.com/scgolang/nsm"
	"github.com/scgolang/osc"
)

// logOutputOptions maps the CLI options for gonzoctl to the appropriate values in the gonzo OSC API.
var logOutputOptions = map[string]int32{"stderr": 2, "stdout": 1}

// ClientLogs gets the logs of a client.
func (app *App) ClientLogs(args []string) error {
	if minimum, got := 1, len(args); got < minimum {
		return errors.Errorf("expected at least %d argument(s), got %d", minimum, got)
	}
	var (
		fs         = flag.NewFlagSet("logs", flag.ExitOnError)
		outputFlag string
	)
	fs.StringVar(&outputFlag, "o", "stderr", "Output stream.")

	if err := fs.Parse(args); err != nil {
		return errors.Wrap(err, "parsing flags for logs command")
	}
	if expected, got := 1, len(fs.Args()); expected != got {
		return errors.New("expected client name in logs command")
	}
	clientName := fs.Args()[0]

	var outputArg int32
	outputArg, outputOK := logOutputOptions[outputFlag]
	if !outputOK {
		return errors.Errorf("expected output option to be either stderr or stdout")
	}
	if err := app.Send(osc.Message{
		Address: nsm.AddressClientLogs,
		Arguments: osc.Arguments{
			osc.String(clientName),
			osc.Int(outputArg),
		},
	}); err != nil {
		return errors.Wrap(err, "sending message to get client logs")
	}

	app.debug("waiting for reply")

	select {
	case errReply := <-app.errors:
		app.debug("got error " + errReply.Error())
	case reply := <-app.replies:
		return errors.Wrap(app.printClientLogs(clientName, reply), "printing client logs")
	case <-time.After(app.Timeout):
		return errors.New("timeout")
	}
	return nil
}

// printClientLogs prints log messages for a client from an OSC message.
func (app *App) printClientLogs(expectedClientName string, msg osc.Message) error {
	const minimumNumArgs = 3

	app.debugf("printing logs from %#v", msg)

	if minimum, got := minimumNumArgs, len(msg.Arguments); got < minimum {
		return errors.Errorf("expected at least %d arguments, got %d", minimum, got)
	}
	addr, err := msg.Arguments[0].ReadString()
	if err != nil {
		return errors.Wrap(err, "reading reply address")
	}
	if expected, got := nsm.AddressClientLogs, addr; expected != got {
		return errors.Errorf("expected %s, got %s", expected, got)
	}
	clientName, err := msg.Arguments[1].ReadString()
	if err != nil {
		return errors.Wrap(err, "reading client name")
	}
	if clientName != expectedClientName {
		return errors.Errorf("expected client name %s, got %s", expectedClientName, clientName)
	}
	numLines, err := msg.Arguments[2].ReadInt32()
	if err != nil {
		return errors.Wrap(err, "reading num log lines")
	}
	if expected, got := numLines, int32(len(msg.Arguments)-minimumNumArgs); expected != got {
		return errors.Errorf("expected %d log lines, got %d", expected, got)
	}
	for i := int32(0); i < numLines; i++ {
		line, err := msg.Arguments[i+minimumNumArgs].ReadString()
		if err != nil {
			return errors.Wrap(err, "reading log line")
		}
		if _, err := fmt.Println(line); err != nil {
			return errors.Wrap(err, "printing log line")
		}
	}
	return nil
}

func init() {
	commandUsage["logs"] = func() error {
		fmt.Fprintf(os.Stderr, "Get the logs of a gonzo client.\n")
		fmt.Fprintf(os.Stderr, "\n")
		fmt.Fprintf(os.Stderr, "Usage:\n")
		fmt.Fprintf(os.Stderr, "gonzoctl logs [OPTIONS] NAME\n")
		fmt.Fprintf(os.Stderr, "\n")
		fmt.Fprintf(os.Stderr, "OPTIONS\n")
		fmt.Fprintf(os.Stderr, "-o stderr|stdout             Show either the client's stderr (default) or stdout.\n")
		fmt.Fprintf(os.Stderr, "\n")
		return nil
	}
}
