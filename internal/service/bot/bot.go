package bot_service

import (
	"encoding/json"
	"fmt"
	"log/slog"

	tele "gopkg.in/telebot.v4"
)

type Bot struct {
	bot *tele.Bot
	log *slog.Logger
}

type noopPoller struct{}

func (p *noopPoller) Poll(b *tele.Bot, updates chan tele.Update, stop chan struct{}) {}

func New(token string, log *slog.Logger) (*Bot, error) {
	b, err := tele.NewBot(tele.Settings{
		Token:   token,
		Poller:  &noopPoller{},
		Offline: true,
	})
	if err != nil {
		return nil, fmt.Errorf("create telebot: %w", err)
	}

	bot := &Bot{bot: b, log: log}
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
