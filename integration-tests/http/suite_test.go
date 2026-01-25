//go:build integration

package http_test

import (
	"context"
	"log/slog"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
	"go.uber.org/mock/gomock"

	"github.com/bpva/ad-marketplace/integration-tests/tools"
	"github.com/bpva/ad-marketplace/internal/config"
	"github.com/bpva/ad-marketplace/internal/http/app"
	"github.com/bpva/ad-marketplace/internal/migrations"
	channel_repo "github.com/bpva/ad-marketplace/internal/repository/channel"
	user_repo "github.com/bpva/ad-marketplace/internal/repository/user"
	"github.com/bpva/ad-marketplace/internal/service/auth"
	bot_service "github.com/bpva/ad-marketplace/internal/service/bot"
	channel_service "github.com/bpva/ad-marketplace/internal/service/channel"
	"github.com/bpva/ad-marketplace/internal/storage"
)

type db interface {
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
	storage.Transactor
}

var (
	testPool   *pgxpool.Pool
	testServer *httptest.Server
	testTools  *tools.Tools
)

const (
	testJWTSecret = "test-jwt-secret-32-bytes-long!!"
	testBotToken  = "123456:ABC-DEF1234ghIkl-zyx57W2v1u123ew11"
)

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

	testDB, err := storage.New(ctx, pgCfg)
	if err != nil {
		slog.Error("failed to create storage", "error", err)
		os.Exit(1)
	}

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

	testTools = tools.New(testPool, testJWTSecret)
	testServer = setupTestServer(testDB)

	code := m.Run()

	testServer.Close()
	testPool.Close()
	_ = pgContainer.Terminate(ctx)

	os.Exit(code)
}

func setupTestServer(testDB db) *httptest.Server {
	log := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	userRepo := user_repo.New(testDB)
	channelRepo := channel_repo.New(testDB)
	authSvc := auth.New(userRepo, testBotToken, testJWTSecret, log)

	httpCfg := config.HTTP{
		Port:        "0",
		FrontendURL: "*",
	}

	ctrl := gomock.NewController(&testing.T{})
	telebotMock := bot_service.NewMockTelebotClient(ctrl)
	telebotMock.EXPECT().Handle(gomock.Any(), gomock.Any()).AnyTimes()
	telebotMock.EXPECT().Token().Return(testBotToken).AnyTimes()
	telebotMock.EXPECT().AdminsOf(gomock.Any()).Return(nil, nil).AnyTimes()

	botSvc := bot_service.New(telebotMock, "http://localhost", log, testDB, channelRepo, userRepo)
	channelSvc := channel_service.New(channelRepo, userRepo, telebotMock, log)

	a := app.New(httpCfg, log, botSvc, authSvc, channelSvc)
	return httptest.NewServer(a.Handler())
}
