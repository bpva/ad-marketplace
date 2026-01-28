package dto

import "github.com/bpva/ad-marketplace/internal/entity"

type UserResponse struct {
	TgID int64  `json:"telegram_id"`
	Name string `json:"name"`
}

func UserResponseFrom(u *entity.User) UserResponse {
	return UserResponse{
		TgID: u.TgID,
		Name: u.Name,
	}
}
