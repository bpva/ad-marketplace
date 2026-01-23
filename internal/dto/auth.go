package dto

import "github.com/golang-jwt/jwt/v5"

type Claims struct {
	jwt.RegisteredClaims
	TelegramID int64 `json:"tid"`
}

type AuthRequest struct {
	InitData string `json:"init_data"`
}

type AuthResponse struct {
	Token string       `json:"token"`
	User  UserResponse `json:"user"`
}
