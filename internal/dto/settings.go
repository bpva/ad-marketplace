package dto

import "github.com/bpva/ad-marketplace/internal/entity"

type SettingsResponse struct {
	Language             string `json:"language"`
	ReceiveNotifications bool   `json:"receive_notifications"`
	PreferredMode        string `json:"preferred_mode"`
	OnboardingFinished   bool   `json:"onboarding_finished"`
}

func SettingsResponseFrom(s *entity.UserSettings) SettingsResponse {
	return SettingsResponse{
		Language:             string(s.Language),
		ReceiveNotifications: s.ReceiveNotifications,
		PreferredMode:        string(s.PreferredMode),
		OnboardingFinished:   s.OnboardingFinished,
	}
}

type UpdateSettingsRequest struct {
	Language             *string `json:"language,omitempty"`
	ReceiveNotifications *bool   `json:"receive_notifications,omitempty"`
	PreferredMode        *string `json:"preferred_mode,omitempty"`
	OnboardingFinished   *bool   `json:"onboarding_finished,omitempty"`
}

type UpdateNameRequest struct {
	Name string `json:"name"`
}

type ProfileResponse struct {
	ID                   string `json:"id"`
	TgID                 int64  `json:"telegram_id"`
	Name                 string `json:"name"`
	Language             string `json:"language"`
	ReceiveNotifications bool   `json:"receive_notifications"`
	PreferredMode        string `json:"preferred_mode"`
	OnboardingFinished   bool   `json:"onboarding_finished"`
}

func ProfileResponseFrom(u *entity.User, s *entity.UserSettings) ProfileResponse {
	return ProfileResponse{
		ID:                   u.ID.String(),
		TgID:                 u.TgID,
		Name:                 u.Name,
		Language:             string(s.Language),
		ReceiveNotifications: s.ReceiveNotifications,
		PreferredMode:        string(s.PreferredMode),
		OnboardingFinished:   s.OnboardingFinished,
	}
}
