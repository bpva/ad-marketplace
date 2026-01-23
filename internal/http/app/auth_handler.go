package app

import (
	"encoding/json"
	"net/http"

	"github.com/bpva/ad-marketplace/internal/dto"
	"github.com/bpva/ad-marketplace/internal/entity"
	"github.com/bpva/ad-marketplace/internal/logx"
)

func userToResponse(u *entity.User) dto.UserResponse {
	return dto.UserResponse{
		ID:         u.ID.String(),
		TelegramID: u.TelegramID,
		Name:       u.Name,
	}
}

func (a *App) HandleAuth() http.HandlerFunc {
	log := a.log.With(logx.Handler("/api/v1/auth"))

	return func(w http.ResponseWriter, r *http.Request) {
		var req dto.AuthRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			log.Error("failed to decode request", "error", err)
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}

		token, user, err := a.auth.Authenticate(r.Context(), req.InitData)
		if err != nil {
			log.Error("failed to authenticate", "error", err)
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		resp := dto.AuthResponse{
			Token: token,
			User:  userToResponse(user),
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			log.Error("failed to encode response", "error", err)
		}
	}
}

func (a *App) HandleMe() http.HandlerFunc {
	log := a.log.With(logx.Handler("/api/v1/me"))

	return func(w http.ResponseWriter, r *http.Request) {
		userCtx, ok := dto.UserFromContext(r.Context())
		if !ok {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		user, err := a.auth.GetUserByID(r.Context(), userCtx.ID)
		if err != nil {
			log.Error("failed to get user", "error", err)
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(userToResponse(user)); err != nil {
			log.Error("failed to encode response", "error", err)
		}
	}
}
