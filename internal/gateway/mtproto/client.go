package mtproto

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/gotd/td/telegram"
	"github.com/gotd/td/tg"

	"github.com/bpva/ad-marketplace/internal/config"
	"github.com/bpva/ad-marketplace/internal/dto"
	"github.com/bpva/ad-marketplace/internal/logx"
)

type Client struct {
	client *telegram.Client
	api    *tg.Client
	log    *slog.Logger
}

func New(ctx context.Context, cfg config.Telegram, log *slog.Logger) (*Client, error) {
	if cfg.APIId == 0 || cfg.APIHash == "" {
		return nil, fmt.Errorf("TG_API_ID or TG_API_HASH not set")
	}

	log = log.With(logx.Service("mtproto"))
	client := telegram.NewClient(cfg.APIId, cfg.APIHash, telegram.Options{})

	c := &Client{
		client: client,
		log:    log,
	}

	ready := make(chan struct{})
	errCh := make(chan error, 1)
	go func() {
		errCh <- client.Run(ctx, func(ctx context.Context) error {
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
			close(ready)

			<-ctx.Done()
			return ctx.Err()
		})
	}()

	select {
	case <-ready:
		c.log.Info("connected")
		return c, nil
	case err := <-errCh:
		return nil, err
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

func (c *Client) Ping(ctx context.Context) error {
	_, err := c.client.API().HelpGetConfig(ctx)
	return err
}

func (c *Client) GetChannelFull(
	ctx context.Context,
	channelID int64,
) (*dto.ChannelFullInfo, error) {
	full, err := c.api.ChannelsGetFullChannel(ctx, &tg.InputChannel{ChannelID: channelID})
	if err != nil {
		return nil, fmt.Errorf("get full channel: %w", err)
	}

	channelFull, ok := full.FullChat.(*tg.ChannelFull)
	if !ok {
		return nil, fmt.Errorf("unexpected chat type: %T", full.FullChat)
	}

	info := &dto.ChannelFullInfo{
		ParticipantsCount: channelFull.ParticipantsCount,
	}

	if channelFull.LinkedChatID != 0 {
		info.LinkedChatID = channelFull.LinkedChatID
	}

	return info, nil
}

func (c *Client) GetBroadcastStats(
	ctx context.Context,
	channelID int64,
) (*dto.BroadcastStats, error) {
	stats, err := c.api.StatsGetBroadcastStats(ctx, &tg.StatsGetBroadcastStatsRequest{
		Channel: &tg.InputChannel{ChannelID: channelID},
	})
	if err != nil {
		return nil, fmt.Errorf("get broadcast stats: %w", err)
	}

	return &dto.BroadcastStats{
		Period: dto.StatsPeriod{
			MinDate: stats.Period.MinDate,
			MaxDate: stats.Period.MaxDate,
		},
		Followers: dto.StatsValue{
			Current:  stats.Followers.Current,
			Previous: stats.Followers.Previous,
		},
		ViewsPerPost: dto.StatsValue{
			Current:  stats.ViewsPerPost.Current,
			Previous: stats.ViewsPerPost.Previous,
		},
		SharesPerPost: dto.StatsValue{
			Current:  stats.SharesPerPost.Current,
			Previous: stats.SharesPerPost.Previous,
		},
		ReactionsPerPost: dto.StatsValue{
			Current:  stats.ReactionsPerPost.Current,
			Previous: stats.ReactionsPerPost.Previous,
		},
		EnabledNotifications: dto.StatsPercentage{
			Part:  stats.EnabledNotifications.Part,
			Total: stats.EnabledNotifications.Total,
		},
	}, nil
}
