package dto

import (
	"time"

	"github.com/bpva/ad-marketplace/internal/entity"
)

const (
	RoleCreator       = "creator"
	RoleAdministrator = "administrator"
)

type ChannelResponse struct {
	TgChannelID int64  `json:"id"`
	Title       string `json:"title"`
	Username    string `json:"username,omitempty"`
	IsListed    bool   `json:"is_listed"`
}

type ChannelWithRoleResponse struct {
	Channel ChannelResponse        `json:"channel"`
	Role    entity.ChannelRoleType `json:"role"`
}

type ChannelAdmin struct {
	TgID      int64  `json:"telegram_id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name,omitempty"`
	Username  string `json:"username,omitempty"`
	Role      string `json:"-"`
}

type ManagerResponse struct {
	TgID      int64                  `json:"telegram_id"`
	Name      string                 `json:"name"`
	Role      entity.ChannelRoleType `json:"role"`
	CreatedAt time.Time              `json:"created_at"`
}

type AddManagerRequest struct {
	TgID int64 `json:"telegram_id" validate:"required"`
}

type ChannelsResponse struct {
	Channels []ChannelWithRoleResponse `json:"channels"`
}

type ChannelAdminsResponse struct {
	Admins []ChannelAdmin `json:"admins"`
}

type ChannelManagersResponse struct {
	Managers []ManagerResponse `json:"managers"`
}

type ChannelFullInfo struct {
	ParticipantsCount int   `json:"participants_count"`
	LinkedChatID      int64 `json:"linked_chat_id,omitempty"`
}

type BroadcastStats struct {
	Period               StatsPeriod     `json:"period"`
	Followers            StatsValue      `json:"followers"`
	ViewsPerPost         StatsValue      `json:"views_per_post"`
	SharesPerPost        StatsValue      `json:"shares_per_post"`
	ReactionsPerPost     StatsValue      `json:"reactions_per_post"`
	EnabledNotifications StatsPercentage `json:"enabled_notifications"`
}

type StatsPeriod struct {
	MinDate int `json:"min_date"`
	MaxDate int `json:"max_date"`
}

type StatsValue struct {
	Current  float64 `json:"current"`
	Previous float64 `json:"previous"`
}

type StatsPercentage struct {
	Part  float64 `json:"part"`
	Total float64 `json:"total"`
}

type UpdateListingRequest struct {
	IsListed bool `json:"is_listed"`
}

type AddAdFormatRequest struct {
	FormatType   entity.AdFormatType `json:"format_type" validate:"required"`
	IsNative     bool                `json:"is_native"`
	FeedHours    int                 `json:"feed_hours" validate:"required,oneof=12 24"`
	TopHours     int                 `json:"top_hours" validate:"required,oneof=2 4"`
	PriceNanoTON int64               `json:"price_nano_ton" validate:"required,gt=0"`
}

type AdFormatResponse struct {
	ID           string              `json:"id"`
	FormatType   entity.AdFormatType `json:"format_type"`
	IsNative     bool                `json:"is_native"`
	FeedHours    int                 `json:"feed_hours"`
	TopHours     int                 `json:"top_hours"`
	PriceNanoTON int64               `json:"price_nano_ton"`
}

type AdFormatsResponse struct {
	AdFormats []AdFormatResponse `json:"ad_formats"`
}
