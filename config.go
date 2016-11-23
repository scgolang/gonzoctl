package main

import (
	"flag"
	"os"
	"time"

	"github.com/pkg/errors"
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
	config.flags = fs

	defaultTimeout, _ := time.ParseDuration("10s") // Never fails

	fs.StringVar(&config.Host, "host", "127.0.0.1", "Remote host")
	fs.IntVar(&config.Port, "port", 56070, "Remote port")
	fs.DurationVar(&config.Timeout, "timeout", defaultTimeout, "Timeout for replies from gonzo server")
	fs.BoolVar(&config.Debug, "debug", false, "Print debugging information")

	if err := fs.Parse(os.Args[1:]); err != nil {
		return config, errors.Wrap(err, "could not parse config")
	}
	return config, nil
}
