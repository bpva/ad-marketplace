package app

import (
	"encoding/json"
	"io"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/bpva/ad-marketplace/internal/dto"
	"github.com/bpva/ad-marketplace/internal/http/bind"
	"github.com/bpva/ad-marketplace/internal/http/respond"
	"github.com/bpva/ad-marketplace/internal/logx"
	"github.com/bpva/ad-marketplace/internal/service/deal"
)

// HandleCreateDeal creates a new advertising deal
//
//	@Summary		Create deal
//	@Tags			deals
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			request	body		dto.CreateDealRequest	true	"Deal details"
//	@Success		201		{object}	dto.DealResponse
//	@Failure		400		{object}	dto.ErrorResponse
//	@Failure		401		{object}	dto.ErrorResponse
//	@Failure		403		{object}	dto.ErrorResponse
//	@Failure		404		{object}	dto.ErrorResponse
//	@Failure		409		{object}	dto.ErrorResponse
//	@Failure		422		{object}	dto.ErrorResponse
//	@Router			/deals [post]
func (a *App) HandleCreateDeal() http.HandlerFunc {
	log := a.log.With(logx.Handler("/api/v1/deals"))

	return func(w http.ResponseWriter, r *http.Request) {
		var req dto.CreateDealRequest
		if err := bind.JSON(r, &req); err != nil {
			respond.Err(w, log, err)
			return
		}

		templatePostID, err := uuid.Parse(req.TemplatePostID)
		if err != nil {
			respond.Err(w, log, dto.ErrBadRequest)
			return
		}

		d, posts, err := a.deal.CreateDeal(r.Context(), deal.CreateDealParams{
			TgChannelID:    req.TgChannelID,
			FormatType:     req.FormatType,
			IsNative:       req.IsNative,
			FeedHours:      req.FeedHours,
			TopHours:       req.TopHours,
			PriceNanoTON:   req.PriceNanoTON,
			TemplatePostID: templatePostID,
			ScheduledAt:    req.ScheduledAt,
		})
		if err != nil {
			respond.Err(w, log, err)
			return
		}

		respond.Created(w, dto.DealResponseFrom(d, posts, req.TgChannelID))
	}
}

// HandleListDeals lists deals for advertiser or publisher
//
//	@Summary		List deals
//	@Tags			deals
//	@Produce		json
//	@Security		BearerAuth
//	@Param			role		query	string	false	"Role"		default(advertiser)
//	@Param			channel_id	query	int		false	"Channel ID"
//	@Param			page		query	int		false	"Page number"	default(1)
//	@Success		200			{object}	dto.DealsResponse
//	@Failure		400			{object}	dto.ErrorResponse
//	@Failure		401			{object}	dto.ErrorResponse
//	@Failure		403			{object}	dto.ErrorResponse
//	@Router			/deals [get]
func (a *App) HandleListDeals() http.HandlerFunc {
	log := a.log.With(logx.Handler("/api/v1/deals"))

	return func(w http.ResponseWriter, r *http.Request) {
		role := r.URL.Query().Get("role")
		if role == "" {
			role = "advertiser"
		}

		page := 1
		if p := r.URL.Query().Get("page"); p != "" {
			parsed, err := strconv.Atoi(p)
			if err == nil && parsed > 0 {
				page = parsed
			}
		}

		const pageSize = 20
		offset := (page - 1) * pageSize

		switch role {
		case "advertiser":
			items, total, err := a.deal.ListAdvertiserDeals(r.Context(), pageSize, offset)
			if err != nil {
				respond.Err(w, log, err)
				return
			}
			respond.OK(w, buildDealsResponse(items, total))

		case "publisher":
			tgChannelID, err := strconv.ParseInt(r.URL.Query().Get("channel_id"), 10, 64)
			if err != nil {
				respond.Err(w, log, dto.ErrInvalidChannelID)
				return
			}
			items, total, err := a.deal.ListPublisherDeals(
				r.Context(),
				tgChannelID,
				pageSize,
				offset,
			)
			if err != nil {
				respond.Err(w, log, err)
				return
			}
			respond.OK(w, buildDealsResponse(items, total))

		default:
			respond.Err(w, log, dto.ErrInvalidRole)
		}
	}
}

func buildDealsResponse(items []dto.DealListItem, total int) dto.DealsResponse {
	deals := make([]dto.DealResponse, len(items))
	for i := range items {
		deals[i] = dto.DealListResponseFrom(items[i])
	}
	return dto.DealsResponse{Deals: deals, Total: total}
}

// HandleGetDeal returns deal details
//
//	@Summary		Get deal
//	@Tags			deals
//	@Produce		json
//	@Security		BearerAuth
//	@Param			dealID	path		string	true	"Deal ID"
//	@Success		200		{object}	dto.DealResponse
//	@Failure		400		{object}	dto.ErrorResponse
//	@Failure		401		{object}	dto.ErrorResponse
//	@Failure		403		{object}	dto.ErrorResponse
//	@Failure		404		{object}	dto.ErrorResponse
//	@Router			/deals/{dealID} [get]
func (a *App) HandleGetDeal() http.HandlerFunc {
	log := a.log.With(logx.Handler("/api/v1/deals/{dealID}"))

	return func(w http.ResponseWriter, r *http.Request) {
		dealID, err := uuid.Parse(chi.URLParam(r, "dealID"))
		if err != nil {
			respond.Err(w, log, dto.ErrInvalidDealID)
			return
		}

		d, posts, tgChannelID, err := a.deal.GetDeal(r.Context(), dealID)
		if err != nil {
			respond.Err(w, log, err)
			return
		}

		respond.OK(w, dto.DealResponseFrom(d, posts, tgChannelID))
	}
}

