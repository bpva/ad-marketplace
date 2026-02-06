package app

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/bpva/ad-marketplace/internal/dto"
	"github.com/bpva/ad-marketplace/internal/http/bind"
	"github.com/bpva/ad-marketplace/internal/http/respond"
	"github.com/bpva/ad-marketplace/internal/logx"
)

// HandleListChannels returns all channels user has access to
//
//	@Summary		List user channels
//	@Tags			channels
//	@Produce		json
//	@Security		BearerAuth
//	@Success		200	{object}	dto.ChannelsResponse
//	@Failure		401	{object}	dto.ErrorResponse
//	@Router			/channels [get]
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

// HandleGetChannel returns channel details
//
//	@Summary		Get channel
//	@Tags			channels
//	@Produce		json
//	@Security		BearerAuth
//	@Param			TgChannelID	path		int	true	"Telegram channel ID"
//	@Success		200			{object}	dto.ChannelResponse
//	@Failure		400			{object}	dto.ErrorResponse
//	@Failure		401			{object}	dto.ErrorResponse
//	@Failure		404			{object}	dto.ErrorResponse
//	@Router			/channels/{TgChannelID} [get]
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

// HandleGetChannelAdmins returns channel admins from Telegram
//
//	@Summary		Get channel admins
//	@Tags			channels
//	@Produce		json
//	@Security		BearerAuth
//	@Param			TgChannelID	path		int	true	"Telegram channel ID"
//	@Success		200			{object}	dto.ChannelAdminsResponse
//	@Failure		400			{object}	dto.ErrorResponse
//	@Failure		401			{object}	dto.ErrorResponse
//	@Failure		404			{object}	dto.ErrorResponse
//	@Router			/channels/{TgChannelID}/admins [get]
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

// HandleGetChannelManagers returns channel managers from system
//
//	@Summary		Get channel managers
//	@Tags			channels
//	@Produce		json
//	@Security		BearerAuth
//	@Param			TgChannelID	path		int	true	"Telegram channel ID"
//	@Success		200			{object}	dto.ChannelManagersResponse
//	@Failure		400			{object}	dto.ErrorResponse
//	@Failure		401			{object}	dto.ErrorResponse
//	@Failure		404			{object}	dto.ErrorResponse
//	@Router			/channels/{TgChannelID}/managers [get]
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

