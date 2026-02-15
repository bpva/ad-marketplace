//go:build integration

package bot_test

import (
	"context"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"

	"go.uber.org/mock/gomock"

	"github.com/bpva/ad-marketplace/integration-tests/tools"
	"github.com/bpva/ad-marketplace/internal/config"
	"github.com/bpva/ad-marketplace/internal/dto"
	"github.com/bpva/ad-marketplace/internal/entity"
	channel_repo "github.com/bpva/ad-marketplace/internal/repository/channel"
	post_repo "github.com/bpva/ad-marketplace/internal/repository/post"
	user_repo "github.com/bpva/ad-marketplace/internal/repository/user"
	"github.com/bpva/ad-marketplace/internal/service/bot"
	"github.com/bpva/ad-marketplace/internal/service/stats"
	"github.com/bpva/ad-marketplace/internal/storage"
	"github.com/bpva/ad-marketplace/migrations"
)

type db interface {
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
	storage.Transactor
}

var (
	testPool    *pgxpool.Pool
	testDB      db
	testTools   *tools.Tools
	channelRepo bot.ChannelRepository
	userRepo    bot.UserRepository
	postRepo    bot.PostRepository
	statsSvc    bot.StatsFetcher
	log         *slog.Logger
)

const testBotToken = "123456:ABC-DEF1234ghIkl-zyx57W2v1u123ew11"

func TestMain(m *testing.M) {
	ctx := context.Background()

	pgContainer, err := postgres.Run(ctx,
		"postgres:16-alpine",
		postgres.WithDatabase("test_db"),
		postgres.WithUsername("test"),
		postgres.WithPassword("test"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2),
		),
	)
	if err != nil {
		slog.Error("failed to start postgres container", "error", err)
		os.Exit(1)
	}

	host, err := pgContainer.Host(ctx)
	if err != nil {
		slog.Error("failed to get host", "error", err)
		os.Exit(1)
	}

	port, err := pgContainer.MappedPort(ctx, "5432/tcp")
	if err != nil {
		slog.Error("failed to get port", "error", err)
		os.Exit(1)
	}

	pgCfg := config.Postgres{
		Host:     host,
		Port:     port.Port(),
		User:     "test",
		Password: "test",
		DB:       "test_db",
	}

	if err := migrations.Run(storage.URL(pgCfg)); err != nil {
		slog.Error("failed to run migrations", "error", err)
		os.Exit(1)
	}

	testDB, err = storage.New(ctx, pgCfg)
	if err != nil {
		slog.Error("failed to create storage", "error", err)
		os.Exit(1)
	}

	channelRepo = channel_repo.New(testDB)
	userRepo = user_repo.New(testDB)
	postRepo = post_repo.New(testDB)
	log = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))

	ctrl := gomock.NewController(&testing.T{})
	mockMTProto := stats.NewMockMTProtoClient(ctrl)
	mockMTProto.EXPECT().
		GetChannelFull(gomock.Any(), gomock.Any()).
		Return(&dto.ChannelFullInfo{
			ParticipantsCount: 1500,
			About:             "Test channel about",
			CanViewStats:      true,
			StatsDC:           2,
		}, nil).
		AnyTimes()
	mockMTProto.EXPECT().
		GetBroadcastStats(gomock.Any(), gomock.Any(), gomock.Any()).
		Return(&entity.BroadcastStats{
			DailyStats: []entity.DailyMetrics{
				{
					Date: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
					Data: entity.ChannelHistoricalDayData{
						Subscribers:  ptrInt64(1450),
						NewFollowers: ptrInt64(50),
					},
				},
				{
					Date: time.Date(2025, 1, 2, 0, 0, 0, 0, time.UTC),
					Data: entity.ChannelHistoricalDayData{
						Subscribers:  ptrInt64(1500),
						NewFollowers: ptrInt64(50),
					},
				},
			},
		}, nil).
		AnyTimes()
	statsSvc = stats.New(mockMTProto, channel_repo.New(testDB), log)

	connStr, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		slog.Error("failed to get connection string", "error", err)
		os.Exit(1)
	}

	testPool, err = pgxpool.New(ctx, connStr)
	if err != nil {
		slog.Error("failed to create pool", "error", err)
		os.Exit(1)
	}

	testTools = tools.New(testPool, "unused-jwt-secret")

	code := m.Run()

	testPool.Close()
	_ = pgContainer.Terminate(ctx)

	os.Exit(code)
}

func ptrInt64(v int64) *int64 {
	return &v
}
