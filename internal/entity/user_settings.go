package entity

import "github.com/google/uuid"

type Language string

const (
	LanguageEN Language = "en"
	LanguageRU Language = "ru"
)

type PreferredMode string

const (
	PreferredModePublisher  PreferredMode = "publisher"
	PreferredModeAdvertiser PreferredMode = "advertiser"
)

type Theme string

const (
	ThemeLight Theme = "light"
	ThemeDark  Theme = "dark"
	ThemeAuto  Theme = "auto"
)

type UserSettings struct {
	UserID               uuid.UUID     `db:"user_id"`
	Language             Language      `db:"language"`
	ReceiveNotifications bool          `db:"receive_notifications"`
	PreferredMode        PreferredMode `db:"preferred_mode"`
	OnboardingFinished   bool          `db:"onboarding_finished"`
	Theme                Theme         `db:"theme"`
}
