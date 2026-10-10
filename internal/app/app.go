package app

import (
	"context"
	"fmt"
	server "ledger/internal/adapters/http"
	"ledger/internal/adapters/postgres"
	"ledger/internal/application"
	"ledger/internal/config"
	"log"
)

func Run(ctx context.Context, cfg *config.Config) error {
	pool, err := postgres.NewPool(ctx, cfg.DatabaseURL)
	if err != nil {
		return err
	}
	defer pool.Close()

	fmt.Println("Connected to database")

	accountRepo := postgres.NewAccountRepository(pool)
	accountService := application.NewAccountService(accountRepo)
	srv := server.NewServer(cfg, accountService)

	if err := srv.Run(ctx); err != nil {
		return err
	}

	log.Println("Server stopped")

	return nil
}
