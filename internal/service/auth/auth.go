package auth

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/url"
	"sort"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"

	"github.com/bpva/ad-marketplace/internal/dto"
	"github.com/bpva/ad-marketplace/internal/entity"
	"github.com/bpva/ad-marketplace/internal/logx"
)

var (
	ErrInvalidInitData = errors.New("invalid init data")
	ErrInvalidToken    = errors.New("invalid token")
)

type UserRepository interface {
	GetByTelegramID(ctx context.Context, telegramID int64) (*entity.User, error)
	GetByID(ctx context.Context, id uuid.UUID) (*entity.User, error)
	Create(ctx context.Context, telegramID int64, name string) (*entity.User, error)
}

type service struct {
	users     UserRepository
	botToken  string
	jwtSecret []byte
	log       *slog.Logger
}

func New(users UserRepository, botToken, jwtSecret string, log *slog.Logger) *service {
	log = log.With(logx.Service("AuthService"))
	return &service{
		users:     users,
		botToken:  botToken,
		jwtSecret: []byte(jwtSecret),
		log:       log,
	}
}

func (s *service) Authenticate(ctx context.Context, initData string) (string, *entity.User, error) {
	tgUser, err := s.validateInitData(initData)
	if err != nil {
		return "", nil, fmt.Errorf("validate init data: %w", err)
	}

	user, err := s.users.GetByTelegramID(ctx, tgUser.ID)
	if errors.Is(err, dto.ErrNotFound) {
		name := tgUser.FirstName
		if tgUser.LastName != "" {
			name += " " + tgUser.LastName
		}
		user, err = s.users.Create(ctx, tgUser.ID, name)
		if err != nil {
			return "", nil, fmt.Errorf("create user: %w", err)
		}
		s.log.Info("user created", "user_id", user.ID, "telegram_id", tgUser.ID)
	} else if err != nil {
		return "", nil, fmt.Errorf("get user: %w", err)
	}

	token, err := s.generateToken(user)
	if err != nil {
		return "", nil, fmt.Errorf("generate token: %w", err)
	}

	return token, user, nil
}

func (s *service) ValidateToken(tokenString string) (*dto.Claims, error) {
	token, err := jwt.ParseWithClaims(
		tokenString,
		&dto.Claims{},
		func(token *jwt.Token) (any, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return s.jwtSecret, nil
		},
	)
	if err != nil {
		return nil, ErrInvalidToken
	}

	claims, ok := token.Claims.(*dto.Claims)
	if !ok || !token.Valid {
		return nil, ErrInvalidToken
	}

	return claims, nil
}

func (s *service) validateInitData(initData string) (*dto.TelegramUser, error) {
	values, err := url.ParseQuery(initData)
	if err != nil {
		return nil, ErrInvalidInitData
	}

	hash := values.Get("hash")
	if hash == "" {
		return nil, ErrInvalidInitData
	}

	var keys []string
	for k := range values {
		if k != "hash" {
			keys = append(keys, k)
		}
	}
	sort.Strings(keys)

	var dataCheckParts []string
	for _, k := range keys {
		dataCheckParts = append(dataCheckParts, k+"="+values.Get(k))
	}
	dataCheckString := strings.Join(dataCheckParts, "\n")

	secretKey := hmac.New(sha256.New, []byte("WebAppData"))
	secretKey.Write([]byte(s.botToken))

	h := hmac.New(sha256.New, secretKey.Sum(nil))
	h.Write([]byte(dataCheckString))
	calculatedHash := hex.EncodeToString(h.Sum(nil))

	if calculatedHash != hash {
		return nil, ErrInvalidInitData
	}

	userJSON := values.Get("user")
	if userJSON == "" {
		return nil, ErrInvalidInitData
	}

	var tgUser dto.TelegramUser
	if err := json.Unmarshal([]byte(userJSON), &tgUser); err != nil {
		return nil, ErrInvalidInitData
	}

	return &tgUser, nil
}

func (s *service) generateToken(user *entity.User) (string, error) {
	claims := dto.Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   user.ID.String(),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
		TelegramID: user.TelegramID,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.jwtSecret)
}

func (s *service) GetUserByID(ctx context.Context, userID uuid.UUID) (*entity.User, error) {
	return s.users.GetByID(ctx, userID)
}
