package user

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/bpva/ad-marketplace/internal/dto"
	"github.com/bpva/ad-marketplace/internal/entity"
)

type db interface {
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
}

type repo struct {
	db db
}

func New(db db) *repo {
	return &repo{db: db}
}

func (r *repo) GetByTgID(ctx context.Context, tgID int64) (*entity.User, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, telegram_id, name, created_at, deleted_at
		FROM users
		WHERE telegram_id = $1 AND deleted_at IS NULL
	`, tgID)
	if err != nil {
		return nil, fmt.Errorf("getting user by telegram id: %w", err)
	}

	u, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[entity.User])
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, fmt.Errorf("getting user by telegram id: %w", dto.ErrNotFound)
	}
	if err != nil {
		return nil, fmt.Errorf("getting user by telegram id: %w", err)
	}

	return &u, nil
}

func (r *repo) Create(ctx context.Context, tgID int64, name string) (*entity.User, error) {
	id, err := uuid.NewV7()
	if err != nil {
		return nil, fmt.Errorf("creating user: %w", err)
	}

	rows, err := r.db.Query(ctx, `
		INSERT INTO users (id, telegram_id, name)
		VALUES ($1, $2, $3)
		RETURNING id, telegram_id, name, created_at, deleted_at
	`, id, tgID, name)
	if err != nil {
		return nil, fmt.Errorf("creating user: %w", err)
	}

	u, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[entity.User])
	if err != nil {
		return nil, fmt.Errorf("creating user: %w", err)
	}

	return &u, nil
}

func (r *repo) GetByID(ctx context.Context, id uuid.UUID) (*entity.User, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, telegram_id, name, created_at, deleted_at
		FROM users
		WHERE id = $1 AND deleted_at IS NULL
	`, id)
	if err != nil {
		return nil, fmt.Errorf("getting user by id: %w", err)
	}

	u, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[entity.User])
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, fmt.Errorf("getting user by id: %w", dto.ErrNotFound)
	}
	if err != nil {
		return nil, fmt.Errorf("getting user by id: %w", err)
	}

	return &u, nil
}
