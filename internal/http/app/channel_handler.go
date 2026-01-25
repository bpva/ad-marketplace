package app

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"github.com/bpva/ad-marketplace/internal/dto"
	"github.com/bpva/ad-marketplace/internal/http/respond"
	"github.com/bpva/ad-marketplace/internal/logx"
)

func (a *App) HandleListChannels() http.HandlerFunc {
	log := a.log.With(logx.Handler("/api/v1/channels"))

	return func(w http.ResponseWriter, r *http.Request) {
		channels, err := a.channel.GetUserChannels(r.Context())
		if err != nil {
			respond.Err(w, log, err)
			return
		}

		respond.OK(w, channels)
	}
}

func (a *App) HandleGetChannel() http.HandlerFunc {
	log := a.log.With(logx.Handler("/api/v1/channels/{TgChannelID}"))

	return func(w http.ResponseWriter, r *http.Request) {
		TgChannelID, err := strconv.ParseInt(chi.URLParam(r, "TgChannelID"), 10, 64)
		if err != nil {
			respond.Err(w, log, dto.ErrInvalidChannelID)
			return
		}

		channel, err := a.channel.GetChannel(r.Context(), TgChannelID)
		if err != nil {
			respond.Err(w, log, err)
			return
		}

		respond.OK(w, channel)
	}
}

func (a *App) HandleGetChannelAdmins() http.HandlerFunc {
	log := a.log.With(logx.Handler("/api/v1/channels/{TgChannelID}/admins"))

	return func(w http.ResponseWriter, r *http.Request) {
		TgChannelID, err := strconv.ParseInt(chi.URLParam(r, "TgChannelID"), 10, 64)
		if err != nil {
			respond.Err(w, log, dto.ErrInvalidChannelID)
			return
		}

		admins, err := a.channel.GetChannelAdmins(r.Context(), TgChannelID)
		if err != nil {
			respond.Err(w, log, err)
			return
		}

		respond.OK(w, admins)
	}
}

func (a *App) HandleGetChannelManagers() http.HandlerFunc {
	log := a.log.With(logx.Handler("/api/v1/channels/{TgChannelID}/managers"))

	return func(w http.ResponseWriter, r *http.Request) {
		TgChannelID, err := strconv.ParseInt(chi.URLParam(r, "TgChannelID"), 10, 64)
		if err != nil {
			respond.Err(w, log, dto.ErrInvalidChannelID)
			return
		}

		managers, err := a.channel.GetChannelManagers(r.Context(), TgChannelID)
		if err != nil {
			respond.Err(w, log, err)
			return
		}

		respond.OK(w, managers)
	}
}

func (a *App) HandleAddManager() http.HandlerFunc {
	log := a.log.With(logx.Handler("/api/v1/channels/{TgChannelID}/managers"))

	return func(w http.ResponseWriter, r *http.Request) {
		TgChannelID, err := strconv.ParseInt(chi.URLParam(r, "TgChannelID"), 10, 64)
		if err != nil {
			respond.Err(w, log, dto.ErrInvalidChannelID)
			return
		}

		var req dto.AddManagerRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			respond.Err(w, log, dto.ErrBadRequest)
			return
		}

		if req.TgID == 0 {
			respond.Err(w, log, dto.ErrTelegramIDRequired)
			return
		}

		err = a.channel.AddManager(r.Context(), TgChannelID, req.TgID)
		if err != nil {
			respond.Err(w, log, err)
			return
		}

		respond.NoContent(w)
	}
}

func (a *App) HandleRemoveManager() http.HandlerFunc {
	log := a.log.With(logx.Handler("/api/v1/channels/{TgChannelID}/managers/{tgID}"))

	return func(w http.ResponseWriter, r *http.Request) {
		TgChannelID, err := strconv.ParseInt(chi.URLParam(r, "TgChannelID"), 10, 64)
		if err != nil {
			respond.Err(w, log, dto.ErrInvalidChannelID)
			return
		}

		tgID, err := strconv.ParseInt(chi.URLParam(r, "tgID"), 10, 64)
		if err != nil {
			respond.Err(w, log, dto.ErrInvalidTelegramID)
			return
		}

		err = a.channel.RemoveManager(r.Context(), TgChannelID, tgID)
		if err != nil {
			respond.Err(w, log, err)
			return
		}

		respond.NoContent(w)
	}
}
