package main

import (
	"flag"
	"os"

	"github.com/pkg/errors"
)

// Config holds the application's configuration.
type Config struct {
	Host string
	Port int

	flags *flag.FlagSet
}

// NewConfig parses the application's config from command line arguments.
func NewConfig() (Config, error) {
	var (
		config = Config{}
		fs     = flag.NewFlagSet("gonzoctl", flag.ExitOnError)
	)
	config.flags = fs

	fs.StringVar(&config.Host, "host", "127.0.0.1", "Remote host")
	fs.IntVar(&config.Port, "port", 56070, "Remote port")

	if err := fs.Parse(os.Args[1:]); err != nil {
		return config, errors.Wrap(err, "could not parse config")
	}
	return config, nil
}
