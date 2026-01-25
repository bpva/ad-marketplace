package app

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"github.com/bpva/ad-marketplace/internal/dto"
	"github.com/bpva/ad-marketplace/internal/logx"
)

func (a *App) HandleListChannels() http.HandlerFunc {
	log := a.log.With(logx.Handler("/api/v1/channels"))

	return func(w http.ResponseWriter, r *http.Request) {
		channels, err := a.channel.GetUserChannels(r.Context())
		if err != nil {
			log.Error("failed to get channels", "error", err)
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(channels); err != nil {
			log.Error("failed to encode response", "error", err)
		}
	}
}

func (a *App) HandleGetChannel() http.HandlerFunc {
	log := a.log.With(logx.Handler("/api/v1/channels/{TgChannelID}"))

	return func(w http.ResponseWriter, r *http.Request) {
		TgChannelID, err := strconv.ParseInt(chi.URLParam(r, "TgChannelID"), 10, 64)
		if err != nil {
			http.Error(w, "invalid channel id", http.StatusBadRequest)
			return
		}

		channel, err := a.channel.GetChannel(r.Context(), TgChannelID)
		if errors.Is(err, dto.ErrNotFound) {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		if errors.Is(err, dto.ErrForbidden) {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}
		if err != nil {
			log.Error("failed to get channel", "error", err)
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(channel); err != nil {
			log.Error("failed to encode response", "error", err)
		}
	}
}

func (a *App) HandleGetChannelAdmins() http.HandlerFunc {
	log := a.log.With(logx.Handler("/api/v1/channels/{TgChannelID}/admins"))

	return func(w http.ResponseWriter, r *http.Request) {
		TgChannelID, err := strconv.ParseInt(chi.URLParam(r, "TgChannelID"), 10, 64)
		if err != nil {
			http.Error(w, "invalid channel id", http.StatusBadRequest)
			return
		}

		admins, err := a.channel.GetChannelAdmins(r.Context(), TgChannelID)
		if errors.Is(err, dto.ErrNotFound) {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		if errors.Is(err, dto.ErrForbidden) {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}
		if err != nil {
			log.Error("failed to get channel admins", "error", err)
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(admins); err != nil {
			log.Error("failed to encode response", "error", err)
		}
	}
}

func (a *App) HandleGetChannelManagers() http.HandlerFunc {
	log := a.log.With(logx.Handler("/api/v1/channels/{TgChannelID}/managers"))

	return func(w http.ResponseWriter, r *http.Request) {
		TgChannelID, err := strconv.ParseInt(chi.URLParam(r, "TgChannelID"), 10, 64)
		if err != nil {
			http.Error(w, "invalid channel id", http.StatusBadRequest)
			return
		}

		managers, err := a.channel.GetChannelManagers(r.Context(), TgChannelID)
		if errors.Is(err, dto.ErrNotFound) {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		if errors.Is(err, dto.ErrForbidden) {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}
		if err != nil {
			log.Error("failed to get channel managers", "error", err)
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(managers); err != nil {
			log.Error("failed to encode response", "error", err)
		}
	}
}

func (a *App) HandleAddManager() http.HandlerFunc {
	log := a.log.With(logx.Handler("/api/v1/channels/{TgChannelID}/managers"))

	return func(w http.ResponseWriter, r *http.Request) {
		TgChannelID, err := strconv.ParseInt(chi.URLParam(r, "TgChannelID"), 10, 64)
		if err != nil {
			http.Error(w, "invalid channel id", http.StatusBadRequest)
			return
		}

		var req dto.AddManagerRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}

		if req.TgID == 0 {
			http.Error(w, "telegram_id required", http.StatusBadRequest)
			return
		}

		err = a.channel.AddManager(r.Context(), TgChannelID, req.TgID)
		if errors.Is(err, dto.ErrNotFound) {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		if errors.Is(err, dto.ErrForbidden) {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}
		if errors.Is(err, dto.ErrUserNotRegistered) {
			http.Error(w, "user not registered", http.StatusBadRequest)
			return
		}
		if err != nil {
			log.Error("failed to add manager", "error", err)
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}

func (a *App) HandleRemoveManager() http.HandlerFunc {
	log := a.log.With(logx.Handler("/api/v1/channels/{TgChannelID}/managers/{tgID}"))

	return func(w http.ResponseWriter, r *http.Request) {
		TgChannelID, err := strconv.ParseInt(chi.URLParam(r, "TgChannelID"), 10, 64)
		if err != nil {
			http.Error(w, "invalid channel id", http.StatusBadRequest)
			return
		}

		tgID, err := strconv.ParseInt(chi.URLParam(r, "tgID"), 10, 64)
		if err != nil {
			http.Error(w, "invalid telegram id", http.StatusBadRequest)
			return
		}

		err = a.channel.RemoveManager(r.Context(), TgChannelID, tgID)
		if errors.Is(err, dto.ErrNotFound) {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		if errors.Is(err, dto.ErrForbidden) {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}
		if errors.Is(err, dto.ErrCannotRemoveOwner) {
			http.Error(w, "cannot remove owner", http.StatusBadRequest)
			return
		}
		if err != nil {
			log.Error("failed to remove manager", "error", err)
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}
