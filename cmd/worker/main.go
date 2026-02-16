package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/bpva/ad-marketplace/internal/config"
	"github.com/bpva/ad-marketplace/internal/logx"
	"github.com/bpva/ad-marketplace/internal/storage"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer stop()

	cfg, err := config.Load()
	if err != nil {
		slog.Error("failed to load config", "error", err)
		os.Exit(1) //nolint:gocritic
	}

	log, logShutdown, err := logx.NewLogger(ctx, cfg.Logger, cfg.Env)
	if err != nil {
		slog.Error("failed to initialize logger", "error", err)
		os.Exit(1)
	}

	db, err := storage.New(ctx, cfg.Postgres)
	if err != nil {
		log.Error("failed to create storage", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	log.Info("worker started")

	<-ctx.Done()
	log.Info("shutdown signal received")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.HTTP.ShutdownTimeout)
	defer cancel()

	if err := logShutdown(shutdownCtx); err != nil {
		log.Error("logger shutdown error", "error", err)
	}
}
