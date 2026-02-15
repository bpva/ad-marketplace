package entity

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

type Channel struct {
	ID               uuid.UUID  `db:"id"`
	TgChannelID      int64      `db:"telegram_channel_id"`
	Title            string     `db:"title"`
	Username         *string    `db:"username"`
	IsListed         bool       `db:"is_listed"`
	PhotoSmallFileID *string    `db:"photo_small_file_id"`
	PhotoBigFileID   *string    `db:"photo_big_file_id"`
	CreatedAt        time.Time  `db:"created_at"`
	DeletedAt        *time.Time `db:"deleted_at"`
}

type MVChannel struct {
	ChannelID               uuid.UUID         `db:"channel_id"`
	TgChannelID             int64             `db:"telegram_channel_id"`
	Title                   string            `db:"title"`
	Username                *string           `db:"username"`
	PhotoSmallFileID        *string           `db:"photo_small_file_id"`
	PhotoBigFileID          *string           `db:"photo_big_file_id"`
	About                   string            `db:"about"`
	Subscribers             *int              `db:"subscribers"`
	LinkedChatID            *int64            `db:"linked_chat_id"`
	Languages               []LanguageShare   `db:"languages"`
	TopHours                []float64         `db:"top_hours"`
	ReactionsByEmotion      map[string]int    `db:"reactions_by_emotion"`
	StoryReactionsByEmotion map[string]int    `db:"story_reactions_by_emotion"`
	RecentPosts             []byte            `db:"recent_posts"`
	AdFormats               []ChannelAdFormat `db:"ad_formats"`
	AvgDailyViews1d         *int              `db:"avg_daily_views_1d"`
	AvgDailyViews7d         *int              `db:"avg_daily_views_7d"`
	AvgDailyViews30d        *int              `db:"avg_daily_views_30d"`
	TotalViews7d            *int              `db:"total_views_7d"`
	TotalViews30d           *int              `db:"total_views_30d"`
	SubGrowth7d             *int              `db:"sub_growth_7d"`
	SubGrowth30d            *int              `db:"sub_growth_30d"`
	AvgInteractions7d       *int              `db:"avg_interactions_7d"`
	AvgInteractions30d      *int              `db:"avg_interactions_30d"`
	EngagementRate7d        *float64          `db:"engagement_rate_7d"`
	EngagementRate30d       *float64          `db:"engagement_rate_30d"`
}

type LanguageShare struct {
	Language   string  `json:"language"`
	Percentage float64 `json:"percentage"`
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
	ID           uuid.UUID    `db:"id" json:"id"`
	ChannelID    uuid.UUID    `db:"channel_id" json:"channel_id"`
	FormatType   AdFormatType `db:"format_type" json:"format_type"`
	IsNative     bool         `db:"is_native" json:"is_native"`
	FeedHours    int          `db:"feed_hours" json:"feed_hours"`
	TopHours     int          `db:"top_hours" json:"top_hours"`
	PriceNanoTON int64        `db:"price_nano_ton" json:"price_nano_ton"`
	CreatedAt    time.Time    `db:"created_at" json:"created_at"`
}
