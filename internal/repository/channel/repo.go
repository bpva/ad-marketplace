package channel

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgerrcode"
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
	tgChannelID int64,
	title string,
	username *string,
) (*entity.Channel, error) {
	id, err := uuid.NewV7()
	if err != nil {
		return nil, fmt.Errorf("creating channel: %w", err)
	}

	rows, err := r.db.Query(ctx, `
		INSERT INTO channels (id, telegram_channel_id, title, username)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (telegram_channel_id) DO UPDATE SET
			title = EXCLUDED.title,
			username = EXCLUDED.username,
			deleted_at = NULL
		RETURNING id, telegram_channel_id, title, username, is_listed,
			photo_small_file_id, photo_big_file_id, created_at
	`, id, tgChannelID, title, username)
	if err != nil {
		return nil, fmt.Errorf("creating channel: %w", err)
	}

	ch, err := pgx.CollectOneRow(rows, pgx.RowToStructByNameLax[entity.Channel])
	if err != nil {
		return nil, fmt.Errorf("creating channel: %w", err)
	}

	return &ch, nil
}

func (r *repo) GetByTgChannelID(
	ctx context.Context,
	tgChannelID int64,
) (*entity.Channel, error) {
	rows, err := r.db.Query(ctx, `
		SELECT * FROM channels
		WHERE telegram_channel_id = $1 AND deleted_at IS NULL
	`, tgChannelID)
	if err != nil {
		return nil, fmt.Errorf("getting channel by telegram channel id: %w", err)
	}

	ch, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[entity.Channel])
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, fmt.Errorf("getting channel by telegram channel id: %w", dto.ErrNotFound)
	}
	if err != nil {
		return nil, fmt.Errorf("getting channel by telegram channel id: %w", err)
	}

	return &ch, nil
}

func (r *repo) GetByID(ctx context.Context, id uuid.UUID) (*entity.Channel, error) {
	rows, err := r.db.Query(ctx, `
		SELECT * FROM channels
		WHERE id = $1 AND deleted_at IS NULL
	`, id)
	if err != nil {
		return nil, fmt.Errorf("getting channel by id: %w", err)
	}

	ch, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[entity.Channel])
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, fmt.Errorf("getting channel by id: %w", dto.ErrNotFound)
	}
	if err != nil {
		return nil, fmt.Errorf("getting channel by id: %w", err)
	}

	return &ch, nil
}

func (r *repo) SoftDelete(ctx context.Context, tgChannelID int64) error {
	tag, err := r.db.Exec(ctx, `
		UPDATE channels
		SET deleted_at = NOW()
		WHERE telegram_channel_id = $1 AND deleted_at IS NULL
	`, tgChannelID)
	if err != nil {
		return fmt.Errorf("soft deleting channel: %w", err)
	}

	if tag.RowsAffected() == 0 {
		return fmt.Errorf("soft deleting channel: %w", dto.ErrNotFound)
	}

	return nil
}

func (r *repo) CreateRole(
	ctx context.Context,
	channelID, userID uuid.UUID,
	role entity.ChannelRoleType,
) (*entity.ChannelRole, error) {
	rows, err := r.db.Query(ctx, `
		INSERT INTO channel_roles (channel_id, user_id, role)
		VALUES ($1, $2, $3)
		ON CONFLICT (channel_id, user_id) DO UPDATE SET role = EXCLUDED.role
		RETURNING channel_id, user_id, role, created_at
	`, channelID, userID, role)
	if err != nil {
		return nil, fmt.Errorf("creating channel role: %w", err)
	}

	cr, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[entity.ChannelRole])
	if err != nil {
		return nil, fmt.Errorf("creating channel role: %w", err)
	}

	return &cr, nil
}

func (r *repo) GetRolesByChannelID(
	ctx context.Context,
	channelID uuid.UUID,
) ([]entity.ChannelRole, error) {
	rows, err := r.db.Query(ctx, `
		SELECT channel_id, user_id, role, created_at
		FROM channel_roles
		WHERE channel_id = $1
	`, channelID)
	if err != nil {
		return nil, fmt.Errorf("getting channel roles: %w", err)
	}

	roles, err := pgx.CollectRows(rows, pgx.RowToStructByName[entity.ChannelRole])
	if err != nil {
		return nil, fmt.Errorf("getting channel roles: %w", err)
	}

	return roles, nil
}

func (r *repo) GetChannelsByUserID(
	ctx context.Context,
	userID uuid.UUID,
) ([]entity.Channel, error) {
	rows, err := r.db.Query(ctx, `
		SELECT c.* FROM channels c
		JOIN channel_roles cr ON c.id = cr.channel_id
		WHERE cr.user_id = $1 AND c.deleted_at IS NULL
	`, userID)
	if err != nil {
		return nil, fmt.Errorf("getting channels by user id: %w", err)
	}

	channels, err := pgx.CollectRows(rows, pgx.RowToStructByName[entity.Channel])
	if err != nil {
		return nil, fmt.Errorf("getting channels by user id: %w", err)
	}

	return channels, nil
}

