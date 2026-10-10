package main

import (
	"context"
	"ledger/internal/app"
	"ledger/internal/config"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	if err := runApp(); err != nil {
		log.Printf("ledger: %v", err)
		os.Exit(1)
	}
}

func runApp() error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	cfg, err := config.Load()

	if err != nil {
		return err
	}

	return app.Run(ctx, cfg)

}
