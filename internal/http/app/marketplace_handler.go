package app

import (
	"net/http"

	"github.com/bpva/ad-marketplace/internal/dto"
	"github.com/bpva/ad-marketplace/internal/http/bind"
	"github.com/bpva/ad-marketplace/internal/http/respond"
	"github.com/bpva/ad-marketplace/internal/logx"
)

// HandleGetMarketplaceChannels returns marketplace channel listing
//
//	@Summary		List marketplace channels
//	@Tags			marketplace
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			request	body		dto.MarketplaceChannelsRequest	true	"Search/sort/filter params"
//	@Success		200		{object}	dto.MarketplaceChannelsResponse
//	@Failure		400		{object}	dto.ErrorResponse
//	@Failure		401		{object}	dto.ErrorResponse
//	@Router			/mp/channels [post]
func (a *App) HandleGetMarketplaceChannels() http.HandlerFunc {
	log := a.log.With(logx.Handler("/api/v1/mp/channels"))

	return func(w http.ResponseWriter, r *http.Request) {
		var req dto.MarketplaceChannelsRequest
		if err := bind.JSON(r, &req); err != nil {
			respond.Err(w, log, err)
			return
		}

		channels, err := a.channel.GetMarketplaceChannels(r.Context(), req)
		if err != nil {
			respond.Err(w, log, err)
			return
		}

		respond.OK(w, channels)
	}
}
