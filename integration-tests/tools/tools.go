package tools

import (
	"context"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
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
	telegramID int64,
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
	`, id, telegramID, name).Scan(
		&user.ID,
		&user.TelegramID,
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
		TelegramID: user.TelegramID,
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
		TelegramID: user.TelegramID,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(t.jwtSecret)
}

func (t *Tools) CreateChannel(
	ctx context.Context,
	telegramChannelID int64,
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
		RETURNING id, telegram_channel_id, title, username, created_at, deleted_at
	`, id, telegramChannelID, title, username).Scan(
		&channel.ID,
		&channel.TelegramChannelID,
		&channel.Title,
		&channel.Username,
		&channel.CreatedAt,
		&channel.DeletedAt,
	)
	if err != nil {
		return nil, err
	}
	return &channel, nil
}

func (t *Tools) CreateChannelRole(
	ctx context.Context,
	channelID, userID uuid.UUID,
	role string,
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

func (t *Tools) GetChannelByTelegramID(
	ctx context.Context,
	telegramChannelID int64,
) (*entity.Channel, error) {
	var channel entity.Channel
	err := t.pool.QueryRow(ctx, `
		SELECT id, telegram_channel_id, title, username, created_at, deleted_at
		FROM channels
		WHERE telegram_channel_id = $1
	`, telegramChannelID).Scan(
		&channel.ID,
		&channel.TelegramChannelID,
		&channel.Title,
		&channel.Username,
		&channel.CreatedAt,
		&channel.DeletedAt,
	)
	if err != nil {
		return nil, err
	}
	return &channel, nil
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

func (t *Tools) SoftDeleteChannel(ctx context.Context, telegramChannelID int64) error {
	_, err := t.pool.Exec(ctx, `
		UPDATE channels SET deleted_at = NOW() WHERE telegram_channel_id = $1
	`, telegramChannelID)
	return err
}

func (t *Tools) GetUserByTelegramID(ctx context.Context, telegramID int64) (*entity.User, error) {
	var user entity.User
	err := t.pool.QueryRow(ctx, `
		SELECT id, telegram_id, name, created_at, deleted_at
		FROM users
		WHERE telegram_id = $1
	`, telegramID).Scan(
		&user.ID,
		&user.TelegramID,
		&user.Name,
		&user.CreatedAt,
		&user.DeletedAt,
	)
	if err != nil {
		return nil, err
	}
	return &user, nil
}
