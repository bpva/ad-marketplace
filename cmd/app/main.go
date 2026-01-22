package main

import (
	"log/slog"
	"os"

	"github.com/bpva/ad-marketplace/internal/config"
	"github.com/bpva/ad-marketplace/internal/http/app"
	"github.com/bpva/ad-marketplace/internal/http/dbg_server"
	bot_service "github.com/bpva/ad-marketplace/internal/service/bot"
)

func main() {
	log := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	cfg, err := config.Load()
	if err != nil {
		log.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	go dbg_server.Run(cfg.HTTP.PrivatePort, log)

	bot, err := bot_service.New(cfg.Telegram.BotToken, log)
	if err != nil {
		log.Error("failed to create bot", "error", err)
		os.Exit(1)
	}

	a := app.New(cfg.HTTP, log, bot)

	if err := a.Serve(); err != nil {
		log.Error("server error", "error", err)
		os.Exit(1)
	}
}
