package dto

import (
	"context"

	"github.com/google/uuid"
)

type ctxKey string

const userCtxKey ctxKey = "user"

type UserContext struct {
	ID         uuid.UUID
	TelegramID int64
}

func ContextWithUser(ctx context.Context, u UserContext) context.Context {
	return context.WithValue(ctx, userCtxKey, u)
}

func UserFromContext(ctx context.Context) (UserContext, bool) {
	user, ok := ctx.Value(userCtxKey).(UserContext)
	return user, ok
}
