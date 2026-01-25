package dto

import "time"

const (
	RoleCreator       = "creator"
	RoleAdministrator = "administrator"
)

type ChannelResponse struct {
	TelegramChannelID int64  `json:"id"`
	Title             string `json:"title"`
	Username          string `json:"username,omitempty"`
}

type ChannelWithRoleResponse struct {
	Channel ChannelResponse `json:"channel"`
	Role    string          `json:"role"`
}

type ChannelAdmin struct {
	TelegramID int64  `json:"telegram_id"`
	FirstName  string `json:"first_name"`
	LastName   string `json:"last_name,omitempty"`
	Username   string `json:"username,omitempty"`
	Role       string `json:"role"`
}

type ManagerResponse struct {
	TelegramID int64     `json:"telegram_id"`
	Name       string    `json:"name"`
	Role       string    `json:"role"`
	CreatedAt  time.Time `json:"created_at"`
}

type AddManagerRequest struct {
	TelegramID int64 `json:"telegram_id"`
}
