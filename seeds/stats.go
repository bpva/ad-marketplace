package seeds

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/bpva/ad-marketplace/internal/entity"
	channel_repo "github.com/bpva/ad-marketplace/internal/repository/channel"
)

func (s *Seeder) seedStats(ctx context.Context, channels []seedChannel) error {
	for _, ch := range channels {
		if err := s.upsertChannelInfo(ctx, ch); err != nil {
			return fmt.Errorf("seed info for %s: %w", ch.entity.Title, err)
		}
		if err := s.upsertHistoricalStats(ctx, ch); err != nil {
			return fmt.Errorf("seed history for %s: %w", ch.entity.Title, err)
		}
	}
	return nil
}

func (s *Seeder) upsertChannelInfo(ctx context.Context, ch seedChannel) error {
	repo := channel_repo.New(s.db)

	languages := []map[string]any{
		{"language": "en", "percentage": 55 + s.rng.IntN(20)},
		{"language": "ru", "percentage": 15 + s.rng.IntN(15)},
		{"language": "uk", "percentage": 5 + s.rng.IntN(10)},
	}

	topHours := make([]float64, 24)
	for i := range topHours {
		topHours[i] = 0.5 + s.rng.Float64()*2.0
	}
	for _, peak := range []int{9, 10, 13, 14, 19, 20, 21} {
		topHours[peak] = 3.0 + s.rng.Float64()*3.0
	}

	reactions := map[string]int{
		"\U0001f44d":   500 + s.rng.IntN(2000),
		"\u2764\ufe0f": 300 + s.rng.IntN(1500),
		"\U0001f525":   200 + s.rng.IntN(1000),
		"\U0001f914":   50 + s.rng.IntN(300),
	}

	info := &entity.ChannelInfo{
		ChannelID:               ch.entity.ID,
		About:                   fmt.Sprintf("Official %s channel", ch.entity.Title),
		Subscribers:             ch.subscribers,
		Languages:               mustJSON(languages),
		TopHours:                mustJSON(topHours),
		ReactionsByEmotion:      mustJSON(reactions),
		StoryReactionsByEmotion: mustJSON(reactions),
	}

	return repo.UpsertInfo(ctx, info)
}

func (s *Seeder) upsertHistoricalStats(ctx context.Context, ch seedChannel) error {
	repo := channel_repo.New(s.db)

	days := 90
	stats := make([]entity.ChannelHistoricalStats, 0, days)
	subs := ch.subscribers - s.rng.IntN(ch.subscribers/5)
	baseViews := ch.subscribers / 5

	for i := days - 1; i >= 0; i-- {
		date := time.Now().AddDate(0, 0, -i).Truncate(24 * time.Hour)

		newFollowers := max(
			0,
			ch.subscribers/500+s.rng.IntN(ch.subscribers/200)-ch.subscribers/400,
		)
		subs += newFollowers

		views := baseViews + s.rng.IntN(baseViews/2) - baseViews/4
		interactions := views/10 + s.rng.IntN(views/20)

		daily := map[string]any{
			"subscribers":   subs,
			"new_followers": newFollowers,
			"interactions":  interactions,
			"views_by_source": map[string]int{
				"search":   views / 3,
				"channels": views / 2,
				"other":    views - views/3 - views/2,
			},
			"followers_by_source": map[string]int{
				"search":   newFollowers / 2,
				"channels": newFollowers / 3,
				"other":    max(0, newFollowers-newFollowers/2-newFollowers/3),
			},
		}

		stats = append(stats, entity.ChannelHistoricalStats{
			ChannelID: ch.entity.ID,
			Date:      date,
			Data:      mustJSON(daily),
		})
	}

	return repo.BatchUpsertHistoricalStats(ctx, stats)
}

func mustJSON(v any) []byte {
	b, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	return b
}
