package post

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

func (r *repo) Create(
	ctx context.Context,
	postType entity.PostType,
	externalID uuid.UUID,
	version *int,
	name *string,
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
			id, type, external_id, version, name, media_group_id, text, entities,
			media_type, media_file_id, has_media_spoiler,
			show_caption_above_media
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		RETURNING
			id, type, external_id, version, name, media_group_id, text, entities,
			media_type, media_file_id, has_media_spoiler,
			show_caption_above_media, created_at, deleted_at
	`, id, postType, externalID, version, name, mediaGroupID, text, entities,
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

func (r *repo) GetTemplatesByOwner(ctx context.Context, ownerID uuid.UUID) ([]entity.Post, error) {
	rows, err := r.db.Query(ctx, `
		SELECT
			id, type, external_id, version, name, media_group_id, text, entities,
			media_type, media_file_id, has_media_spoiler,
			show_caption_above_media, created_at, deleted_at
		FROM posts
		WHERE type = 'template' AND external_id = $1 AND deleted_at IS NULL
		ORDER BY created_at DESC
	`, ownerID)
	if err != nil {
		return nil, fmt.Errorf("getting templates by owner: %w", err)
	}

	posts, err := pgx.CollectRows(rows, pgx.RowToStructByName[entity.Post])
	if err != nil {
		return nil, fmt.Errorf("getting templates by owner: %w", err)
	}

	return posts, nil
}

func (r *repo) GetByID(ctx context.Context, id uuid.UUID) (*entity.Post, error) {
	rows, err := r.db.Query(ctx, `
		SELECT
			id, type, external_id, version, name, media_group_id, text, entities,
			media_type, media_file_id, has_media_spoiler,
			show_caption_above_media, created_at, deleted_at
		FROM posts
		WHERE id = $1 AND deleted_at IS NULL
	`, id)
	if err != nil {
		return nil, fmt.Errorf("getting post by id: %w", err)
	}

	p, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[entity.Post])
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, fmt.Errorf("getting post by id: %w", dto.ErrNotFound)
	}
	if err != nil {
		return nil, fmt.Errorf("getting post by id: %w", err)
	}

	return &p, nil
}

func (r *repo) GetByMediaGroupID(ctx context.Context, mediaGroupID string) ([]entity.Post, error) {
	rows, err := r.db.Query(ctx, `
		SELECT
			id, type, external_id, version, name, media_group_id, text, entities,
			media_type, media_file_id, has_media_spoiler,
			show_caption_above_media, created_at, deleted_at
		FROM posts
		WHERE media_group_id = $1 AND deleted_at IS NULL
		ORDER BY created_at ASC
	`, mediaGroupID)
	if err != nil {
		return nil, fmt.Errorf("getting posts by media group id: %w", err)
	}

	posts, err := pgx.CollectRows(rows, pgx.RowToStructByName[entity.Post])
	if err != nil {
		return nil, fmt.Errorf("getting posts by media group id: %w", err)
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

func (r *repo) CopyAsAd(
	ctx context.Context, templatePostID, dealID uuid.UUID, version int,
) ([]entity.Post, error) {
	source, err := r.GetByID(ctx, templatePostID)
	if err != nil {
		return nil, fmt.Errorf("copying as ad: %w", err)
	}

	var sources []entity.Post
	if source.MediaGroupID != nil {
		sources, err = r.GetByMediaGroupID(ctx, *source.MediaGroupID)
		if err != nil {
			return nil, fmt.Errorf("copying as ad: %w", err)
		}
	} else {
		sources = []entity.Post{*source}
	}

	var newMediaGroupID *string
	if source.MediaGroupID != nil {
		mgID := uuid.Must(uuid.NewV7()).String()
		newMediaGroupID = &mgID
	}

	var result []entity.Post
	for i, s := range sources {
		var name *string
		if i == 0 {
			name = s.Name
		}
		p, err := r.Create(ctx, entity.PostTypeAd, dealID, &version, name,
			newMediaGroupID, s.Text, s.Entities, s.MediaType, s.MediaFileID,
			s.HasMediaSpoiler, s.ShowCaptionAboveMedia)
		if err != nil {
			return nil, fmt.Errorf("copying as ad: %w", err)
		}
		result = append(result, *p)
	}

	return result, nil
}

func (r *repo) AddAdVersion(
	ctx context.Context, dealID uuid.UUID, version int, posts []entity.Post,
) ([]entity.Post, error) {
	var newMediaGroupID *string
	if len(posts) > 1 {
		mgID := uuid.Must(uuid.NewV7()).String()
		newMediaGroupID = &mgID
	}

	var result []entity.Post
	for _, s := range posts {
		p, err := r.Create(ctx, entity.PostTypeAd, dealID, &version, s.Name,
			newMediaGroupID, s.Text, s.Entities, s.MediaType, s.MediaFileID,
			s.HasMediaSpoiler, s.ShowCaptionAboveMedia)
		if err != nil {
			return nil, fmt.Errorf("adding ad version: %w", err)
		}
		result = append(result, *p)
	}

	return result, nil
}

func (r *repo) GetAdVersions(ctx context.Context, dealID uuid.UUID) (map[int][]entity.Post, error) {
	rows, err := r.db.Query(ctx, `
		SELECT
			id, type, external_id, version, name, media_group_id, text, entities,
			media_type, media_file_id, has_media_spoiler,
			show_caption_above_media, created_at, deleted_at
		FROM posts
		WHERE type = 'ad' AND external_id = $1 AND deleted_at IS NULL
		ORDER BY version ASC, created_at ASC
	`, dealID)
	if err != nil {
		return nil, fmt.Errorf("getting ad versions: %w", err)
	}

	posts, err := pgx.CollectRows(rows, pgx.RowToStructByName[entity.Post])
	if err != nil {
		return nil, fmt.Errorf("getting ad versions: %w", err)
	}

	versions := make(map[int][]entity.Post)
	for _, p := range posts {
		if p.Version != nil {
			versions[*p.Version] = append(versions[*p.Version], p)
		}
	}

	return versions, nil
}

func (r *repo) GetLatestAd(ctx context.Context, dealID uuid.UUID) ([]entity.Post, error) {
	rows, err := r.db.Query(ctx, `
		SELECT
			id, type, external_id, version, name, media_group_id, text, entities,
			media_type, media_file_id, has_media_spoiler,
			show_caption_above_media, created_at, deleted_at
		FROM posts
		WHERE type = 'ad' AND external_id = $1 AND deleted_at IS NULL
			AND version = (
				SELECT MAX(version) FROM posts
				WHERE type = 'ad' AND external_id = $1 AND deleted_at IS NULL
			)
		ORDER BY created_at ASC
	`, dealID)
	if err != nil {
		return nil, fmt.Errorf("getting latest ad: %w", err)
	}

	posts, err := pgx.CollectRows(rows, pgx.RowToStructByName[entity.Post])
	if err != nil {
		return nil, fmt.Errorf("getting latest ad: %w", err)
	}

	return posts, nil
}