// HandleAddManager adds a manager to the channel
//
//	@Summary		Add channel manager
//	@Tags			channels
//	@Accept			json
//	@Security		BearerAuth
//	@Param		TgChannelID	path	int					true	"Telegram channel ID"
//	@Param		request		body	dto.AddManagerRequest	true	"Manager telegram ID"
//	@Success		204
//	@Failure		400	{object}	dto.ErrorResponse
//	@Failure		401	{object}	dto.ErrorResponse
//	@Failure		403	{object}	dto.ErrorResponse
//	@Router			/channels/{TgChannelID}/managers [post]
func (a *App) HandleAddManager() http.HandlerFunc {
	log := a.log.With(logx.Handler("/api/v1/channels/{TgChannelID}/managers"))

	return func(w http.ResponseWriter, r *http.Request) {
		TgChannelID, err := strconv.ParseInt(chi.URLParam(r, "TgChannelID"), 10, 64)
		if err != nil {
			respond.Err(w, log, dto.ErrInvalidChannelID)
			return
		}

		var req dto.AddManagerRequest
		if err := bind.JSON(r, &req); err != nil {
			respond.Err(w, log, err)
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

// HandleRemoveManager removes a manager from the channel
//
//	@Summary		Remove channel manager
//	@Tags			channels
//	@Security		BearerAuth
//	@Param			TgChannelID	path	int	true	"Telegram channel ID"
//	@Param			tgID		path	int	true	"Manager telegram ID"
//	@Success		204
//	@Failure		400	{object}	dto.ErrorResponse
//	@Failure		401	{object}	dto.ErrorResponse
//	@Failure		403	{object}	dto.ErrorResponse
//	@Failure		404	{object}	dto.ErrorResponse
//	@Router			/channels/{TgChannelID}/managers/{tgID} [delete]
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

// HandleUpdateListing updates channel listing status
//
//	@Summary		Update channel listing
//	@Tags			channels
//	@Accept			json
//	@Security		BearerAuth
//	@Param		TgChannelID	path	int	true	"Telegram channel ID"
//	@Param		request	body	dto.UpdateListingRequest	true	"Listing status"
//	@Success		204
//	@Failure		400	{object}	dto.ErrorResponse
//	@Failure		401	{object}	dto.ErrorResponse
//	@Failure		403	{object}	dto.ErrorResponse
//	@Failure		404	{object}	dto.ErrorResponse
//	@Router			/channels/{TgChannelID}/listing [patch]
func (a *App) HandleUpdateListing() http.HandlerFunc {
	log := a.log.With(logx.Handler("/api/v1/channels/{TgChannelID}/listing"))

	return func(w http.ResponseWriter, r *http.Request) {
		TgChannelID, err := strconv.ParseInt(chi.URLParam(r, "TgChannelID"), 10, 64)
		if err != nil {
			respond.Err(w, log, dto.ErrInvalidChannelID)
			return
		}

		var req dto.UpdateListingRequest
		if err := bind.JSON(r, &req); err != nil {
			respond.Err(w, log, err)
			return
		}

		if err := a.channel.UpdateListing(r.Context(), TgChannelID, req.IsListed); err != nil {
			respond.Err(w, log, err)
			return
		}

		respond.NoContent(w)
	}
}

// HandleGetAdFormats returns channel ad formats
//
//	@Summary		Get channel ad formats
//	@Tags			channels
//	@Produce		json
//	@Security		BearerAuth
//	@Param			TgChannelID	path		int	true	"Telegram channel ID"
//	@Success		200			{object}	dto.AdFormatsResponse
//	@Failure		400			{object}	dto.ErrorResponse
//	@Failure		401			{object}	dto.ErrorResponse
//	@Failure		403			{object}	dto.ErrorResponse
//	@Failure		404			{object}	dto.ErrorResponse
//	@Router			/channels/{TgChannelID}/ad-formats [get]
func (a *App) HandleGetAdFormats() http.HandlerFunc {
	log := a.log.With(logx.Handler("/api/v1/channels/{TgChannelID}/ad-formats"))

	return func(w http.ResponseWriter, r *http.Request) {
		TgChannelID, err := strconv.ParseInt(chi.URLParam(r, "TgChannelID"), 10, 64)
		if err != nil {
			respond.Err(w, log, dto.ErrInvalidChannelID)
			return
		}

		formats, err := a.channel.GetAdFormats(r.Context(), TgChannelID)
		if err != nil {
			respond.Err(w, log, err)
			return
		}

		respond.OK(w, formats)
	}
}

// HandleAddAdFormat adds an ad format to the channel
//
//	@Summary		Add channel ad format
//	@Tags			channels
//	@Accept			json
//	@Security		BearerAuth
//	@Param		TgChannelID	path	int	true	"Telegram channel ID"
//	@Param		request	body	dto.AddAdFormatRequest	true	"Ad format details"
//	@Success		204
//	@Failure		400	{object}	dto.ErrorResponse
//	@Failure		401	{object}	dto.ErrorResponse
//	@Failure		403	{object}	dto.ErrorResponse
//	@Failure		404	{object}	dto.ErrorResponse
//	@Failure		409	{object}	dto.ErrorResponse
//	@Router			/channels/{TgChannelID}/ad-formats [post]
func (a *App) HandleAddAdFormat() http.HandlerFunc {
	log := a.log.With(logx.Handler("/api/v1/channels/{TgChannelID}/ad-formats"))

	return func(w http.ResponseWriter, r *http.Request) {
		TgChannelID, err := strconv.ParseInt(chi.URLParam(r, "TgChannelID"), 10, 64)
		if err != nil {
			respond.Err(w, log, dto.ErrInvalidChannelID)
			return
		}

		var req dto.AddAdFormatRequest
		if err := bind.JSON(r, &req); err != nil {
			respond.Err(w, log, err)
			return
		}

		if err := a.channel.AddAdFormat(r.Context(), TgChannelID, req); err != nil {
			respond.Err(w, log, err)
			return
		}

		respond.NoContent(w)
	}
}

// HandleRemoveAdFormat removes an ad format from the channel
//
//	@Summary		Remove channel ad format
//	@Tags			channels
//	@Security		BearerAuth
//	@Param			TgChannelID	path	int		true	"Telegram channel ID"
//	@Param			formatID	path	string	true	"Ad format UUID"
//	@Success		204
//	@Failure		400	{object}	dto.ErrorResponse
//	@Failure		401	{object}	dto.ErrorResponse
//	@Failure		403	{object}	dto.ErrorResponse
//	@Failure		404	{object}	dto.ErrorResponse
//	@Router			/channels/{TgChannelID}/ad-formats/{formatID} [delete]
func (a *App) HandleRemoveAdFormat() http.HandlerFunc {
	log := a.log.With(logx.Handler("/api/v1/channels/{TgChannelID}/ad-formats/{formatID}"))

	return func(w http.ResponseWriter, r *http.Request) {
		TgChannelID, err := strconv.ParseInt(chi.URLParam(r, "TgChannelID"), 10, 64)
		if err != nil {
			respond.Err(w, log, dto.ErrInvalidChannelID)
			return
		}

		formatID, err := uuid.Parse(chi.URLParam(r, "formatID"))
		if err != nil {
			respond.Err(w, log, dto.ErrInvalidFormatID)
			return
		}

		if err := a.channel.RemoveAdFormat(r.Context(), TgChannelID, formatID); err != nil {
			respond.Err(w, log, err)
			return
		}

		respond.NoContent(w)
	}
}
