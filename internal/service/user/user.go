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
	UpdateWalletAddress(ctx context.Context, id uuid.UUID, address *string) error
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

	if err := s.userRepo.UpdateName(ctx, user.ID, name); err != nil {
		return fmt.Errorf("update name: %w", err)
	}

	s.log.Info("user name updated", "user_id", user.ID)
	return nil
}

func (s *svc) LinkWallet(ctx context.Context, address string) error {
	user, ok := dto.UserFromContext(ctx)
	if !ok {
		return fmt.Errorf("link wallet: %w", dto.ErrForbidden)
	}

	if err := s.userRepo.UpdateWalletAddress(ctx, user.ID, &address); err != nil {
		return fmt.Errorf("link wallet: %w", err)
	}

	s.log.Info("wallet linked", "user_id", user.ID)
	return nil
}

func (s *svc) UnlinkWallet(ctx context.Context) error {
	user, ok := dto.UserFromContext(ctx)
	if !ok {
		return fmt.Errorf("unlink wallet: %w", dto.ErrForbidden)
	}

	if err := s.userRepo.UpdateWalletAddress(ctx, user.ID, nil); err != nil {
		return fmt.Errorf("unlink wallet: %w", err)
	}

	s.log.Info("wallet unlinked", "user_id", user.ID)
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
		settings.Language = *req.Language
	}
	if req.ReceiveNotifications != nil {
		settings.ReceiveNotifications = *req.ReceiveNotifications
	}
	if req.PreferredMode != nil {
		settings.PreferredMode = *req.PreferredMode
	}
	if req.OnboardingFinished != nil {
		settings.OnboardingFinished = *req.OnboardingFinished
	}
	if req.Theme != nil {
		settings.Theme = *req.Theme
	}

	if err := s.settingsRepo.Update(ctx, settings); err != nil {
		return fmt.Errorf("update settings: %w", err)
	}

	s.log.Info("user settings updated", "user_id", user.ID)
	return nil
}

func (s *svc) getOrCreateSettings(
	ctx context.Context,
	userID uuid.UUID,
) (*entity.UserSettings, error) {
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
