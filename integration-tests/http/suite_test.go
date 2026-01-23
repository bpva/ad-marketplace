//go:build integration

package http_test

import (
	"context"
	"log/slog"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
	"go.uber.org/mock/gomock"

	"github.com/bpva/ad-marketplace/integration-tests/tools"
	"github.com/bpva/ad-marketplace/internal/config"
	"github.com/bpva/ad-marketplace/internal/http/app"
	"github.com/bpva/ad-marketplace/internal/migrations"
	user_repo "github.com/bpva/ad-marketplace/internal/repository/user"
	"github.com/bpva/ad-marketplace/internal/service/auth"
)

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

	connStr, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		slog.Error("failed to get connection string", "error", err)
		os.Exit(1)
	}

	if err := migrations.Run(connStr); err != nil {
		slog.Error("failed to run migrations", "error", err)
		os.Exit(1)
	}

	testPool, err = pgxpool.New(ctx, connStr)
	if err != nil {
		slog.Error("failed to create pool", "error", err)
		os.Exit(1)
	}

	testTools = tools.New(testPool, testJWTSecret)
	testServer = setupTestServer()

	code := m.Run()

	testServer.Close()
	testPool.Close()
	_ = pgContainer.Terminate(ctx)

	os.Exit(code)
}

func setupTestServer() *httptest.Server {
	log := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))

	userRepo := user_repo.New(testPool)
	authSvc := auth.New(userRepo, testBotToken, testJWTSecret, log)

	httpCfg := config.HTTP{
		Port:        "0",
		FrontendURL: "*",
	}

	ctrl := gomock.NewController(&testing.T{})
	botMock := app.NewMockBotService(ctrl)
	botMock.EXPECT().Token().Return(testBotToken).AnyTimes()
	botMock.EXPECT().ProcessUpdate(gomock.Any()).Return(nil).AnyTimes()

	a := app.New(httpCfg, log, botMock, authSvc)
	return httptest.NewServer(a.Handler())
}
