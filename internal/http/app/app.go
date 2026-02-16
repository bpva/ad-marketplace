package app

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/google/uuid"

	"github.com/bpva/ad-marketplace/internal/config"
	"github.com/bpva/ad-marketplace/internal/dto"
	"github.com/bpva/ad-marketplace/internal/entity"
	"github.com/bpva/ad-marketplace/internal/http/middleware"
)

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
	UpdateCategories(ctx context.Context, TgChannelID int64, categories []string) error
	GetChannelPhoto(ctx context.Context, tgChannelID int64, size string) ([]byte, error)
	GetMarketplaceChannels(
		ctx context.Context,
		req dto.MarketplaceChannelsRequest,
	) (*dto.MarketplaceChannelsResponse, error)
}

type UserService interface {
	GetProfile(ctx context.Context) (*dto.ProfileResponse, error)
	UpdateName(ctx context.Context, name string) error
	UpdateSettings(ctx context.Context, req dto.UpdateSettingsRequest) error
	LinkWallet(ctx context.Context, address string) error
	UnlinkWallet(ctx context.Context) error
}

type PostService interface {
	GetUserTemplates(ctx context.Context) (*dto.TemplatesResponse, error)
	GetPostMedia(ctx context.Context, postID uuid.UUID) ([]byte, error)
	SendPreview(ctx context.Context, postID uuid.UUID) error
}

type TonRatesService interface {
	GetRates(ctx context.Context) (*dto.TonRatesResponse, error)
}

type App struct {
	log      *slog.Logger
	bot      BotService
	auth     AuthService
	channel  ChannelService
	user     UserService
	post     PostService
	tonRates TonRatesService
	srv      *http.Server
}

func New(
	cfg config.HTTP,
	log *slog.Logger,
	bot BotService,
	authSvc AuthService,
	channelSvc ChannelService,
	userSvc UserService,
	postSvc PostService,
	tonRatesSvc TonRatesService,
) *App {
	a := &App{
		log:      log,
		bot:      bot,
		auth:     authSvc,
		channel:  channelSvc,
		user:     userSvc,
		post:     postSvc,
		tonRates: tonRatesSvc,
	}

	r := chi.NewRouter()
	r.Use(chimw.Logger)
	r.Use(chimw.Recoverer)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{cfg.FrontendURL},
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
		r.Get("/ton-rates", a.HandleGetTonRates())

		r.Group(func(r chi.Router) {
			r.Use(middleware.Auth(authSvc, log))
			r.Get("/channels/{TgChannelID}/photo", a.HandleGetChannelPhoto())
			r.Get("/posts/{postID}/media", a.HandleGetPostMedia())
		})

		r.Group(func(r chi.Router) {
			r.Use(middleware.Auth(authSvc, log))
			r.Use(openAPIMw)
			r.Get("/me", a.HandleMe())

			r.Route("/user", func(r chi.Router) {
				r.Get("/profile", a.HandleGetProfile())
				r.Patch("/name", a.HandleUpdateName())
				r.Patch("/settings", a.HandleUpdateSettings())
				r.Put("/wallet", a.HandleLinkWallet())
				r.Delete("/wallet", a.HandleUnlinkWallet())
			})

			r.Route("/mp", func(r chi.Router) {
				r.Post("/channels", a.HandleGetMarketplaceChannels())
			})

			r.Get("/posts", a.HandleListTemplates())
			r.Post("/posts/{postID}/preview", a.HandleSendPreview())

			r.Route("/channels", func(r chi.Router) {
				r.Get("/", a.HandleListChannels())
				r.Get("/{TgChannelID}", a.HandleGetChannel())
				r.Get("/{TgChannelID}/admins", a.HandleGetChannelAdmins())
				r.Get("/{TgChannelID}/managers", a.HandleGetChannelManagers())
				r.Post("/{TgChannelID}/managers", a.HandleAddManager())
				r.Delete("/{TgChannelID}/managers/{tgID}", a.HandleRemoveManager())
				r.Patch("/{TgChannelID}/listing", a.HandleUpdateListing())
				r.Patch("/{TgChannelID}/categories", a.HandleUpdateCategories())
				r.Get("/{TgChannelID}/ad-formats", a.HandleGetAdFormats())
				r.Post("/{TgChannelID}/ad-formats", a.HandleAddAdFormat())
				r.Delete("/{TgChannelID}/ad-formats/{formatID}", a.HandleRemoveAdFormat())
			})
		})
	})

	a.srv = &http.Server{
		Addr:    fmt.Sprintf(":%s", cfg.Port),
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
