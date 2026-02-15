package mtproto

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/gotd/td/session"
	"github.com/gotd/td/telegram"
	"github.com/gotd/td/tg"

	"github.com/bpva/ad-marketplace/internal/config"
	"github.com/bpva/ad-marketplace/internal/dto"
	"github.com/bpva/ad-marketplace/internal/logx"
)

type gateway struct {
	client     *telegram.Client
	api        *tg.Client
	userClient *telegram.Client
	userAPI    *tg.Client
	log        *slog.Logger
}

func New(ctx context.Context, cfg config.Telegram, log *slog.Logger) (*gateway, error) {
	if cfg.APIId == 0 || cfg.APIHash == "" {
		return nil, fmt.Errorf("TG_API_ID or TG_API_HASH not set")
	}

	log = log.With(logx.Service("mtproto"))
	client := telegram.NewClient(cfg.APIId, cfg.APIHash, telegram.Options{})
	userClient := telegram.NewClient(cfg.APIId, cfg.APIHash, telegram.Options{
		SessionStorage: &session.FileStorage{Path: ".session.json"},
	})

	c := &gateway{
		client:     client,
		userClient: userClient,
		log:        log,
	}

	botReady := make(chan struct{})
	botErrCh := make(chan error, 1)
	go func() {
		botErrCh <- client.Run(ctx, func(ctx context.Context) error {
			status, err := client.Auth().Status(ctx)
			if err != nil {
				return fmt.Errorf("auth status: %w", err)
			}

			if !status.Authorized {
				if _, err := client.Auth().Bot(ctx, cfg.BotToken); err != nil {
					return fmt.Errorf("bot auth: %w", err)
				}
			}

			c.api = client.API()
			close(botReady)

			<-ctx.Done()
			return ctx.Err()
		})
	}()

	userReady := make(chan struct{})
	userErrCh := make(chan error, 1)
	go func() {
		userErrCh <- userClient.Run(ctx, func(ctx context.Context) error {
			status, err := userClient.Auth().Status(ctx)
			if err != nil {
				return fmt.Errorf("user auth status: %w", err)
			}
			if !status.Authorized {
				return fmt.Errorf("user session is not authorized in .session.json")
			}

			c.userAPI = userClient.API()
			close(userReady)

			<-ctx.Done()
			return ctx.Err()
		})
	}()

	botConnected := false
	userConnected := false
	for !botConnected || !userConnected {
		select {
		case <-botReady:
			botConnected = true
		case <-userReady:
			userConnected = true
		case err := <-botErrCh:
			return nil, err
		case err := <-userErrCh:
			return nil, err
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}

	c.log.Info("connected")
	return c, nil
}

func (c *gateway) Ping(ctx context.Context) error {
	_, err := c.client.API().HelpGetConfig(ctx)
	return err
}

func BotAPIToMTProto(botAPIID int64) int64 {
	if botAPIID < -1_000_000_000_000 {
		return -botAPIID - 1_000_000_000_000
	}
	return botAPIID
}

func (c *gateway) resolveChannel(
	ctx context.Context,
	api *tg.Client,
	channelID int64,
) (int64, error) {
	res, err := api.ChannelsGetChannels(ctx, []tg.InputChannelClass{
		&tg.InputChannel{ChannelID: channelID, AccessHash: 0},
	})
	if err != nil {
		return 0, fmt.Errorf("resolve channel %d: %w", channelID, err)
	}

	chats, ok := res.(*tg.MessagesChats)
	if !ok {
		return 0, fmt.Errorf("unexpected response type: %T", res)
	}

	for _, ch := range chats.Chats {
		if channel, ok := ch.(*tg.Channel); ok && channel.ID == channelID {
			return channel.AccessHash, nil
		}
	}
	return 0, fmt.Errorf("channel %d not found", channelID)
}

func (c *gateway) GetChannelFull(
	ctx context.Context,
	channelID int64,
) (*dto.ChannelFullInfo, error) {
	accessHash, err := c.resolveChannel(ctx, c.api, channelID)
	if err != nil {
		return nil, err
	}

	full, err := c.api.ChannelsGetFullChannel(ctx, &tg.InputChannel{
		ChannelID:  channelID,
		AccessHash: accessHash,
	})
	if err != nil {
		return nil, fmt.Errorf("get full channel: %w", err)
	}

	channelFull, ok := full.FullChat.(*tg.ChannelFull)
	if !ok {
		return nil, fmt.Errorf("unexpected chat type: %T", full.FullChat)
	}

	info := &dto.ChannelFullInfo{
		About:        channelFull.GetAbout(),
		CanViewStats: channelFull.CanViewStats,
	}

	if v, ok := channelFull.GetParticipantsCount(); ok {
		info.ParticipantsCount = v
	}
	if v, ok := channelFull.GetLinkedChatID(); ok {
		info.LinkedChatID = v
	}
	if v, ok := channelFull.GetAdminsCount(); ok {
		info.AdminsCount = v
	}
	if v, ok := channelFull.GetOnlineCount(); ok {
		info.OnlineCount = v
	}
	if v, ok := channelFull.GetStatsDC(); ok {
		info.StatsDC = v
	}

	return info, nil
}
