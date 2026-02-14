package storage

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/bpva/ad-marketplace/internal/config"
)

type Transactor interface {
	WithTx(ctx context.Context, f func(ctx context.Context) error) error
}

type txCtxKey struct{}

type db struct {
	pool *pgxpool.Pool
}

func New(ctx context.Context, cfg config.Postgres) (*db, error) {
	dsn := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=disable",
		cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.DB,
	)

	poolCfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}

	pool, err := pgxpool.NewWithConfig(ctx, poolCfg)
	if err != nil {
		return nil, fmt.Errorf("create pool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("ping: %w", err)
	}

	return &db{pool: pool}, nil
}

func (d *db) Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error) {
	if tx, ok := ctx.Value(txCtxKey{}).(pgx.Tx); ok {
		return tx.Query(ctx, sql, args...)
	}
	return d.pool.Query(ctx, sql, args...)
}

func (d *db) Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error) {
	if tx, ok := ctx.Value(txCtxKey{}).(pgx.Tx); ok {
		return tx.Exec(ctx, sql, args...)
	}
	return d.pool.Exec(ctx, sql, args...)
}

func (d *db) WithTx(ctx context.Context, f func(ctx context.Context) error) error {
	tx, err := d.pool.Begin(ctx)
	if err != nil {
		return err
	}

	txCtx := context.WithValue(ctx, txCtxKey{}, tx)
	if err := f(txCtx); err != nil {
		if rbErr := tx.Rollback(ctx); rbErr != nil {
			return fmt.Errorf("%w (rollback failed: %v)", err, rbErr)
		}
		return err
	}

	return tx.Commit(ctx)
}

func (d *db) Close() {
	d.pool.Close()
}

func URL(cfg config.Postgres) string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=disable",
		cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.DB,
	)
}
