package post

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

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

func (r *repo) Create(
	ctx context.Context,
	userID uuid.UUID,
	mediaGroupID *string,
	text *string,
	entities []byte,
	mediaType *entity.MediaType,
	mediaFileID *string,
	hasMediaSpoiler bool,
	showCaptionAboveMedia bool,
) (*entity.Post, error) {
	id, err := uuid.NewV7()
	if err != nil {
		return nil, fmt.Errorf("creating post: %w", err)
	}

	rows, err := r.db.Query(ctx, `
		INSERT INTO posts (
			id, user_id, media_group_id, text, entities,
			media_type, media_file_id, has_media_spoiler,
			show_caption_above_media
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING
			id, user_id, media_group_id, text, entities,
			media_type, media_file_id, has_media_spoiler,
			show_caption_above_media, created_at, deleted_at
	`, id, userID, mediaGroupID, text, entities,
		mediaType, mediaFileID, hasMediaSpoiler, showCaptionAboveMedia)
	if err != nil {
		return nil, fmt.Errorf("creating post: %w", err)
	}

	p, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[entity.Post])
	if err != nil {
		return nil, fmt.Errorf("creating post: %w", err)
	}

	return &p, nil
}

func (r *repo) GetByUserID(ctx context.Context, userID uuid.UUID) ([]entity.Post, error) {
	rows, err := r.db.Query(ctx, `
		SELECT
			id, user_id, media_group_id, text, entities,
			media_type, media_file_id, has_media_spoiler,
			show_caption_above_media, created_at, deleted_at
		FROM posts
		WHERE user_id = $1 AND deleted_at IS NULL
		ORDER BY created_at DESC
	`, userID)
	if err != nil {
		return nil, fmt.Errorf("getting posts by user id: %w", err)
	}

	posts, err := pgx.CollectRows(rows, pgx.RowToStructByName[entity.Post])
	if err != nil {
		return nil, fmt.Errorf("getting posts by user id: %w", err)
	}

	return posts, nil
}

func (r *repo) SoftDelete(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.Exec(ctx, `UPDATE posts SET deleted_at = NOW() WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("soft deleting post: %w", err)
	}
	return nil
}
