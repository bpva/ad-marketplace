package channel

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/bpva/ad-marketplace/internal/dto"
	"github.com/bpva/ad-marketplace/internal/entity"
)

func (r *repo) UpsertInfo(ctx context.Context, info *entity.ChannelInfo) error {
	languages, _ := json.Marshal(info.Languages)
	topHours, _ := json.Marshal(info.TopHours)
	reactionsByEmotion, _ := json.Marshal(info.ReactionsByEmotion)
	storyReactionsByEmotion, _ := json.Marshal(info.StoryReactionsByEmotion)
	recentPosts, _ := json.Marshal(info.RecentPosts)

	_, err := r.db.Exec(ctx, `
		INSERT INTO channel_info (
			channel_id, about, subscribers, linked_chat_id,
			languages, top_hours, reactions_by_emotion,
			story_reactions_by_emotion, recent_posts, fetched_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, NOW())
		ON CONFLICT (channel_id) DO UPDATE SET
			about = EXCLUDED.about,
			subscribers = EXCLUDED.subscribers,
			linked_chat_id = EXCLUDED.linked_chat_id,
			languages = EXCLUDED.languages,
			top_hours = EXCLUDED.top_hours,
			reactions_by_emotion = EXCLUDED.reactions_by_emotion,
			story_reactions_by_emotion = EXCLUDED.story_reactions_by_emotion,
			recent_posts = EXCLUDED.recent_posts,
			fetched_at = NOW()
	`,
		info.ChannelID, info.About, info.Subscribers, info.LinkedChatID,
		languages, topHours, reactionsByEmotion,
		storyReactionsByEmotion, recentPosts,
	)
	if err != nil {
		return fmt.Errorf("upserting channel info: %w", err)
	}
	return nil
}

func (r *repo) GetInfo(ctx context.Context, channelID uuid.UUID) (*entity.ChannelInfo, error) {
	rows, err := r.db.Query(ctx, `
		SELECT channel_id, about, subscribers, linked_chat_id,
			languages, top_hours, reactions_by_emotion,
			story_reactions_by_emotion, recent_posts, fetched_at
		FROM channel_info
		WHERE channel_id = $1
	`, channelID)
	if err != nil {
		return nil, fmt.Errorf("getting channel info: %w", err)
	}

	info, err := pgx.CollectOneRow(rows, pgx.RowToStructByNameLax[entity.ChannelInfo])
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, fmt.Errorf("getting channel info: %w", dto.ErrNotFound)
	}
	if err != nil {
		return nil, fmt.Errorf("getting channel info: %w", err)
	}

	return &info, nil
}

func (r *repo) BatchUpsertHistoricalStats(
	ctx context.Context,
	channelID uuid.UUID,
	stats []entity.DailyMetrics,
) error {
	for _, dm := range stats {
		data, err := json.Marshal(dm.Data)
		if err != nil {
			continue
		}
		_, err = r.db.Exec(ctx, `
			INSERT INTO channel_historical_stats (channel_id, date, data)
			VALUES ($1, $2, $3)
			ON CONFLICT (channel_id, date) DO NOTHING
		`, channelID, dm.Date, data)
		if err != nil {
			return fmt.Errorf("upserting historical stats: %w", err)
		}
	}
	return nil
}
