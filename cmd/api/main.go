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

	_ "github.com/giovanibrioni/golang-template/docs"
	"github.com/giovanibrioni/golang-template/internal/api"
	"github.com/giovanibrioni/golang-template/internal/core/config"
	"github.com/giovanibrioni/golang-template/internal/core/db"
	"github.com/giovanibrioni/golang-template/internal/core/logging"
	"github.com/giovanibrioni/golang-template/internal/core/tracing"
	"github.com/giovanibrioni/golang-template/internal/items"
)

// @title Golang Template API
// @version 1.0
// @description A minimal Go HTTP template organized by domain
// @host localhost:8081
// @BasePath /api/v1
// @schemes http
func main() {
	if err := run(); err != nil {
		slog.Error("shutdown", "error", err)
		os.Exit(1)
	}
}

func run() error {
	// migrate-only flag: runs migrations and exits.
	// In production, consider running migrations as a separate init container/job
	// to avoid coupling migration logic with the API binary.
	migrateOnly := flag.Bool("migrate-only", false, "apply migrations and exit")
	flag.Parse()

	cfg := config.Load()
	logging.Setup(cfg.LogLevel, cfg.LogJSON)

	ctx := context.Background()

	// Initialize OpenTelemetry tracing
	shutdownTracing, err := tracing.Init(ctx, "golang-template")
	if err != nil {
		slog.WarnContext(ctx, "tracing init failed", "error", err)
	} else {
		defer shutdownTracing(ctx)
	}

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
