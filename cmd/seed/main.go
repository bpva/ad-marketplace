package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"

	"github.com/bpva/ad-marketplace/internal/config"
	"github.com/bpva/ad-marketplace/internal/storage"
	"github.com/bpva/ad-marketplace/migrations"
	"github.com/bpva/ad-marketplace/seeds"
)

func main() {
	ctx := context.Background()
	log := slog.New(slog.NewTextHandler(os.Stdout, nil))

	_ = godotenv.Load()

	cfg, err := config.Load()
	if err != nil {
		log.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	// outside docker so
	cfg.Postgres.Host = "localhost"

	dsn := storage.URL(cfg.Postgres)

	if err := migrations.Run(dsn); err != nil {
		log.Error("migrations failed", "error", err)
		os.Exit(1)
	}

	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		log.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer pool.Close()

	if err := seeds.New(pool, log).Run(ctx); err != nil {
		log.Error("seeding failed", "error", err)
		os.Exit(1) //nolint:gocritic
	}

	fmt.Println("seeding complete")
}
