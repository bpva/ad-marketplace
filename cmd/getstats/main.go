package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/bpva/ad-marketplace/internal/config"
	"github.com/bpva/ad-marketplace/internal/gateway/mtproto"
	"github.com/bpva/ad-marketplace/internal/logx"
	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load()
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer stop()

	cfg, err := config.Load()
	if err != nil {
		slog.Error("failed to load config", "error", err)
		os.Exit(1) //nolint:gocritic
	}

	log, _, err := logx.NewLogger(ctx, cfg.Logger, cfg.Env)
	if err != nil {
		slog.Error("failed to initialize logger", "error", err)
		os.Exit(1)
	}

	cl, err := mtproto.New(ctx, cfg.Telegram, log)
	if err != nil {
		log.Error("failed to create mtproto client", "error", err)
		os.Exit(1)
	}

	id := mtproto.BotAPIToMTProto(-1001165372896)

	info, err := cl.GetChannelFull(ctx, id)
	if err != nil {
		log.Error("failed to get channel info", "error", err)
		os.Exit(1)
	}

	all, err := cl.GetBroadcastStats(ctx, id, 0)
	if err != nil {
		log.Error("failed to get broadcast stats", "error", err)
		os.Exit(1)
	}

	log.Info("broadcast stats", "data", all)

	log.Info("channel info", "data", info)
}
