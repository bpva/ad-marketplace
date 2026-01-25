package telebot

import (
	"fmt"

	tele "gopkg.in/telebot.v4"

	"github.com/bpva/ad-marketplace/internal/dto"
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

func (c *Client) AdminsOf(channelID int64) ([]dto.ChannelAdmin, error) {
	members, err := c.bot.AdminsOf(&tele.Chat{ID: channelID})
	if err != nil {
		return nil, err
	}

	admins := make([]dto.ChannelAdmin, 0, len(members))
	for _, m := range members {
		if m.User == nil {
			continue
		}
		admins = append(admins, dto.ChannelAdmin{
			TgID:      m.User.ID,
			FirstName: m.User.FirstName,
			LastName:  m.User.LastName,
			Username:  m.User.Username,
			Role:      string(m.Role),
		})
	}

	return admins, nil
}
