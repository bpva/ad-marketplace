package app

import (
	"net/http"

	_ "github.com/bpva/ad-marketplace/internal/dto"
	"github.com/bpva/ad-marketplace/internal/http/respond"
	"github.com/bpva/ad-marketplace/internal/logx"
)

// HandleGetTonRates returns current TON exchange rates
//
//	@Summary		Get TON exchange rates
//	@Tags			rates
//	@Produce		json
//	@Success		200	{object}	dto.TonRatesResponse
//	@Failure		500	{object}	dto.ErrorResponse
//	@Router			/ton-rates [get]
func (a *App) HandleGetTonRates() http.HandlerFunc {
	log := a.log.With(logx.Handler("/api/v1/ton-rates"))

	return func(w http.ResponseWriter, r *http.Request) {
		rates, err := a.tonRates.GetRates(r.Context())
		if err != nil {
			respond.Err(w, log, err)
			return
		}

		respond.OK(w, rates)
	}
}
