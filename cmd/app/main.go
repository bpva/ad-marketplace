package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/bpva/ad-marketplace/internal/config"
	"github.com/bpva/ad-marketplace/internal/http/app"
	"github.com/bpva/ad-marketplace/internal/http/dbg_server"
	bot_service "github.com/bpva/ad-marketplace/internal/service/bot"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer stop()

	cfg, err := config.Load()
	if err != nil {
		slog.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	log := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	log = log.With("env", cfg.Env)

	go dbg_server.Run(cfg.HTTP.PrivatePort, log)

	bot, err := bot_service.New(cfg.Telegram, log)
	if err != nil {
		log.Error("failed to create bot", "error", err)
		os.Exit(1)
	}

	if cfg.Env == "prod" {
		if err := bot.SetWebhook(); err != nil {
			log.Error("failed to set webhook", "error", err)
			os.Exit(1)
		}
	} else {
		// TODO: implement polling for local development
		log.Warn("bot will not receive updates")
	}

	a := app.New(cfg.HTTP, log, bot)

	go func() {
		if err := a.Serve(); err != nil {
			log.Error("server error", "error", err)
			stop()
		}
	}()

	<-ctx.Done()
	log.Info("shutdown signal received")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.HTTP.ShutdownTimeout)
	defer cancel()

	if err := a.Shutdown(shutdownCtx); err != nil {
		log.Error("server shutdown error", "error", err)
	}

	log.Info("server stopped")
}
