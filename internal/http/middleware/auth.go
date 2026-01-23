package middleware

import (
	"net/http"
	"strings"

	"github.com/google/uuid"

	"github.com/bpva/ad-marketplace/internal/dto"
)

type TokenValidator interface {
	ValidateToken(tokenString string) (*dto.Claims, error)
}

func Auth(validator TokenValidator) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			header := r.Header.Get("Authorization")
			if header == "" {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}

			token := strings.TrimPrefix(header, "Bearer ")
			if token == header {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}

			claims, err := validator.ValidateToken(token)
			if err != nil {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}

			userID, err := uuid.Parse(claims.Subject)
			if err != nil {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}

			userCtx := dto.UserContext{
				ID:         userID,
				TelegramID: claims.TelegramID,
			}

			ctx := dto.ContextWithUser(r.Context(), userCtx)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
