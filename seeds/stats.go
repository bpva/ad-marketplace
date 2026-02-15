package seeds

import (
	"context"
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

	languages := []entity.LanguageShare{
		{Language: "en", Percentage: float64(55 + s.rng.IntN(20))},
		{Language: "ru", Percentage: float64(15 + s.rng.IntN(15))},
		{Language: "uk", Percentage: float64(5 + s.rng.IntN(10))},
	}

	topHours := make([]float64, 24)
	for i := range topHours {
		topHours[i] = 0.5 + s.rng.Float64()*2.0
	}
	for _, peak := range []int{9, 10, 13, 14, 19, 20, 21} {
		topHours[peak] = 3.0 + s.rng.Float64()*3.0
	}

	reactions := map[string]int64{
		"\U0001f44d":   int64(500 + s.rng.IntN(2000)),
		"\u2764\ufe0f": int64(300 + s.rng.IntN(1500)),
		"\U0001f525":   int64(200 + s.rng.IntN(1000)),
		"\U0001f914":   int64(50 + s.rng.IntN(300)),
	}

	storyReactions := map[string]int64{
		"\U0001f44d":   int64(100 + s.rng.IntN(500)),
		"\u2764\ufe0f": int64(80 + s.rng.IntN(400)),
		"\U0001f525":   int64(40 + s.rng.IntN(200)),
	}

	recentPosts := make([]entity.RecentPost, 0, 10)
	for i := range 10 {
		recentPosts = append(recentPosts, entity.RecentPost{
			Type:      "message",
			ID:        1000 + i,
			Views:     ch.subscribers/10 + s.rng.IntN(ch.subscribers/5),
			Forwards:  s.rng.IntN(50),
			Reactions: s.rng.IntN(200),
		})
	}

	info := &entity.ChannelInfo{
		ChannelID:               ch.entity.ID,
		About:                   fmt.Sprintf("Official %s channel", ch.entity.Title),
		Subscribers:             ch.subscribers,
		Languages:               languages,
		TopHours:                topHours,
		ReactionsByEmotion:      reactions,
		StoryReactionsByEmotion: storyReactions,
		RecentPosts:             recentPosts,
	}

	return repo.UpsertInfo(ctx, info)
}

func (s *Seeder) upsertHistoricalStats(ctx context.Context, ch seedChannel) error {
	repo := channel_repo.New(s.db)

	days := 90
	stats := make([]entity.DailyMetrics, 0, days)
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
		ivInteractions := s.rng.IntN(views / 20)
		mutePct := 0.5 + s.rng.Float64()*4.0
		storyViews := s.rng.IntN(views / 10)
		storyShares := s.rng.IntN(max(1, storyViews/20))

		subsI64 := int64(subs)
		newFollowersI64 := int64(newFollowers)
		interactionsI64 := int64(interactions)
		ivInteractionsI64 := int64(ivInteractions)

		stats = append(stats, entity.DailyMetrics{
			Date: date,
			Data: entity.ChannelHistoricalDayData{
				Subscribers:    &subsI64,
				NewFollowers:   &newFollowersI64,
				MutePct:        &mutePct,
				Interactions:   &interactionsI64,
				IVInteractions: &ivInteractionsI64,
				ViewsBySource: map[string]int64{
					"search":   int64(views / 3),
					"channels": int64(views / 2),
					"other":    int64(views - views/3 - views/2),
				},
				FollowersBySource: map[string]int64{
					"search":   int64(newFollowers / 2),
					"channels": int64(newFollowers / 3),
					"other":    int64(max(0, newFollowers-newFollowers/2-newFollowers/3)),
				},
				StoryInteractions: map[string]int64{
					"Views":  int64(storyViews),
					"Shares": int64(storyShares),
				},
				ReactionsByEmotion: map[string]int64{
					"\U0001f44d":   int64(s.rng.IntN(interactions/5 + 1)),
					"\u2764\ufe0f": int64(s.rng.IntN(interactions/8 + 1)),
					"\U0001f525":   int64(s.rng.IntN(interactions/10 + 1)),
				},
				StoryReactionsByEmotion: map[string]int64{
					"\U0001f44d":   int64(s.rng.IntN(max(1, storyViews/10))),
					"\u2764\ufe0f": int64(s.rng.IntN(max(1, storyViews/15))),
				},
			},
		})
	}

	return repo.BatchUpsertHistoricalStats(ctx, ch.entity.ID, stats)
}
