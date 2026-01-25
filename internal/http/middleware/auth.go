package middleware

import (
	"log/slog"
	"net/http"
	"strings"

	"github.com/google/uuid"

	"github.com/bpva/ad-marketplace/internal/dto"
	"github.com/bpva/ad-marketplace/internal/http/respond"
	"github.com/bpva/ad-marketplace/internal/logx"
)

type TokenValidator interface {
	ValidateToken(tokenString string) (*dto.Claims, error)
}

func Auth(validator TokenValidator, log *slog.Logger) func(http.Handler) http.Handler {
	log = log.With(logx.Handler("auth middleware"))

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			header := r.Header.Get("Authorization")
			if header == "" {
				respond.Err(w, log, dto.ErrUnauthorized)
				return
			}

			token := strings.TrimPrefix(header, "Bearer ")
			if token == header {
				respond.Err(w, log, dto.ErrUnauthorized)
				return
			}

			claims, err := validator.ValidateToken(token)
			if err != nil {
				respond.Err(w, log, dto.ErrUnauthorized.Wrap(err))
				return
			}

			userID, err := uuid.Parse(claims.Subject)
			if err != nil {
				respond.Err(w, log, dto.ErrUnauthorized.Wrap(err))
				return
			}

			userCtx := dto.UserContext{
				ID:   userID,
				TgID: claims.TgID,
			}

			ctx := dto.ContextWithUser(r.Context(), userCtx)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
