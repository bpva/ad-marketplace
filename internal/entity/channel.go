package entity

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

type Channel struct {
	ID          uuid.UUID  `db:"id"`
	TgChannelID int64      `db:"telegram_channel_id"`
	Title       string     `db:"title"`
	Username    *string    `db:"username"`
	IsListed    bool       `db:"is_listed"`
	CreatedAt   time.Time  `db:"created_at"`
	DeletedAt   *time.Time `db:"deleted_at"`
}

type ChannelRoleType string

const (
	ChannelRoleTypeUndefined ChannelRoleType = "undefined"
	ChannelRoleTypeOwner     ChannelRoleType = "owner"
	ChannelRoleTypeManager   ChannelRoleType = "manager"
)

func (r *ChannelRoleType) Scan(src any) error {
	switch v := src.(type) {
	case string:
		switch v {
		case "owner":
			*r = ChannelRoleTypeOwner
		case "manager":
			*r = ChannelRoleTypeManager
		default:
			*r = ChannelRoleTypeUndefined
		}
	case []byte:
		return r.Scan(string(v))
	default:
		return fmt.Errorf("cannot scan %T into ChannelRoleType", src)
	}
	return nil
}

type ChannelRole struct {
	ChannelID uuid.UUID       `db:"channel_id"`
	UserID    uuid.UUID       `db:"user_id"`
	Role      ChannelRoleType `db:"role"`
	CreatedAt time.Time       `db:"created_at"`
}

type AdFormatType string

const (
	AdFormatTypePost   AdFormatType = "post"
	AdFormatTypeRepost AdFormatType = "repost"
	AdFormatTypeStory  AdFormatType = "story"
)

type ChannelAdFormat struct {
	ID           uuid.UUID    `db:"id"`
	ChannelID    uuid.UUID    `db:"channel_id"`
	FormatType   AdFormatType `db:"format_type"`
	IsNative     bool         `db:"is_native"`
	FeedHours    int          `db:"feed_hours"`
	TopHours     int          `db:"top_hours"`
	PriceNanoTON int64        `db:"price_nano_ton"`
	CreatedAt    time.Time    `db:"created_at"`
}
