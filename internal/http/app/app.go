package app

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/google/uuid"

	"github.com/bpva/ad-marketplace/internal/config"
	"github.com/bpva/ad-marketplace/internal/dto"
	"github.com/bpva/ad-marketplace/internal/entity"
	"github.com/bpva/ad-marketplace/internal/http/middleware"
)

//go:generate mockgen -destination=mocks.go -package=app . BotService

type BotService interface {
	ProcessUpdate(data []byte) error
	Token() string
}

type AuthService interface {
	Authenticate(ctx context.Context, initData string) (string, *entity.User, error)
	ValidateToken(tokenString string) (*dto.Claims, error)
	GetUserByID(ctx context.Context, userID uuid.UUID) (*entity.User, error)
}

type App struct {
	log  *slog.Logger
	bot  BotService
	auth AuthService
	srv  *http.Server
}

func New(httpCfg config.HTTP, log *slog.Logger, bot BotService, authSvc AuthService) *App {
	a := &App{log: log, bot: bot, auth: authSvc}

	r := chi.NewRouter()
	r.Use(chimw.Logger)
	r.Use(chimw.Recoverer)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{httpCfg.FrontendURL},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	r.Route("/api/v1", func(r chi.Router) {
		r.Post("/bot/{token}/webhook", a.HandleBotWebhook())
		r.Post("/auth", a.HandleAuth())

		r.Group(func(r chi.Router) {
			r.Use(middleware.Auth(authSvc))
			r.Get("/me", a.HandleMe())
		})
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

func (a *App) Handler() http.Handler {
	return a.srv.Handler
}
