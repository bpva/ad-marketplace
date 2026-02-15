package seeds

import (
	"context"
	"fmt"
	"log/slog"
	"math/rand/v2"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Seeder struct {
	db  *pgxpool.Pool
	rng *rand.Rand
	log *slog.Logger
}

func New(db *pgxpool.Pool, log *slog.Logger) *Seeder {
	return &Seeder{
		db:  db,
		rng: rand.New(rand.NewPCG(42, 0)),
		log: log,
	}
}

func (s *Seeder) Run(ctx context.Context) error {
	if err := s.truncate(ctx); err != nil {
		return fmt.Errorf("truncate: %w", err)
	}
	s.log.Info("tables truncated")

	users, err := s.seedUsers(ctx)
	if err != nil {
		return fmt.Errorf("seed users: %w", err)
	}
	s.log.Info("users seeded", "count", len(users))

	channels, err := s.seedChannels(ctx, users)
	if err != nil {
		return fmt.Errorf("seed channels: %w", err)
	}
	s.log.Info("channels seeded", "count", len(channels))

	if err := s.seedPosts(ctx, users); err != nil {
		return fmt.Errorf("seed posts: %w", err)
	}
	s.log.Info("posts seeded")

	if err := s.seedStats(ctx, channels); err != nil {
		return fmt.Errorf("seed stats: %w", err)
	}
	s.log.Info("stats seeded")

	if _, err := s.db.Exec(ctx, "REFRESH MATERIALIZED VIEW channel_marketplace"); err != nil {
		return fmt.Errorf("refresh marketplace mv: %w", err)
	}
	s.log.Info("marketplace mv refreshed")

	return nil
}

func (s *Seeder) truncate(ctx context.Context) error {
	_, err := s.db.Exec(ctx, `
		TRUNCATE
			posts,
			channel_historical_stats,
			channel_info,
			channel_ad_formats,
			channel_roles,
			user_settings,
			channels,
			users
		CASCADE
	`)
	return err
}
