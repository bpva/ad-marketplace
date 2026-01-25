package entity

import (
	"time"

	"github.com/google/uuid"
)

type Channel struct {
	ID                uuid.UUID  `db:"id"`
	TelegramChannelID int64      `db:"telegram_channel_id"`
	Title             string     `db:"title"`
	Username          *string    `db:"username"`
	CreatedAt         time.Time  `db:"created_at"`
	DeletedAt         *time.Time `db:"deleted_at"`
}

type ChannelRole struct {
	ChannelID uuid.UUID `db:"channel_id"`
	UserID    uuid.UUID `db:"user_id"`
	Role      string    `db:"role"`
	CreatedAt time.Time `db:"created_at"`
}
