package bot_service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/google/uuid"

	"github.com/bpva/ad-marketplace/internal/entity"
	"github.com/bpva/ad-marketplace/internal/logx"
	"github.com/bpva/ad-marketplace/internal/storage"
	tele "gopkg.in/telebot.v4"
)

//go:generate mockgen -destination=mocks.go -package=bot_service . TelebotClient
type TelebotClient interface {
	Handle(endpoint any, h tele.HandlerFunc)
	ProcessUpdate(upd tele.Update)
	Token() string
	AdminsOf(chat *tele.Chat) ([]tele.ChatMember, error)
}

type ChannelRepository interface {
	Create(
		ctx context.Context,
		telegramChannelID int64,
		title string,
		username *string,
	) (*entity.Channel, error)
	CreateRole(
		ctx context.Context,
		channelID, userID uuid.UUID,
		role string,
	) (*entity.ChannelRole, error)
	SoftDelete(ctx context.Context, telegramChannelID int64) error
}

type UserRepository interface {
	GetByTelegramID(ctx context.Context, telegramID int64) (*entity.User, error)
	Create(ctx context.Context, telegramID int64, name string) (*entity.User, error)
}

type svc struct {
	client      TelebotClient
	log         *slog.Logger
	baseURL     string
	tx          storage.Transactor
	channelRepo ChannelRepository
	userRepo    UserRepository
}

func New(
	client TelebotClient,
	baseURL string,
	log *slog.Logger,
	tx storage.Transactor,
	channels ChannelRepository,
	users UserRepository,
) *svc {
	log = log.With(logx.Service("BotService"))

	s := &svc{
		client:      client,
		log:         log,
		baseURL:     baseURL,
		tx:          tx,
		channelRepo: channels,
		userRepo:    users,
	}

	s.registerHandlers()

	return s
}

func (b *svc) registerHandlers() {
	b.client.Handle("/start", func(c tele.Context) error {
		return c.Send("hey")
	})

	b.client.Handle(tele.OnText, func(c tele.Context) error {
		return c.Send("confusing...")
	})

	b.client.Handle(tele.OnMyChatMember, b.handleMyChatMember)
}

func (b *svc) ProcessUpdate(data []byte) error {
	var update tele.Update
	if err := json.Unmarshal(data, &update); err != nil {
		return fmt.Errorf("unmarshal update: %w", err)
	}
	b.client.ProcessUpdate(update)
	return nil
}

func (b *svc) Token() string {
	return b.client.Token()
}

func (b *svc) SetWebhook() error {
	webhookURL := fmt.Sprintf("%s/api/v1/bot/%s/webhook", b.baseURL, b.client.Token())
	apiURL := fmt.Sprintf("https://api.telegram.org/bot%s/setWebhook", b.client.Token())

	body, err := json.Marshal(map[string]any{
		"url":             webhookURL,
		"allowed_updates": []string{"message", "my_chat_member"},
	})
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
