package telebot

import (
	"fmt"

	tele "gopkg.in/telebot.v4"
)

type Client struct {
	bot *tele.Bot
}

type noopPoller struct{}

func (p *noopPoller) Poll(b *tele.Bot, updates chan tele.Update, stop chan struct{}) {}

func New(token string) (*Client, error) {
	b, err := tele.NewBot(tele.Settings{
		Token:   token,
		Poller:  &noopPoller{},
		Offline: true,
	})
	if err != nil {
		return nil, fmt.Errorf("create telebot: %w", err)
	}

	return &Client{bot: b}, nil
}

func (c *Client) Handle(endpoint any, h tele.HandlerFunc) {
	c.bot.Handle(endpoint, h)
}

func (c *Client) ProcessUpdate(upd tele.Update) {
	c.bot.ProcessUpdate(upd)
}

func (c *Client) Token() string {
	return c.bot.Token
}

func (c *Client) AdminsOf(chat *tele.Chat) ([]tele.ChatMember, error) {
	return c.bot.AdminsOf(chat)
}
