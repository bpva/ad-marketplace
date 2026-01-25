package dto

type UserResponse struct {
	ID   string `json:"id"`
	TgID int64  `json:"telegram_id"`
	Name string `json:"name"`
}
