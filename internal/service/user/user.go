package user

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/google/uuid"

	"github.com/bpva/ad-marketplace/internal/dto"
	"github.com/bpva/ad-marketplace/internal/entity"
	"github.com/bpva/ad-marketplace/internal/logx"
)

type UserRepository interface {
	GetByID(ctx context.Context, id uuid.UUID) (*entity.User, error)
	UpdateName(ctx context.Context, id uuid.UUID, name string) error
}

type SettingsRepository interface {
	GetByUserID(ctx context.Context, userID uuid.UUID) (*entity.UserSettings, error)
	Create(ctx context.Context, userID uuid.UUID) (*entity.UserSettings, error)
	Update(ctx context.Context, s *entity.UserSettings) error
}

type svc struct {
	userRepo     UserRepository
	settingsRepo SettingsRepository
	log          *slog.Logger
}

func New(
	userRepo UserRepository,
	settingsRepo SettingsRepository,
	log *slog.Logger,
) *svc {
	log = log.With(logx.Service("UserService"))
	return &svc{
		userRepo:     userRepo,
		settingsRepo: settingsRepo,
		log:          log,
	}
}

func (s *svc) GetProfile(ctx context.Context) (*dto.ProfileResponse, error) {
	user, ok := dto.UserFromContext(ctx)
	if !ok {
		return nil, fmt.Errorf("get profile: %w", dto.ErrForbidden)
	}

	u, err := s.userRepo.GetByID(ctx, user.ID)
	if err != nil {
		return nil, fmt.Errorf("get user: %w", err)
	}

	settings, err := s.getOrCreateSettings(ctx, user.ID)
	if err != nil {
		return nil, err
	}

	resp := dto.ProfileResponseFrom(u, settings)
	return &resp, nil
}

func (s *svc) UpdateName(ctx context.Context, name string) error {
	user, ok := dto.UserFromContext(ctx)
	if !ok {
		return fmt.Errorf("update name: %w", dto.ErrForbidden)
	}

	if name == "" {
		return fmt.Errorf("update name: %w", dto.ErrBadRequest)
	}

	if err := s.userRepo.UpdateName(ctx, user.ID, name); err != nil {
		return fmt.Errorf("update name: %w", err)
	}

	s.log.Info("user name updated", "user_id", user.ID)
	return nil
}

func (s *svc) UpdateSettings(ctx context.Context, req dto.UpdateSettingsRequest) error {
	user, ok := dto.UserFromContext(ctx)
	if !ok {
		return fmt.Errorf("update settings: %w", dto.ErrForbidden)
	}

	settings, err := s.getOrCreateSettings(ctx, user.ID)
	if err != nil {
		return err
	}

	if req.Language != nil {
		lang := entity.Language(*req.Language)
		if lang != entity.LanguageEN && lang != entity.LanguageRU {
			return fmt.Errorf("update settings: invalid language: %w", dto.ErrBadRequest)
		}
		settings.Language = lang
	}

	if req.ReceiveNotifications != nil {
		settings.ReceiveNotifications = *req.ReceiveNotifications
	}

	if req.PreferredMode != nil {
		mode := entity.PreferredMode(*req.PreferredMode)
		if mode != entity.PreferredModePublisher && mode != entity.PreferredModeAdvertiser {
			return fmt.Errorf("update settings: invalid preferred mode: %w", dto.ErrBadRequest)
		}
		settings.PreferredMode = mode
	}

	if req.OnboardingFinished != nil {
		settings.OnboardingFinished = *req.OnboardingFinished
	}

	if err := s.settingsRepo.Update(ctx, settings); err != nil {
		return fmt.Errorf("update settings: %w", err)
	}

	s.log.Info("user settings updated", "user_id", user.ID)
	return nil
}

func (s *svc) getOrCreateSettings(ctx context.Context, userID uuid.UUID) (*entity.UserSettings, error) {
	settings, err := s.settingsRepo.GetByUserID(ctx, userID)
	if errors.Is(err, dto.ErrNotFound) {
		settings, err = s.settingsRepo.Create(ctx, userID)
		if err != nil {
			return nil, fmt.Errorf("create settings: %w", err)
		}
		s.log.Info("user settings created", "user_id", userID)
	} else if err != nil {
		return nil, fmt.Errorf("get settings: %w", err)
	}
	return settings, nil
}
