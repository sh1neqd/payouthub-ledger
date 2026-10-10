package server

import (
	"context"
	"errors"
	"ledger/internal/config"
	"log"
	"net"
	"time"

	"ledger/internal/application"

	"github.com/gofiber/fiber/v3"
)

type Server struct {
	cfg *config.Config
	app *fiber.App
}

func NewServer(cfg *config.Config, accountService *application.AccountService) *Server {
	app := fiber.New(fiber.Config{
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	})

	app.Get("/healthCheck", func(c fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status": "ok",
		})
	})

	app.Post("/accounts", createAccountHandler(accountService))
	app.Get("/accounts/:id", getAccountByIDHandler(accountService))

	return &Server{
		cfg: cfg,
		app: app,
	}
}

func (s *Server) Run(ctx context.Context) error {
	if ctx.Err() != nil {
		return nil
	}

	listener, err := net.Listen("tcp", s.cfg.HTTPAddr)
	if err != nil {
		return err
	}
	defer listener.Close()

	errCh := make(chan error, 1)

	go func() {
		errCh <- s.app.Listener(listener, fiber.ListenConfig{
			DisableStartupMessage: true,
		})
	}()

	log.Printf("server listening on %s", listener.Addr())

	var serveErr error

	select {
	case serveErr = <-errCh:
		errCh = nil
	case <-ctx.Done():
	}

	closeErr := listener.Close()
	if errors.Is(closeErr, net.ErrClosed) {
		closeErr = nil
	}

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	shutdownErr := s.app.ShutdownWithContext(shutdownCtx)

	if errors.Is(shutdownErr, net.ErrClosed) {
		shutdownErr = nil
	}

	if errCh != nil {
		serveErr = <-errCh
	}

	return errors.Join(serveErr, shutdownErr, closeErr)

}
