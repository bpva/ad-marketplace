package app

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"

	"github.com/bpva/ad-marketplace/internal/config"
)

type BotService interface {
	ProcessUpdate(data []byte) error
	Token() string
}

type App struct {
	log *slog.Logger
	bot BotService
	srv *http.Server
}

func New(httpCfg config.HTTP, log *slog.Logger, bot BotService) *App {
	a := &App{log: log, bot: bot}

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{httpCfg.FrontendURL},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	r.Route("/api/v1", func(r chi.Router) {
		r.Post("/bot/{token}/webhook", a.HandleBotWebhook())
	})

	a.srv = &http.Server{
		Addr:    ":" + httpCfg.Port,
		Handler: r,
	}

	return a
}

func (a *App) Serve() error {
	a.log.Info("starting server", "addr", a.srv.Addr)
	if err := a.srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return err
	}
	return nil
}

func (a *App) Shutdown(ctx context.Context) error {
	a.log.Info("shutting down server")
	return a.srv.Shutdown(ctx)
}
