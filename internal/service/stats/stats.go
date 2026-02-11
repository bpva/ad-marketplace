package stats

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"

	"github.com/bpva/ad-marketplace/internal/dto"
	"github.com/bpva/ad-marketplace/internal/entity"
	"github.com/bpva/ad-marketplace/internal/gateway/mtproto"
	"github.com/bpva/ad-marketplace/internal/logx"
)

type ChannelRepository interface {
	GetByTgChannelID(ctx context.Context, tgChannelID int64) (*entity.Channel, error)
	UpsertInfo(ctx context.Context, info *entity.ChannelInfo) error
	GetInfo(ctx context.Context, channelID uuid.UUID) (*entity.ChannelInfo, error)
	BatchUpsertHistoricalStats(ctx context.Context, stats []entity.ChannelHistoricalStats) error
	GetHistoricalStats(
		ctx context.Context,
		channelID uuid.UUID,
		from, to time.Time,
	) ([]entity.ChannelHistoricalStats, error)
}

//go:generate mockgen -destination=mocks.go -package=stats . MTProtoClient
type MTProtoClient interface {
	ResolveChannel(ctx context.Context, channelID int64) (int64, error)
	GetChannelFull(ctx context.Context, channelID, accessHash int64) (*dto.ChannelFullInfo, error)
	GetBroadcastStats(
		ctx context.Context,
		channelID, accessHash int64,
		statsDC int,
	) (*dto.BroadcastStatsResult, error)
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

	accessHash, err := s.mtproto.ResolveChannel(ctx, mtprotoID)
	if err != nil {
		return fmt.Errorf("resolve channel: %w", err)
	}

	fullInfo, err := s.mtproto.GetChannelFull(ctx, mtprotoID, accessHash)
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
		statsResult, err := s.mtproto.GetBroadcastStats(ctx, mtprotoID, accessHash, fullInfo.StatsDC)
		if err != nil {
			s.log.Warn("failed to get broadcast stats, storing channel info only",
				"channel_id", channelID, "error", err)
		} else {
			info.Languages = statsResult.Languages
			info.TopHours = statsResult.TopHours
			info.ReactionsByEmotion = statsResult.ReactionsByEmotion
			info.StoryReactionsByEmotion = statsResult.StoryReactionsByEmotion
			info.RecentPosts = statsResult.RecentPosts

			if len(statsResult.DailyStats) > 0 {
				historicalRows := make(
					[]entity.ChannelHistoricalStats,
					0,
					len(statsResult.DailyStats),
				)
				for dateStr, metrics := range statsResult.DailyStats {
					date, err := time.Parse("2006-01-02", dateStr)
					if err != nil {
						continue
					}
					data, err := json.Marshal(metrics)
					if err != nil {
						continue
					}
					historicalRows = append(historicalRows, entity.ChannelHistoricalStats{
						ChannelID: channelID,
						Date:      date,
						Data:      data,
					})
				}
				if err := s.channelRepo.BatchUpsertHistoricalStats(
					ctx,
					historicalRows,
				); err != nil {
					return fmt.Errorf("batch upsert historical stats: %w", err)
				}
			}
		}
	}

	if err := s.channelRepo.UpsertInfo(ctx, info); err != nil {
		return fmt.Errorf("upsert channel info: %w", err)
	}

	s.log.Info("channel stats fetched",
		"channel_id", channelID,
		"subscribers", info.Subscribers)

	return nil
}

func (s *svc) GetInfo(ctx context.Context, tgChannelID int64) (*dto.ChannelInfoResponse, error) {
	ch, err := s.channelRepo.GetByTgChannelID(ctx, tgChannelID)
	if err != nil {
		return nil, fmt.Errorf("get channel: %w", err)
	}

	info, err := s.channelRepo.GetInfo(ctx, ch.ID)
	if err != nil {
		return nil, fmt.Errorf("get channel info: %w", err)
	}

	return &dto.ChannelInfoResponse{
		About:                   info.About,
		Subscribers:             info.Subscribers,
		Languages:               info.Languages,
		TopHours:                info.TopHours,
		ReactionsByEmotion:      info.ReactionsByEmotion,
		StoryReactionsByEmotion: info.StoryReactionsByEmotion,
		RecentPosts:             info.RecentPosts,
		FetchedAt:               info.FetchedAt,
	}, nil
}

func (s *svc) GetHistory(
	ctx context.Context,
	tgChannelID int64,
	from, to time.Time,
) (*dto.ChannelHistoricalStatsResponse, error) {
	ch, err := s.channelRepo.GetByTgChannelID(ctx, tgChannelID)
	if err != nil {
		return nil, fmt.Errorf("get channel: %w", err)
	}

	if from.IsZero() {
		from = time.Now().AddDate(0, 0, -30)
	}
	if to.IsZero() {
		to = time.Now()
	}

	rows, err := s.channelRepo.GetHistoricalStats(ctx, ch.ID, from, to)
	if err != nil {
		return nil, fmt.Errorf("get historical stats: %w", err)
	}

	result := make([]dto.ChannelDailyStats, 0, len(rows))
	for _, row := range rows {
		result = append(result, dto.ChannelDailyStats{
			Date: row.Date.Format("2006-01-02"),
			Data: row.Data,
		})
	}

	return &dto.ChannelHistoricalStatsResponse{Stats: result}, nil
}
