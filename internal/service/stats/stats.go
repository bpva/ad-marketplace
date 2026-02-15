package stats

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/google/uuid"

	"github.com/bpva/ad-marketplace/internal/dto"
	"github.com/bpva/ad-marketplace/internal/entity"
	"github.com/bpva/ad-marketplace/internal/gateway/mtproto"
	"github.com/bpva/ad-marketplace/internal/logx"
)

type ChannelRepository interface {
	GetByTgChannelID(ctx context.Context, tgChannelID int64) (*entity.Channel, error)
	UpsertInfo(ctx context.Context, info *entity.ChannelInfo) error
	BatchUpsertHistoricalStats(
		ctx context.Context,
		channelID uuid.UUID,
		stats []entity.DailyMetrics,
	) error
	RefreshMV(ctx context.Context) error
}

//go:generate mockgen -destination=mocks.go -package=stats . MTProtoClient
type MTProtoClient interface {
	GetChannelFull(ctx context.Context, channelID int64) (*dto.ChannelFullInfo, error)
	GetBroadcastStats(
		ctx context.Context,
		channelID int64,
		statsDC int,
	) (*entity.BroadcastStats, error)
}

type svc struct {
	mtproto     MTProtoClient
	channelRepo ChannelRepository
	log         *slog.Logger
}

func New(
	mtprotoClient MTProtoClient,
	channelRepo ChannelRepository,
	log *slog.Logger,
) *svc {
	log = log.With(logx.Service("StatsService"))
	return &svc{
		mtproto:     mtprotoClient,
		channelRepo: channelRepo,
		log:         log,
	}
}

func (s *svc) FetchAndStore(ctx context.Context, channelID uuid.UUID, tgChannelID int64) error {
	mtprotoID := mtproto.BotAPIToMTProto(tgChannelID)

	fullInfo, err := s.mtproto.GetChannelFull(ctx, mtprotoID)
	if err != nil {
		return fmt.Errorf("get channel full: %w", err)
	}

	info := &entity.ChannelInfo{
		ChannelID:   channelID,
		About:       fullInfo.About,
		Subscribers: fullInfo.ParticipantsCount,
	}

	if fullInfo.LinkedChatID != 0 {
		info.LinkedChatID = &fullInfo.LinkedChatID
	}

	if fullInfo.CanViewStats {
		bs, err := s.mtproto.GetBroadcastStats(ctx, mtprotoID, fullInfo.StatsDC)
		if err != nil {
			s.log.Warn("failed to get broadcast stats, storing channel info only",
				"channel_id", channelID, "error", err)
		} else {
			info.Languages = bs.Languages
			info.TopHours = bs.TopHours
			info.ReactionsByEmotion = bs.ReactionsByEmotion
			info.StoryReactionsByEmotion = bs.StoryReactionsByEmotion
			info.RecentPosts = bs.RecentPosts

			if len(bs.DailyStats) > 0 {
				if err := s.channelRepo.BatchUpsertHistoricalStats(
					ctx, channelID, bs.DailyStats,
				); err != nil {
					return fmt.Errorf("batch upsert historical stats: %w", err)
				}
			}
		}
	}

	if err := s.channelRepo.UpsertInfo(ctx, info); err != nil {
		return fmt.Errorf("upsert channel info: %w", err)
	}

	go func() {
		if err := s.channelRepo.RefreshMV(context.Background()); err != nil {
			s.log.Warn("refresh marketplace mv", "error", err)
		}
	}()

	s.log.Info("channel stats fetched",
		"channel_id", channelID,
		"subscribers", info.Subscribers)

	return nil
}
