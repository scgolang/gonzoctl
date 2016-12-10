package main

import (
	"context"
	"log"
)

func main() {
	config, err := NewConfig()
	if err != nil {
		log.Fatal(err)
	}

	app, err := NewApp(context.Background(), config)
	if err != nil {
		log.Fatal(err)
	}

	if err := app.Run(); err != nil {
		if err != context.Canceled && err != context.DeadlineExceeded {
			log.Fatal(err)
		}
	}
}
