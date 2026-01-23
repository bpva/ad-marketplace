package entity

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID         uuid.UUID  `db:"id"`
	TelegramID int64      `db:"telegram_id"`
	Name       string     `db:"name"`
	CreatedAt  time.Time  `db:"created_at"`
	DeletedAt  *time.Time `db:"deleted_at"`
}
