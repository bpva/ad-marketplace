package tools

import (
	"context"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/bpva/ad-marketplace/internal/dto"
	"github.com/bpva/ad-marketplace/internal/entity"
)

type Tools struct {
	pool      *pgxpool.Pool
	jwtSecret []byte
}

func New(pool *pgxpool.Pool, jwtSecret string) *Tools {
	return &Tools{pool: pool, jwtSecret: []byte(jwtSecret)}
}

func (t *Tools) CreateUser(
	ctx context.Context,
	tgID int64,
	name string,
) (*entity.User, error) {
	id, err := uuid.NewV7()
	if err != nil {
		return nil, err
	}

	var user entity.User
	err = t.pool.QueryRow(ctx, `
		INSERT INTO users (id, telegram_id, name)
		VALUES ($1, $2, $3)
		RETURNING id, telegram_id, name, created_at, deleted_at
	`, id, tgID, name).Scan(
		&user.ID,
		&user.TgID,
		&user.Name,
		&user.CreatedAt,
		&user.DeletedAt,
	)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (t *Tools) GenerateToken(user *entity.User) (string, error) {
	claims := dto.Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   user.ID.String(),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
		TgID: user.TgID,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(t.jwtSecret)
}

func (t *Tools) GenerateExpiredToken(user *entity.User) (string, error) {
	claims := dto.Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   user.ID.String(),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(-1 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now().Add(-2 * time.Hour)),
		},
		TgID: user.TgID,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(t.jwtSecret)
}

