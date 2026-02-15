package telebot

import (
	"fmt"
	"io"
	"log/slog"

	tele "gopkg.in/telebot.v4"

	"github.com/bpva/ad-marketplace/internal/dto"
	"github.com/bpva/ad-marketplace/internal/logx"
)

type Client struct {
	bot *tele.Bot
}

type noopPoller struct{}

func (p *noopPoller) Poll(b *tele.Bot, updates chan tele.Update, stop chan struct{}) {}

func New(token string, log *slog.Logger) (*Client, error) {
	log = log.With(logx.Service("Telebot"))

	b, err := tele.NewBot(tele.Settings{
		Token:   token,
		Poller:  &noopPoller{},
		Offline: true,
		OnError: func(err error, c tele.Context) {
			if c != nil {
				log.Error("handler error",
					"update_id", c.Update().ID,
					"error", err,
				)
			} else {
				log.Error("handler error", "error", err)
			}
		},
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

func (c *Client) GetChatPhoto(chatID int64) (smallFileID, bigFileID string, err error) {
	chat, err := c.bot.ChatByID(chatID)
	if err != nil {
		return "", "", fmt.Errorf("get chat: %w", err)
	}
	if chat.Photo == nil {
		return "", "", nil
	}
	return chat.Photo.SmallFileID, chat.Photo.BigFileID, nil
}

func (c *Client) DownloadFile(fileID string) ([]byte, error) {
	f := tele.File{FileID: fileID}
	rc, err := c.bot.File(&f)
	if err != nil {
		return nil, fmt.Errorf("download file: %w", err)
	}
	defer func() { _ = rc.Close() }()
	return io.ReadAll(rc)
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
