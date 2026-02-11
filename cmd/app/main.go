package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/bpva/ad-marketplace/internal/config"
	"github.com/bpva/ad-marketplace/internal/gateway/mtproto"
	"github.com/bpva/ad-marketplace/internal/gateway/telebot"
	"github.com/bpva/ad-marketplace/internal/http/app"
	"github.com/bpva/ad-marketplace/internal/http/dbgserver"
	"github.com/bpva/ad-marketplace/internal/logx"
	channel_repo "github.com/bpva/ad-marketplace/internal/repository/channel"
	settings_repo "github.com/bpva/ad-marketplace/internal/repository/settings"
	user_repo "github.com/bpva/ad-marketplace/internal/repository/user"
	"github.com/bpva/ad-marketplace/internal/service/auth"
	"github.com/bpva/ad-marketplace/internal/service/bot"
	channel_service "github.com/bpva/ad-marketplace/internal/service/channel"
	"github.com/bpva/ad-marketplace/internal/service/stats"
	user_service "github.com/bpva/ad-marketplace/internal/service/user"
	"github.com/bpva/ad-marketplace/internal/storage"
	"github.com/bpva/ad-marketplace/migrations"
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
	settingsRepo := settings_repo.New(db)
	authSvc := auth.New(userRepo, cfg.Telegram.BotToken, cfg.JWT.Secret, log)

	go dbgserver.Run(cfg.HTTP.PrivatePort, log)

	telebotClient, err := telebot.New(cfg.Telegram.BotToken)
	if err != nil {
		log.Error("failed to create telebot client", "error", err)
		os.Exit(1)
	}

	mtprotoClient, err := mtproto.New(ctx, cfg.Telegram, log)
	if err != nil {
		log.Error("failed to create mtproto client", "error", err)
		os.Exit(1)
	}

	statsSvc := stats.New(mtprotoClient, channelRepo, log)

	botSvc := bot.New(telebotClient, cfg.Telegram, log, db, channelRepo, userRepo, statsSvc)

	if cfg.Env == "prod" {
		if err := botSvc.SetWebhook(); err != nil {
			log.Error("failed to set webhook", "error", err)
			os.Exit(1)
		}
	} else {
		log.Warn("bot will not receive updates")
	}

	channelSvc := channel_service.New(channelRepo, userRepo, telebotClient, db, log)
	userSvc := user_service.New(userRepo, settingsRepo, log)

	a := app.New(cfg.HTTP, log, botSvc, authSvc, channelSvc, userSvc, statsSvc)

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

	if err := logShutdown(shutdownCtx); err != nil {
		log.Error("logger shutdown error", "error", err)
	}
}
