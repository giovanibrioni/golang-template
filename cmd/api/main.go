package main

import (
	"context"
	"errors"
	"flag"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/giovani/golang-template/internal/api"
	"github.com/giovani/golang-template/internal/core/config"
	"github.com/giovani/golang-template/internal/core/db"
	"github.com/giovani/golang-template/internal/core/logging"
	"github.com/giovani/golang-template/internal/items"
)

func main() {
	if err := run(); err != nil {
		slog.Error("shutdown", "error", err)
		os.Exit(1)
	}
}

func run() error {
	migrateOnly := flag.Bool("migrate-only", false, "apply migrations and exit")
	flag.Parse()

	cfg := config.Load()
	logging.Setup(cfg.LogLevel, cfg.LogJSON)

	ctx := context.Background()
	pool, err := db.NewPool(ctx, cfg.DatabaseURL)
	if err != nil {
		return err
	}
	defer pool.Close()

	if err := db.Migrate(ctx, pool); err != nil {
		return err
	}
	if *migrateOnly {
		slog.Info("migrations applied")
		return nil
	}

	router := api.NewRouter(pool, items.NewService(items.NewPostgresRepository()))
	server := &http.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
	}

	serveErr := make(chan error, 1)
	go func() {
		slog.Info("listening", "addr", cfg.HTTPAddr)
		serveErr <- server.ListenAndServe()
	}()

	signalCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	select {
	case err := <-serveErr:
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return err
	case <-signalCtx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		return server.Shutdown(shutdownCtx)
	}
}
