package dto

import "github.com/bpva/ad-marketplace/internal/entity"

type UpdateSettingsRequest struct {
	Language             *string `json:"language,omitempty"`
	ReceiveNotifications *bool   `json:"receive_notifications,omitempty"`
	PreferredMode        *string `json:"preferred_mode,omitempty"`
	OnboardingFinished   *bool   `json:"onboarding_finished,omitempty"`
	Theme                *string `json:"theme,omitempty"`
}

type UpdateNameRequest struct {
	Name string `json:"name"`
}

type ProfileResponse struct {
	TgID                 int64  `json:"telegram_id"`
	Name                 string `json:"name"`
	Language             string `json:"language"`
	ReceiveNotifications bool   `json:"receive_notifications"`
	PreferredMode        string `json:"preferred_mode"`
	OnboardingFinished   bool   `json:"onboarding_finished"`
	Theme                string `json:"theme"`
}

func ProfileResponseFrom(u *entity.User, s *entity.UserSettings) ProfileResponse {
	return ProfileResponse{
		TgID:                 u.TgID,
		Name:                 u.Name,
		Language:             string(s.Language),
		ReceiveNotifications: s.ReceiveNotifications,
		PreferredMode:        string(s.PreferredMode),
		OnboardingFinished:   s.OnboardingFinished,
		Theme:                string(s.Theme),
	}
}
