package app

import (
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/bpva/ad-marketplace/internal/dto"
	"github.com/bpva/ad-marketplace/internal/http/respond"
	"github.com/bpva/ad-marketplace/internal/logx"
)

// HandleGetChannelInfo returns channel info and stats summary
//
//	@Summary		Get channel info
//	@Tags			channels
//	@Produce		json
//	@Param			TgChannelID	path		int	true	"Telegram channel ID"
//	@Success		200			{object}	dto.ChannelInfoResponse
//	@Failure		400			{object}	dto.ErrorResponse
//	@Failure		404			{object}	dto.ErrorResponse
//	@Router			/channels/{TgChannelID}/info [get]
func (a *App) HandleGetChannelInfo() http.HandlerFunc {
	log := a.log.With(logx.Handler("/api/v1/channels/{TgChannelID}/info"))

	return func(w http.ResponseWriter, r *http.Request) {
		tgChannelID, err := strconv.ParseInt(chi.URLParam(r, "TgChannelID"), 10, 64)
		if err != nil {
			respond.Err(w, log, dto.ErrInvalidChannelID)
			return
		}

		info, err := a.stats.GetInfo(r.Context(), tgChannelID)
		if err != nil {
			respond.Err(w, log, err)
			return
		}

		respond.OK(w, info)
	}
}

// HandleGetChannelStats returns channel historical stats
//
//	@Summary		Get channel historical stats
//	@Tags			channels
//	@Produce		json
//	@Param			TgChannelID	path		int		true	"Telegram channel ID"
//	@Param			from		query		string	false	"Start date (YYYY-MM-DD)"
//	@Param			to			query		string	false	"End date (YYYY-MM-DD)"
//	@Success		200			{object}	dto.ChannelHistoricalStatsResponse
//	@Failure		400			{object}	dto.ErrorResponse
//	@Failure		404			{object}	dto.ErrorResponse
//	@Router			/channels/{TgChannelID}/stats [get]
func (a *App) HandleGetChannelStats() http.HandlerFunc {
	log := a.log.With(logx.Handler("/api/v1/channels/{TgChannelID}/stats"))

	return func(w http.ResponseWriter, r *http.Request) {
		tgChannelID, err := strconv.ParseInt(chi.URLParam(r, "TgChannelID"), 10, 64)
		if err != nil {
			respond.Err(w, log, dto.ErrInvalidChannelID)
			return
		}

		var from, to time.Time
		if v := r.URL.Query().Get("from"); v != "" {
			from, err = time.Parse("2006-01-02", v)
			if err != nil {
				respond.Err(w, log, dto.ErrBadRequest)
				return
			}
		}
		if v := r.URL.Query().Get("to"); v != "" {
			to, err = time.Parse("2006-01-02", v)
			if err != nil {
				respond.Err(w, log, dto.ErrBadRequest)
				return
			}
		}

		stats, err := a.stats.GetHistory(r.Context(), tgChannelID, from, to)
		if err != nil {
			respond.Err(w, log, err)
			return
		}

		respond.OK(w, stats)
	}
}