// HandleApproveDeal approves a deal
//
//	@Summary		Approve deal
//	@Tags			deals
//	@Security		BearerAuth
//	@Param			dealID	path	string	true	"Deal ID"
//	@Success		204
//	@Failure		400	{object}	dto.ErrorResponse
//	@Failure		401	{object}	dto.ErrorResponse
//	@Failure		403	{object}	dto.ErrorResponse
//	@Failure		404	{object}	dto.ErrorResponse
//	@Router			/deals/{dealID}/approve [post]
func (a *App) HandleApproveDeal() http.HandlerFunc {
	log := a.log.With(logx.Handler("/api/v1/deals/{dealID}/approve"))

	return func(w http.ResponseWriter, r *http.Request) {
		dealID, err := uuid.Parse(chi.URLParam(r, "dealID"))
		if err != nil {
			respond.Err(w, log, dto.ErrInvalidDealID)
			return
		}

		if err := a.deal.Approve(r.Context(), dealID); err != nil {
			respond.Err(w, log, err)
			return
		}

		respond.NoContent(w)
	}
}

// HandleRejectDeal rejects a deal
//
//	@Summary		Reject deal
//	@Tags			deals
//	@Accept			json
//	@Security		BearerAuth
//	@Param			dealID	path	string				true	"Deal ID"
//	@Param			request	body	dto.RejectRequest	false	"Rejection reason"
//	@Success		204
//	@Failure		400	{object}	dto.ErrorResponse
//	@Failure		401	{object}	dto.ErrorResponse
//	@Failure		403	{object}	dto.ErrorResponse
//	@Failure		404	{object}	dto.ErrorResponse
//	@Router			/deals/{dealID}/reject [post]
func (a *App) HandleRejectDeal() http.HandlerFunc {
	log := a.log.With(logx.Handler("/api/v1/deals/{dealID}/reject"))

	return func(w http.ResponseWriter, r *http.Request) {
		dealID, err := uuid.Parse(chi.URLParam(r, "dealID"))
		if err != nil {
			respond.Err(w, log, dto.ErrInvalidDealID)
			return
		}

		var req dto.RejectRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil && err != io.EOF {
			respond.Err(w, log, dto.ErrBadRequest)
			return
		}

		if err := a.deal.Reject(r.Context(), dealID, req.Reason); err != nil {
			respond.Err(w, log, err)
			return
		}

		respond.NoContent(w)
	}
}

// HandleRequestChanges requests changes on a deal
//
//	@Summary		Request changes on deal
//	@Tags			deals
//	@Accept			json
//	@Security		BearerAuth
//	@Param			dealID	path	string						true	"Deal ID"
//	@Param			request	body	dto.RequestChangesRequest	true	"Details"
//	@Success		204
//	@Failure		400	{object}	dto.ErrorResponse
//	@Failure		401	{object}	dto.ErrorResponse
//	@Failure		403	{object}	dto.ErrorResponse
//	@Failure		404	{object}	dto.ErrorResponse
//	@Router			/deals/{dealID}/request-changes [post]
func (a *App) HandleRequestChanges() http.HandlerFunc {
	log := a.log.With(logx.Handler("/api/v1/deals/{dealID}/request-changes"))

	return func(w http.ResponseWriter, r *http.Request) {
		dealID, err := uuid.Parse(chi.URLParam(r, "dealID"))
		if err != nil {
			respond.Err(w, log, dto.ErrInvalidDealID)
			return
		}

		var req dto.RequestChangesRequest
		if err := bind.JSON(r, &req); err != nil {
			respond.Err(w, log, err)
			return
		}

		if err := a.deal.RequestChanges(r.Context(), dealID, req.Note); err != nil {
			respond.Err(w, log, err)
			return
		}

		respond.NoContent(w)
	}
}

// HandleCancelDeal cancels a deal
//
//	@Summary		Cancel deal
//	@Tags			deals
//	@Security		BearerAuth
//	@Param			dealID	path	string	true	"Deal ID"
//	@Success		204
//	@Failure		400	{object}	dto.ErrorResponse
//	@Failure		401	{object}	dto.ErrorResponse
//	@Failure		403	{object}	dto.ErrorResponse
//	@Failure		404	{object}	dto.ErrorResponse
//	@Router			/deals/{dealID}/cancel [post]
func (a *App) HandleCancelDeal() http.HandlerFunc {
	log := a.log.With(logx.Handler("/api/v1/deals/{dealID}/cancel"))

	return func(w http.ResponseWriter, r *http.Request) {
		dealID, err := uuid.Parse(chi.URLParam(r, "dealID"))
		if err != nil {
			respond.Err(w, log, dto.ErrInvalidDealID)
			return
		}

		if err := a.deal.Cancel(r.Context(), dealID); err != nil {
			respond.Err(w, log, err)
			return
		}

		respond.NoContent(w)
	}
}
