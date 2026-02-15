package entity

import (
	"time"

	"github.com/google/uuid"
)

type RecentPost struct {
	Type      string `json:"type"`
	ID        int    `json:"id"`
	Views     int    `json:"views"`
	Forwards  int    `json:"forwards"`
	Reactions int    `json:"reactions"`
}

type ChannelHistoricalDayData struct {
	Subscribers             *int64           `json:"subscribers,omitempty"`
	NewFollowers            *int64           `json:"new_followers,omitempty"`
	MutePct                 *float64         `json:"mute_pct,omitempty"`
	Interactions            *int64           `json:"interactions,omitempty"`
	IVInteractions          *int64           `json:"iv_interactions,omitempty"`
	ViewsBySource           map[string]int64 `json:"views_by_source,omitempty"`
	FollowersBySource       map[string]int64 `json:"followers_by_source,omitempty"`
	StoryInteractions       map[string]int64 `json:"story_interactions,omitempty"`
	ReactionsByEmotion      map[string]int64 `json:"reactions_by_emotion,omitempty"`
	StoryReactionsByEmotion map[string]int64 `json:"story_reactions_by_emotion,omitempty"`
}

type DailyMetrics struct {
	Date time.Time
	Data ChannelHistoricalDayData
}

type BroadcastStats struct {
	DailyStats              []DailyMetrics
	Languages               []LanguageShare
	TopHours                []float64
	ReactionsByEmotion      map[string]int64
	StoryReactionsByEmotion map[string]int64
	RecentPosts             []RecentPost
}

type ChannelInfo struct {
	ChannelID               uuid.UUID        `db:"channel_id"`
	About                   string           `db:"about"`
	Subscribers             int              `db:"subscribers"`
	LinkedChatID            *int64           `db:"linked_chat_id"`
	Languages               []LanguageShare  `db:"languages"`
	TopHours                []float64        `db:"top_hours"`
	ReactionsByEmotion      map[string]int64 `db:"reactions_by_emotion"`
	StoryReactionsByEmotion map[string]int64 `db:"story_reactions_by_emotion"`
	RecentPosts             []RecentPost     `db:"recent_posts"`
	FetchedAt               time.Time        `db:"fetched_at"`
}
