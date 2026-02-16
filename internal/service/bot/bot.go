package bot

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"runtime/debug"
	"sync"
	"time"

	"github.com/google/uuid"
	tele "gopkg.in/telebot.v4"

	"github.com/bpva/ad-marketplace/internal/config"
	"github.com/bpva/ad-marketplace/internal/dto"
	"github.com/bpva/ad-marketplace/internal/entity"
	"github.com/bpva/ad-marketplace/internal/logx"
	"github.com/bpva/ad-marketplace/internal/storage"
)

//go:generate mockgen -destination=mocks.go -package=bot . TelebotClient
type TelebotClient interface {
	Handle(endpoint any, h tele.HandlerFunc)
	ProcessUpdate(upd tele.Update)
	Token() string
	AdminsOf(channelID int64) ([]dto.ChannelAdmin, error)
	GetChatPhoto(chatID int64) (smallFileID, bigFileID string, err error)
	DownloadFile(fileID string) ([]byte, error)
	Send(to tele.Recipient, what any, opts ...any) (*tele.Message, error)
	SendAlbum(to tele.Recipient, a tele.Album, opts ...any) ([]tele.Message, error)
}

type ChannelRepository interface {
	Create(
		ctx context.Context,
		TgChannelID int64,
		title string,
		username *string,
	) (*entity.Channel, error)
	CreateRole(
		ctx context.Context,
		channelID, userID uuid.UUID,
		role entity.ChannelRoleType,
	) (*entity.ChannelRole, error)
	SoftDelete(ctx context.Context, TgChannelID int64) error
	UpdatePhoto(ctx context.Context, channelID uuid.UUID, smallFileID, bigFileID *string) error
}

type UserRepository interface {
	GetByTgID(ctx context.Context, tgID int64) (*entity.User, error)
	Create(ctx context.Context, tgID int64, name string) (*entity.User, error)
}

type StatsFetcher interface {
	FetchAndStore(ctx context.Context, channelID uuid.UUID, tgChannelID int64) error
}

type PostRepository interface {
	Create(
		ctx context.Context,
		userID uuid.UUID,
		name *string,
		mediaGroupID *string,
		text *string,
		entities []byte,
		mediaType *entity.MediaType,
		mediaFileID *string,
		hasMediaSpoiler bool,
		showCaptionAboveMedia bool,
	) (*entity.Post, error)
}

type svc struct {
	client        TelebotClient
	log           *slog.Logger
	cfg           config.Telegram
	tx            storage.Transactor
	channelRepo   ChannelRepository
	userRepo      UserRepository
	stats         StatsFetcher
	postRepo      PostRepository
	awaitingPost  sync.Map
	pendingGroups sync.Map
}

func New(
	client TelebotClient,
	cfg config.Telegram,
	log *slog.Logger,
	tx storage.Transactor,
	channels ChannelRepository,
	users UserRepository,
	stats StatsFetcher,
	posts PostRepository,
) *svc {
	log = log.With(logx.Service("BotService"))

	s := &svc{
		client:      client,
		log:         log,
		cfg:         cfg,
		tx:          tx,
		channelRepo: channels,
		userRepo:    users,
		stats:       stats,
		postRepo:    posts,
	}

	s.registerHandlers()

	return s
}

func (b *svc) handle(endpoint any, h tele.HandlerFunc) {
	handler := endpointName(endpoint)
	b.client.Handle(endpoint, func(c tele.Context) (err error) {
		start := time.Now()
		log := b.log.With("update_id", c.Update().ID, "handler", handler)

		log.Info("webhook received", updateAttrs(c)...)

		defer func() {
			duration := time.Since(start)
			if r := recover(); r != nil {
				log.Error("bot handler panic",
					"panic", r,
					"duration", duration,
					"stack", string(debug.Stack()))
				err = fmt.Errorf("panic: %v", r)
				return
			}
			if err != nil {
				log.Error("webhook processing failed",
					"duration", duration,
					"error", err)
			} else {
				log.Info("webhook processed", "duration", duration)
			}
		}()
		return h(c)
	})
}

func endpointName(endpoint any) string {
	s := fmt.Sprint(endpoint)
	if len(s) > 0 && s[0] == '\a' {
		return s[1:]
	}
	return s
}

func updateAttrs(c tele.Context) []any {
	var attrs []any
	if sender := c.Sender(); sender != nil {
		attrs = append(attrs, "sender_id", sender.ID)
	}
	if chat := c.Chat(); chat != nil {
		attrs = append(attrs, "chat_id", chat.ID, "chat_type", string(chat.Type))
	}
	return attrs
}

func (b *svc) registerHandlers() {
	b.handle("/start", func(c tele.Context) error {
		menu := &tele.ReplyMarkup{}
		btn := menu.WebApp("Launch Marketplace", &tele.WebApp{URL: b.cfg.MiniAppURL})
		menu.Inline(menu.Row(btn))
		return c.Send("Welcome to ADxCHANGE!", menu)
	})

	b.handle("/add_promo", b.handleAddPromo)
	b.handle(tele.OnText, b.handleIncomingMessage)
	b.handle(tele.OnPhoto, b.handleIncomingMessage)
	b.handle(tele.OnVideo, b.handleIncomingMessage)
	b.handle(tele.OnDocument, b.handleIncomingMessage)
	b.handle(tele.OnAnimation, b.handleIncomingMessage)
	b.handle(tele.OnAudio, b.handleIncomingMessage)
	b.handle(tele.OnVoice, b.handleIncomingMessage)
	b.handle(tele.OnVideoNote, b.handleIncomingMessage)
	b.handle(tele.OnSticker, b.handleIncomingMessage)
	b.handle(tele.OnMyChatMember, b.handleMyChatMember)
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
	webhookURL := fmt.Sprintf("%s/api/v1/bot/%s/webhook", b.cfg.BaseURL, b.client.Token())
	apiURL := fmt.Sprintf("https://api.telegram.org/bot%s/setWebhook", b.client.Token())

	body, err := json.Marshal(map[string]any{
		"url":             webhookURL,
		"allowed_updates": []string{"message", "my_chat_member"},
	})
	if err != nil {
		return fmt.Errorf("marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(
		context.Background(), http.MethodPost, apiURL, bytes.NewReader(body),
	)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("send request: %w", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			b.log.Error("failed to close response body", "error", err)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("telegram api returned %d", resp.StatusCode)
	}

	b.log.Info("webhook registered")
	return nil
}
