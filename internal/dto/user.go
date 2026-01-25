package dto

import "github.com/bpva/ad-marketplace/internal/entity"

type UserResponse struct {
	ID   string `json:"id"`
	TgID int64  `json:"telegram_id"`
	Name string `json:"name"`
}

func UserResponseFrom(u *entity.User) UserResponse {
	return UserResponse{
		ID:   u.ID.String(),
		TgID: u.TgID,
		Name: u.Name,
	}
}
