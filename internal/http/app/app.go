package app

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/google/uuid"

	"github.com/bpva/ad-marketplace/internal/config"
	"github.com/bpva/ad-marketplace/internal/dto"
	"github.com/bpva/ad-marketplace/internal/entity"
	"github.com/bpva/ad-marketplace/internal/http/middleware"
)

type StatsService interface {
	GetInfo(ctx context.Context, tgChannelID int64) (*dto.ChannelInfoResponse, error)
	GetHistory(
		ctx context.Context,
		tgChannelID int64,
		from, to time.Time,
	) (*dto.ChannelHistoricalStatsResponse, error)
}

type BotService interface {
	ProcessUpdate(data []byte) error
	Token() string
}

type AuthService interface {
	Authenticate(ctx context.Context, initData string) (string, *entity.User, error)
	ValidateToken(tokenString string) (*dto.Claims, error)
	GetUserByID(ctx context.Context, userID uuid.UUID) (*entity.User, error)
}

type ChannelService interface {
	GetUserChannels(ctx context.Context) (*dto.ChannelsResponse, error)
	GetChannel(ctx context.Context, TgChannelID int64) (*dto.ChannelResponse, error)
	GetChannelAdmins(ctx context.Context, TgChannelID int64) (*dto.ChannelAdminsResponse, error)
	GetChannelManagers(ctx context.Context, TgChannelID int64) (*dto.ChannelManagersResponse, error)
	AddManager(ctx context.Context, TgChannelID int64, tgID int64) error
	RemoveManager(ctx context.Context, TgChannelID int64, tgID int64) error
	UpdateListing(ctx context.Context, TgChannelID int64, isListed bool) error
	GetAdFormats(ctx context.Context, TgChannelID int64) (*dto.AdFormatsResponse, error)
	AddAdFormat(ctx context.Context, TgChannelID int64, req dto.AddAdFormatRequest) error
	RemoveAdFormat(ctx context.Context, TgChannelID int64, formatID uuid.UUID) error
}

type UserService interface {
	GetProfile(ctx context.Context) (*dto.ProfileResponse, error)
	UpdateName(ctx context.Context, name string) error
	UpdateSettings(ctx context.Context, req dto.UpdateSettingsRequest) error
}

type App struct {
	log     *slog.Logger
	bot     BotService
	auth    AuthService
	channel ChannelService
	user    UserService
	stats   StatsService
	srv     *http.Server
}

func New(
	httpCfg config.HTTP,
	log *slog.Logger,
	bot BotService,
	authSvc AuthService,
	channelSvc ChannelService,
	userSvc UserService,
	statsSvc StatsService,
) *App {
	a := &App{
		log:     log,
		bot:     bot,
		auth:    authSvc,
		channel: channelSvc,
		user:    userSvc,
		stats:   statsSvc,
	}

	r := chi.NewRouter()
	r.Use(chimw.Logger)
	r.Use(chimw.Recoverer)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{httpCfg.FrontendURL},
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	openAPIMw, err := middleware.NewOpenAPIValidator(log)
	if err != nil {
		log.Error("initialize OpenAPI middleware", "error", err)
		os.Exit(1)
	}

	r.Route("/api/v1", func(r chi.Router) {
		r.Post("/bot/{token}/webhook", a.HandleBotWebhook())
		r.Post("/auth", a.HandleAuth())

		r.Group(func(r chi.Router) {
			r.Use(middleware.Auth(authSvc, log))
			r.Use(openAPIMw)
			r.Get("/me", a.HandleMe())

			r.Route("/user", func(r chi.Router) {
				r.Get("/profile", a.HandleGetProfile())
				r.Patch("/name", a.HandleUpdateName())
				r.Patch("/settings", a.HandleUpdateSettings())
			})

			r.Route("/channels", func(r chi.Router) {
				r.Get("/", a.HandleListChannels())
				r.Get("/{TgChannelID}", a.HandleGetChannel())
				r.Get("/{TgChannelID}/admins", a.HandleGetChannelAdmins())
				r.Get("/{TgChannelID}/managers", a.HandleGetChannelManagers())
				r.Post("/{TgChannelID}/managers", a.HandleAddManager())
				r.Delete("/{TgChannelID}/managers/{tgID}", a.HandleRemoveManager())
				r.Patch("/{TgChannelID}/listing", a.HandleUpdateListing())
				r.Get("/{TgChannelID}/ad-formats", a.HandleGetAdFormats())
				r.Post("/{TgChannelID}/ad-formats", a.HandleAddAdFormat())
				r.Delete("/{TgChannelID}/ad-formats/{formatID}", a.HandleRemoveAdFormat())
				r.Get("/{TgChannelID}/info", a.HandleGetChannelInfo())
				r.Get("/{TgChannelID}/stats", a.HandleGetChannelStats())
			})
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
