package app

import (
	"encoding/json"
	"net/http"

	"github.com/bpva/ad-marketplace/internal/dto"
	"github.com/bpva/ad-marketplace/internal/http/respond"
	"github.com/bpva/ad-marketplace/internal/logx"
)

func (a *App) HandleGetProfile() http.HandlerFunc {
	log := a.log.With(logx.Handler("/api/v1/user/profile"))

	return func(w http.ResponseWriter, r *http.Request) {
		profile, err := a.user.GetProfile(r.Context())
		if err != nil {
			respond.Err(w, log, err)
			return
		}

		respond.OK(w, profile)
	}
}

func (a *App) HandleUpdateName() http.HandlerFunc {
	log := a.log.With(logx.Handler("/api/v1/user/name"))

	return func(w http.ResponseWriter, r *http.Request) {
		var req dto.UpdateNameRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			respond.Err(w, log, dto.ErrBadRequest)
			return
		}

		if err := a.user.UpdateName(r.Context(), req.Name); err != nil {
			respond.Err(w, log, err)
			return
		}

		respond.NoContent(w)
	}
}

func (a *App) HandleUpdateSettings() http.HandlerFunc {
	log := a.log.With(logx.Handler("/api/v1/user/settings"))

	return func(w http.ResponseWriter, r *http.Request) {
		var req dto.UpdateSettingsRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			respond.Err(w, log, dto.ErrBadRequest)
			return
		}

		if err := a.user.UpdateSettings(r.Context(), req); err != nil {
			respond.Err(w, log, err)
			return
		}

		respond.NoContent(w)
	}
}
