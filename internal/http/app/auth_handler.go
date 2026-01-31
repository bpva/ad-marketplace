package app

import (
	"encoding/json"
	"net/http"

	"github.com/bpva/ad-marketplace/internal/dto"
	"github.com/bpva/ad-marketplace/internal/http/respond"
	"github.com/bpva/ad-marketplace/internal/logx"
)

// HandleAuth authenticates user via Telegram init data
//
//	@Summary		Authenticate user
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Param			request	body		dto.AuthRequest	true	"Telegram init data"
//	@Success		200		{object}	dto.AuthResponse
//	@Failure		401		{object}	dto.ErrorResponse
//	@Router			/auth [post]
func (a *App) HandleAuth() http.HandlerFunc {
	log := a.log.With(logx.Handler("/api/v1/auth"))

	return func(w http.ResponseWriter, r *http.Request) {
		var req dto.AuthRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			respond.Err(w, log, dto.ErrBadRequest)
			return
		}

		token, user, err := a.auth.Authenticate(r.Context(), req.InitData)
		if err != nil {
			respond.Err(w, log, dto.ErrUnauthorized.Wrap(err))
			return
		}

		respond.OK(w, dto.AuthResponse{
			Token: token,
			User:  dto.UserResponseFrom(user),
		})
	}
}

// HandleMe returns current authenticated user
//
//	@Summary		Get current user
//	@Tags			auth
//	@Produce		json
//	@Security		BearerAuth
//	@Success		200	{object}	dto.UserResponse
//	@Failure		401	{object}	dto.ErrorResponse
//	@Router			/me [get]
func (a *App) HandleMe() http.HandlerFunc {
	log := a.log.With(logx.Handler("/api/v1/me"))

	return func(w http.ResponseWriter, r *http.Request) {
		userCtx, ok := dto.UserFromContext(r.Context())
		if !ok {
			respond.Err(w, log, dto.ErrUnauthorized)
			return
		}

		user, err := a.auth.GetUserByID(r.Context(), userCtx.ID)
		if err != nil {
			respond.Err(w, log, err)
			return
		}

		respond.OK(w, dto.UserResponseFrom(user))
	}
}
