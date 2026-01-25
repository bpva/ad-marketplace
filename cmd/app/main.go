package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/bpva/ad-marketplace/internal/config"
	"github.com/bpva/ad-marketplace/internal/gateway/telebot"
	"github.com/bpva/ad-marketplace/internal/http/app"
	"github.com/bpva/ad-marketplace/internal/http/dbg_server"
	"github.com/bpva/ad-marketplace/internal/migrations"
	channel_repo "github.com/bpva/ad-marketplace/internal/repository/channel"
	user_repo "github.com/bpva/ad-marketplace/internal/repository/user"
	"github.com/bpva/ad-marketplace/internal/service/auth"
	bot_service "github.com/bpva/ad-marketplace/internal/service/bot"
	"github.com/bpva/ad-marketplace/internal/storage"
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

	if err := migrations.Run(storage.URL(cfg.Postgres)); err != nil {
		log.Error("failed to run migrations", "error", err)
		os.Exit(1)
	}
	log.Info("migrations completed")

	db, err := storage.New(ctx, cfg.Postgres)
	if err != nil {
		log.Error("failed to create storage", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	userRepo := user_repo.New(db)
	channelRepo := channel_repo.New(db)
	authSvc := auth.New(userRepo, cfg.Telegram.BotToken, cfg.JWT.Secret, log)

	go dbg_server.Run(cfg.HTTP.PrivatePort, log)

	telebotClient, err := telebot.New(cfg.Telegram.BotToken)
	if err != nil {
		log.Error("failed to create telebot client", "error", err)
		os.Exit(1)
	}

	bot := bot_service.New(telebotClient, cfg.Telegram.BaseURL, log, db, channelRepo, userRepo)

	if cfg.Env == "prod" {
		if err := bot.SetWebhook(); err != nil {
			log.Error("failed to set webhook", "error", err)
			os.Exit(1)
		}
	} else {
		log.Warn("bot will not receive updates")
	}

	a := app.New(cfg.HTTP, log, bot, authSvc)

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
