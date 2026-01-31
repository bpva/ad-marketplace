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
	TgID int64 `json:"telegram_id"`
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
