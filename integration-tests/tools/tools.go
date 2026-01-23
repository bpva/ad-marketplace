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
