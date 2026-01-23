package bot_service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/bpva/ad-marketplace/internal/config"
	"github.com/bpva/ad-marketplace/internal/logx"
	tele "gopkg.in/telebot.v4"
)

type Bot struct {
	bot     *tele.Bot
	log     *slog.Logger
	baseURL string
}

type noopPoller struct{}

func (p *noopPoller) Poll(b *tele.Bot, updates chan tele.Update, stop chan struct{}) {}

func New(cfg config.Telegram, log *slog.Logger) (*Bot, error) {
	log = log.With(logx.Service("BotService"))

	b, err := tele.NewBot(tele.Settings{
		Token:   cfg.BotToken,
		Poller:  &noopPoller{},
		Offline: true,
	})
	if err != nil {
		return nil, fmt.Errorf("create telebot: %w", err)
	}

	bot := &Bot{
		bot:     b,
		log:     log,
		baseURL: cfg.BaseURL,
	}
	bot.registerHandlers()

	return bot, nil
}

func (b *Bot) registerHandlers() {
	b.bot.Handle("/start", func(c tele.Context) error {
		return c.Send("hey")
	})

	b.bot.Handle(tele.OnText, func(c tele.Context) error {
		return c.Send("confusing...")
	})
}

func (b *Bot) ProcessUpdate(data []byte) error {
	var update tele.Update
	if err := json.Unmarshal(data, &update); err != nil {
		return fmt.Errorf("unmarshal update: %w", err)
	}
	b.bot.ProcessUpdate(update)
	return nil
}

func (b *Bot) Token() string {
	return b.bot.Token
}

func (b *Bot) SetWebhook() error {
	webhookURL := fmt.Sprintf("%s/api/v1/bot/%s/webhook", b.baseURL, b.bot.Token)
	apiURL := fmt.Sprintf("https://api.telegram.org/bot%s/setWebhook", b.bot.Token)

	body, err := json.Marshal(map[string]string{"url": webhookURL})
	if err != nil {
		return fmt.Errorf("marshal request: %w", err)
	}

	resp, err := http.Post(apiURL, "application/json", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("telegram api returned %d", resp.StatusCode)
	}

	b.log.Info("webhook registered")
	return nil
}
