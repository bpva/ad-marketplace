package dto

type UserResponse struct {
	ID         string `json:"id"`
	TelegramID int64  `json:"telegram_id"`
	Name       string `json:"name"`
}
