package settings

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/bpva/ad-marketplace/internal/dto"
	"github.com/bpva/ad-marketplace/internal/entity"
)

type db interface {
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
}

type repo struct {
	db db
}

func New(db db) *repo {
	return &repo{db: db}
}

func (r *repo) GetByUserID(ctx context.Context, userID uuid.UUID) (*entity.UserSettings, error) {
	rows, err := r.db.Query(ctx, `
		SELECT user_id, language, receive_notifications, preferred_mode, onboarding_finished
		FROM user_settings
		WHERE user_id = $1
	`, userID)
	if err != nil {
		return nil, fmt.Errorf("getting settings by user id: %w", err)
	}

	s, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[entity.UserSettings])
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, fmt.Errorf("getting settings by user id: %w", dto.ErrNotFound)
	}
	if err != nil {
		return nil, fmt.Errorf("getting settings by user id: %w", err)
	}

	return &s, nil
}

func (r *repo) Create(ctx context.Context, userID uuid.UUID) (*entity.UserSettings, error) {
	rows, err := r.db.Query(ctx, `
		INSERT INTO user_settings (user_id)
		VALUES ($1)
		RETURNING user_id, language, receive_notifications, preferred_mode, onboarding_finished
	`, userID)
	if err != nil {
		return nil, fmt.Errorf("creating settings: %w", err)
	}

	s, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[entity.UserSettings])
	if err != nil {
		return nil, fmt.Errorf("creating settings: %w", err)
	}

	return &s, nil
}

func (r *repo) Update(ctx context.Context, s *entity.UserSettings) error {
	_, err := r.db.Exec(ctx, `
		UPDATE user_settings
		SET language = $2, receive_notifications = $3, preferred_mode = $4, onboarding_finished = $5
		WHERE user_id = $1
	`, s.UserID, s.Language, s.ReceiveNotifications, s.PreferredMode, s.OnboardingFinished)
	if err != nil {
		return fmt.Errorf("updating settings: %w", err)
	}
	return nil
}
