package entity

import (
	"time"

	"github.com/google/uuid"
)

type ChannelInfo struct {
	ChannelID               uuid.UUID `db:"channel_id"`
	About                   string    `db:"about"`
	Subscribers             int       `db:"subscribers"`
	LinkedChatID            *int64    `db:"linked_chat_id"`
	Languages               []byte    `db:"languages"`
	TopHours                []byte    `db:"top_hours"`
	ReactionsByEmotion      []byte    `db:"reactions_by_emotion"`
	StoryReactionsByEmotion []byte    `db:"story_reactions_by_emotion"`
	RecentPosts             []byte    `db:"recent_posts"`
	FetchedAt               time.Time `db:"fetched_at"`
}

type ChannelHistoricalStats struct {
	ChannelID uuid.UUID `db:"channel_id"`
	Date      time.Time `db:"date"`
	Data      []byte    `db:"data"`
}
