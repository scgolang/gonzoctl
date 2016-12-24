package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/pkg/errors"
)

const (
	// DefaultPort is the default gonzo port.
	DefaultPort = 56070
)

// Config holds the application's configuration.
type Config struct {
	Host    string        `json:"host"`
	Port    int           `json:"port"`
	Timeout time.Duration `json:"timeout"`
	Debug   bool          `json:"debug"`

	flags *flag.FlagSet
}

// NewConfig parses the application's config from command line arguments.
func NewConfig() (Config, error) {
	var (
		config = Config{}
		fs     = flag.NewFlagSet("gonzoctl", flag.ExitOnError)
	)
	fs.Usage = usage
	config.flags = fs

	defaultTimeout, _ := time.ParseDuration("10s") // Never fails

	fs.StringVar(&config.Host, "host", "127.0.0.1", "Remote host")
	fs.IntVar(&config.Port, "port", DefaultPort, "Remote port")
	fs.DurationVar(&config.Timeout, "timeout", defaultTimeout, "Timeout for replies from gonzo server")
	fs.BoolVar(&config.Debug, "debug", false, "Print debugging information")

	if err := fs.Parse(os.Args[1:]); err != nil {
		return config, errors.Wrap(err, "could not parse config")
	}
	return config, nil
}

// usage prints a usage message to stderr.
func usage() {
	fmt.Fprintf(os.Stderr, "Usage:\n")
	fmt.Fprintf(os.Stderr, "gonzoctl [GLOBAL_OPTIONS] COMMAND [COMMAND_OPTIONS]\n")
	fmt.Fprintf(os.Stderr, "\n")
	fmt.Fprintf(os.Stderr, "Global Options:\n")
	fmt.Fprintf(os.Stderr, "-host HOST              Host or IP of a gonzo server (default is 127.0.0.1).\n")
	fmt.Fprintf(os.Stderr, "-port PORT              Listening port of a gonzo server (default is 56070).\n")
	fmt.Fprintf(os.Stderr, "-timeout DURATION       Timeout used when waiting for replies from a gonzo server (default is 10s).\n")
	fmt.Fprintf(os.Stderr, "-debug                  Enable debug logging (default is false).\n")
	fmt.Fprintf(os.Stderr, "\n")
	fmt.Fprintf(os.Stderr, "Commands:\n")
	fmt.Fprintf(os.Stderr, "add             Add a client to the current session.\n")
	fmt.Fprintf(os.Stderr, "help            Print this usage message.\n")
	fmt.Fprintf(os.Stderr, "lc              List clients for the current session.\n")
	fmt.Fprintf(os.Stderr, "ls              List sessions.\n")
	fmt.Fprintf(os.Stderr, "ping            Ping a gonzo server.\n")
	fmt.Fprintf(os.Stderr, "\n")
	fmt.Fprintf(os.Stderr, "To see usage of a single command do:\n")
	fmt.Fprintf(os.Stderr, "\n")
	fmt.Fprintf(os.Stderr, "gonzoctl help COMMAND\n")
	fmt.Fprintf(os.Stderr, "\n")
}

var commandUsage = map[string]func() error{}

// usageCmd is for the "help" command.
func usageCmd(args []string) error {
	switch len(args) {
	case 1:
		// gonzoctl help COMMAND
		var (
			cmd = args[0]
			cu  = commandUsage[cmd]
		)
		if cu != nil {
			return cu()
		}
		return errors.New("unrecognized command: " + cmd)
	default:
		usage()
	}
	return nil
}
