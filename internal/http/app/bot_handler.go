package app

import (
	"io"
	"net/http"

	"github.com/bpva/ad-marketplace/internal/logx"
	"github.com/go-chi/chi/v5"
)

func (a *App) HandleBotWebhook() http.HandlerFunc {
	log := a.log.With(logx.Handler("/api/v1/bot/{token}/webhook"))

	return func(w http.ResponseWriter, r *http.Request) {
		token := chi.URLParam(r, "token")
		if token != a.bot.Token() {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		body, err := io.ReadAll(r.Body)
		if err != nil {
			log.Error("failed to read body", "error", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		if err := a.bot.ProcessUpdate(body); err != nil {
			log.Error("failed to process update", "error", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}