func (t *Tools) CreateChannel(
	ctx context.Context,
	tgChannelID int64,
	title string,
	username *string,
) (*entity.Channel, error) {
	id, err := uuid.NewV7()
	if err != nil {
		return nil, err
	}

	var channel entity.Channel
	err = t.pool.QueryRow(ctx, `
		INSERT INTO channels (id, telegram_channel_id, title, username)
		VALUES ($1, $2, $3, $4)
		RETURNING id, telegram_channel_id, title, username, is_listed, created_at
	`, id, tgChannelID, title, username).Scan(
		&channel.ID,
		&channel.TgChannelID,
		&channel.Title,
		&channel.Username,
		&channel.IsListed,
		&channel.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &channel, nil
}

func (t *Tools) CreateChannelRole(
	ctx context.Context,
	channelID, userID uuid.UUID,
	role entity.ChannelRoleType,
) (*entity.ChannelRole, error) {
	var cr entity.ChannelRole
	err := t.pool.QueryRow(ctx, `
		INSERT INTO channel_roles (channel_id, user_id, role)
		VALUES ($1, $2, $3)
		RETURNING channel_id, user_id, role, created_at
	`, channelID, userID, role).Scan(
		&cr.ChannelID,
		&cr.UserID,
		&cr.Role,
		&cr.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &cr, nil
}

func (t *Tools) GetChannelByTgID(
	ctx context.Context,
	tgChannelID int64,
) (*entity.Channel, error) {
	var channel entity.Channel
	err := t.pool.QueryRow(ctx, `
		SELECT id, telegram_channel_id, title, username, is_listed, created_at
		FROM channels
		WHERE telegram_channel_id = $1
	`, tgChannelID).Scan(
		&channel.ID,
		&channel.TgChannelID,
		&channel.Title,
		&channel.Username,
		&channel.IsListed,
		&channel.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &channel, nil
}

func (t *Tools) IsChannelSoftDeleted(ctx context.Context, tgChannelID int64) (bool, error) {
	var deletedAt *time.Time
	err := t.pool.QueryRow(ctx, `
		SELECT deleted_at FROM channels WHERE telegram_channel_id = $1
	`, tgChannelID).Scan(&deletedAt)
	if err != nil {
		return false, err
	}
	return deletedAt != nil, nil
}

func (t *Tools) GetChannelRolesByChannelID(
	ctx context.Context,
	channelID uuid.UUID,
) ([]entity.ChannelRole, error) {
	rows, err := t.pool.Query(ctx, `
		SELECT channel_id, user_id, role, created_at
		FROM channel_roles
		WHERE channel_id = $1
	`, channelID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var roles []entity.ChannelRole
	for rows.Next() {
		var cr entity.ChannelRole
		if err := rows.Scan(&cr.ChannelID, &cr.UserID, &cr.Role, &cr.CreatedAt); err != nil {
			return nil, err
		}
		roles = append(roles, cr)
	}
	return roles, rows.Err()
}

func (t *Tools) SoftDeleteChannel(ctx context.Context, tgChannelID int64) error {
	_, err := t.pool.Exec(ctx, `
		UPDATE channels SET deleted_at = NOW() WHERE telegram_channel_id = $1
	`, tgChannelID)
	return err
}

func (t *Tools) GetUserByTgID(ctx context.Context, tgID int64) (*entity.User, error) {
	var user entity.User
	err := t.pool.QueryRow(ctx, `
		SELECT id, telegram_id, name, created_at, deleted_at
		FROM users
		WHERE telegram_id = $1
	`, tgID).Scan(
		&user.ID,
		&user.TgID,
		&user.Name,
		&user.CreatedAt,
		&user.DeletedAt,
	)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (t *Tools) CreatePost(
	ctx context.Context,
	postType entity.PostType,
	externalID uuid.UUID,
	version *int,
	mediaGroupID *string,
	text *string,
	entities []byte,
	mediaType *entity.MediaType,
	mediaFileID *string,
) (*entity.Post, error) {
	id, err := uuid.NewV7()
	if err != nil {
		return nil, err
	}

	var p entity.Post
	err = t.pool.QueryRow(ctx, `
		INSERT INTO posts (id, type, external_id, version, media_group_id, text, entities, media_type, media_file_id)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id, type, external_id, version, name, media_group_id, text, entities, media_type, media_file_id,
			has_media_spoiler, show_caption_above_media, created_at, deleted_at
	`, id, postType, externalID, version, mediaGroupID, text, entities, mediaType, mediaFileID).Scan(
		&p.ID, &p.Type, &p.ExternalID, &p.Version, &p.Name, &p.MediaGroupID, &p.Text, &p.Entities,
		&p.MediaType, &p.MediaFileID, &p.HasMediaSpoiler,
		&p.ShowCaptionAboveMedia, &p.CreatedAt, &p.DeletedAt,
	)
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (t *Tools) GetTemplatesByOwner(
	ctx context.Context,
	ownerID uuid.UUID,
) ([]entity.Post, error) {
	rows, err := t.pool.Query(ctx, `
		SELECT id, type, external_id, version, name, media_group_id, text, entities, media_type, media_file_id,
			has_media_spoiler, show_caption_above_media, created_at, deleted_at
		FROM posts
		WHERE type = 'template' AND external_id = $1 AND deleted_at IS NULL
		ORDER BY created_at DESC
	`, ownerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var posts []entity.Post
	for rows.Next() {
		var p entity.Post
		if err := rows.Scan(
			&p.ID,
			&p.Type,
			&p.ExternalID,
			&p.Version,
			&p.Name,
			&p.MediaGroupID,
			&p.Text,
			&p.Entities,
			&p.MediaType,
			&p.MediaFileID,
			&p.HasMediaSpoiler,
			&p.ShowCaptionAboveMedia,
			&p.CreatedAt,
			&p.DeletedAt,
		); err != nil {
			return nil, err
		}
		posts = append(posts, p)
	}
	return posts, rows.Err()
}

func (t *Tools) SetWalletAddress(ctx context.Context, userID uuid.UUID, address string) error {
	_, err := t.pool.Exec(ctx, `
		UPDATE users SET wallet_address = $2 WHERE id = $1
	`, userID, address)
	return err
}

func (t *Tools) UpdateChannelListing(
	ctx context.Context,
	channelID uuid.UUID,
	isListed bool,
) error {
	_, err := t.pool.Exec(ctx, `
		UPDATE channels SET is_listed = $2 WHERE id = $1
	`, channelID, isListed)
	return err
}

func (t *Tools) CreateAdFormat(
	ctx context.Context,
	channelID uuid.UUID,
	formatType entity.AdFormatType,
	isNative bool,
	feedHours, topHours int,
	priceNanoTON int64,
) (*entity.ChannelAdFormat, error) {
	id, err := uuid.NewV7()
	if err != nil {
		return nil, err
	}

	var af entity.ChannelAdFormat
	err = t.pool.QueryRow(ctx, `
		INSERT INTO channel_ad_formats
			(id, channel_id, format_type, is_native, feed_hours, top_hours, price_nano_ton)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, channel_id, format_type, is_native, feed_hours, top_hours,
			price_nano_ton, created_at
	`, id, channelID, formatType, isNative, feedHours, topHours, priceNanoTON).Scan(
		&af.ID,
		&af.ChannelID,
		&af.FormatType,
		&af.IsNative,
		&af.FeedHours,
		&af.TopHours,
		&af.PriceNanoTON,
		&af.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &af, nil
}

func (t *Tools) CreateDeal(
	ctx context.Context,
	channelID, advertiserID uuid.UUID,
	status entity.DealStatus,
	scheduledAt time.Time,
	formatType entity.AdFormatType,
	isNative bool,
	feedHours, topHours int,
	priceNanoTON int64,
) (*entity.Deal, error) {
	id, err := uuid.NewV7()
	if err != nil {
		return nil, err
	}

	rows, err := t.pool.Query(ctx, `
		INSERT INTO deals (
			id, channel_id, advertiser_id, status, scheduled_at,
			format_type, is_native, feed_hours, top_hours, price_nano_ton
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING
			id, channel_id, advertiser_id, status, scheduled_at,
			publisher_note, escrow_wallet_address, advertiser_wallet_address,
			payout_wallet_address, format_type, is_native, feed_hours,
			top_hours, price_nano_ton, posted_message_ids,
			paid_at, payment_tx_hash, posted_at, release_tx_hash,
			refund_tx_hash, created_at, updated_at
	`, id, channelID, advertiserID, status, scheduledAt,
		formatType, isNative, feedHours, topHours, priceNanoTON)
	if err != nil {
		return nil, err
	}

	d, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[entity.Deal])
	if err != nil {
		return nil, err
	}

	return &d, nil
}

func (t *Tools) GetAdFormatsByChannelID(
	ctx context.Context,
	channelID uuid.UUID,
) ([]entity.ChannelAdFormat, error) {
	rows, err := t.pool.Query(ctx, `
		SELECT id, channel_id, format_type, is_native, feed_hours, top_hours,
			price_nano_ton, created_at
		FROM channel_ad_formats
		WHERE channel_id = $1
	`, channelID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var formats []entity.ChannelAdFormat
	for rows.Next() {
		var af entity.ChannelAdFormat
		if err := rows.Scan(
			&af.ID, &af.ChannelID, &af.FormatType, &af.IsNative,
			&af.FeedHours, &af.TopHours, &af.PriceNanoTON, &af.CreatedAt,
		); err != nil {
			return nil, err
		}
		formats = append(formats, af)
	}
	return formats, rows.Err()
}
