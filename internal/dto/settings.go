package dto

import "github.com/bpva/ad-marketplace/internal/entity"

type UpdateSettingsRequest struct {
	Language             *entity.Language      `json:"language,omitempty"`
	ReceiveNotifications *bool                 `json:"receive_notifications,omitempty"`
	PreferredMode        *entity.PreferredMode `json:"preferred_mode,omitempty"`
	OnboardingFinished   *bool                 `json:"onboarding_finished,omitempty"`
	Theme                *entity.Theme         `json:"theme,omitempty"`
}

type UpdateNameRequest struct {
	Name string `json:"name" validate:"required"`
}

type ProfileResponse struct {
	TgID                 int64                `json:"telegram_id"`
	Name                 string               `json:"name"`
	WalletAddress        *string              `json:"wallet_address"`
	Language             entity.Language      `json:"language"`
	ReceiveNotifications bool                 `json:"receive_notifications"`
	PreferredMode        entity.PreferredMode `json:"preferred_mode"`
	OnboardingFinished   bool                 `json:"onboarding_finished"`
	Theme                entity.Theme         `json:"theme"`
}

func ProfileResponseFrom(u *entity.User, s *entity.UserSettings) ProfileResponse {
	return ProfileResponse{
		TgID:                 u.TgID,
		Name:                 u.Name,
		WalletAddress:        u.WalletAddress,
		Language:             s.Language,
		ReceiveNotifications: s.ReceiveNotifications,
		PreferredMode:        s.PreferredMode,
		OnboardingFinished:   s.OnboardingFinished,
		Theme:                s.Theme,
	}
}

type LinkWalletRequest struct {
	Address string `json:"address" validate:"required"`
}
