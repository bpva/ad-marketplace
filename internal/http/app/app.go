package app

import (
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
	httpCfg config.HTTP
	log     *slog.Logger
	bot     BotService
}

func New(httpCfg config.HTTP, log *slog.Logger, bot BotService) *App {
	return &App{
		httpCfg: httpCfg,
		log:     log,
		bot:     bot,
	}
}

func (a *App) Serve() error {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{a.httpCfg.FrontendURL},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	r.Route("/api/v1", func(r chi.Router) {
		r.Post("/bot/{token}/webhook", a.HandleBotWebhook())
	})

	a.log.Info("starting server", "port", a.httpCfg.Port)
	return http.ListenAndServe(":"+a.httpCfg.Port, r)
}
