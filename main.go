package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/erikfastermann/lam/server"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	config, err := server.ConfigFromEnv("LAM")
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	logger := log.New(os.Stderr, "", log.LstdFlags)
	return server.ListenAndServe(ctx, config, logger)
}