func (r *repo) GetRole(
	ctx context.Context,
	channelID, userID uuid.UUID,
) (*entity.ChannelRole, error) {
	rows, err := r.db.Query(ctx, `
		SELECT channel_id, user_id, role, created_at
		FROM channel_roles
		WHERE channel_id = $1 AND user_id = $2
	`, channelID, userID)
	if err != nil {
		return nil, fmt.Errorf("getting channel role: %w", err)
	}

	cr, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[entity.ChannelRole])
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, fmt.Errorf("getting channel role: %w", dto.ErrNotFound)
	}
	if err != nil {
		return nil, fmt.Errorf("getting channel role: %w", err)
	}

	return &cr, nil
}

func (r *repo) DeleteRole(ctx context.Context, channelID, userID uuid.UUID) error {
	tag, err := r.db.Exec(ctx, `
		DELETE FROM channel_roles
		WHERE channel_id = $1 AND user_id = $2
	`, channelID, userID)
	if err != nil {
		return fmt.Errorf("deleting channel role: %w", err)
	}

	if tag.RowsAffected() == 0 {
		return fmt.Errorf("deleting channel role: %w", dto.ErrNotFound)
	}

	return nil
}

func (r *repo) UpdatePhoto(
	ctx context.Context,
	channelID uuid.UUID,
	smallFileID, bigFileID *string,
) error {
	tag, err := r.db.Exec(ctx, `
		UPDATE channels
		SET photo_small_file_id = $2, photo_big_file_id = $3
		WHERE id = $1 AND deleted_at IS NULL
	`, channelID, smallFileID, bigFileID)
	if err != nil {
		return fmt.Errorf("updating channel photo: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("updating channel photo: %w", dto.ErrNotFound)
	}
	return nil
}

func (r *repo) UpdateListing(ctx context.Context, channelID uuid.UUID, isListed bool) error {
	tag, err := r.db.Exec(ctx, `
		UPDATE channels
		SET is_listed = $2
		WHERE id = $1 AND deleted_at IS NULL
	`, channelID, isListed)
	if err != nil {
		return fmt.Errorf("updating listing status: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("updating listing status: %w", dto.ErrNotFound)
	}
	return nil
}

func (r *repo) CreateAdFormat(
	ctx context.Context,
	channelID uuid.UUID,
	formatType entity.AdFormatType,
	isNative bool,
	feedHours, topHours int,
	priceNanoTON int64,
) (*entity.ChannelAdFormat, error) {
	id, err := uuid.NewV7()
	if err != nil {
		return nil, fmt.Errorf("creating ad format: %w", err)
	}

	rows, err := r.db.Query(ctx, `
		INSERT INTO channel_ad_formats
			(id, channel_id, format_type, is_native, feed_hours, top_hours, price_nano_ton)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, channel_id, format_type, is_native, feed_hours, top_hours,
			price_nano_ton, created_at
	`, id, channelID, formatType, isNative, feedHours, topHours, priceNanoTON)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
			return nil, fmt.Errorf("creating ad format: %w", dto.ErrAdFormatExists)
		}
		return nil, fmt.Errorf("creating ad format: %w", err)
	}

	af, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[entity.ChannelAdFormat])
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
			return nil, fmt.Errorf("creating ad format: %w", dto.ErrAdFormatExists)
		}
		return nil, fmt.Errorf("creating ad format: %w", err)
	}

	return &af, nil
}

func (r *repo) GetAdFormatsByChannelID(
	ctx context.Context,
	channelID uuid.UUID,
) ([]entity.ChannelAdFormat, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, channel_id, format_type, is_native, feed_hours, top_hours,
			price_nano_ton, created_at
		FROM channel_ad_formats
		WHERE channel_id = $1
		ORDER BY created_at
	`, channelID)
	if err != nil {
		return nil, fmt.Errorf("getting ad formats: %w", err)
	}

	formats, err := pgx.CollectRows(rows, pgx.RowToStructByName[entity.ChannelAdFormat])
	if err != nil {
		return nil, fmt.Errorf("getting ad formats: %w", err)
	}

	return formats, nil
}

func (r *repo) GetAdFormatByID(
	ctx context.Context,
	formatID uuid.UUID,
) (*entity.ChannelAdFormat, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, channel_id, format_type, is_native, feed_hours, top_hours,
			price_nano_ton, created_at
		FROM channel_ad_formats
		WHERE id = $1
	`, formatID)
	if err != nil {
		return nil, fmt.Errorf("getting ad format: %w", err)
	}

	af, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[entity.ChannelAdFormat])
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, fmt.Errorf("getting ad format: %w", dto.ErrNotFound)
	}
	if err != nil {
		return nil, fmt.Errorf("getting ad format: %w", err)
	}

	return &af, nil
}

func (r *repo) DeleteAdFormat(ctx context.Context, formatID uuid.UUID) error {
	tag, err := r.db.Exec(ctx, `
		DELETE FROM channel_ad_formats WHERE id = $1
	`, formatID)
	if err != nil {
		return fmt.Errorf("deleting ad format: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("deleting ad format: %w", dto.ErrNotFound)
	}
	return nil
}
