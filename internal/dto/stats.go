package dto

import (
	"encoding/json"
	"time"
)

type ChannelFullInfo struct {
	ParticipantsCount int    `json:"participants_count"`
	LinkedChatID      int64  `json:"linked_chat_id,omitempty"`
	AdminsCount       int    `json:"admins_count"`
	OnlineCount       int    `json:"online_count"`
	About             string `json:"about"`
	CanViewStats      bool   `json:"can_view_stats"`
	StatsDC           int    `json:"stats_dc"`
}

type BroadcastStats struct {
	Period               StatsPeriod     `json:"period"`
	Followers            StatsValue      `json:"followers"`
	ViewsPerPost         StatsValue      `json:"views_per_post"`
	SharesPerPost        StatsValue      `json:"shares_per_post"`
	ReactionsPerPost     StatsValue      `json:"reactions_per_post"`
	ViewsPerStory        StatsValue      `json:"views_per_story"`
	SharesPerStory       StatsValue      `json:"shares_per_story"`
	ReactionsPerStory    StatsValue      `json:"reactions_per_story"`
	EnabledNotifications StatsPercentage `json:"enabled_notifications"`
}

type StatsPeriod struct {
	MinDate int `json:"min_date"`
	MaxDate int `json:"max_date"`
}

type StatsValue struct {
	Current  float64 `json:"current"`
	Previous float64 `json:"previous"`
}

type StatsPercentage struct {
	Part  float64 `json:"part"`
	Total float64 `json:"total"`
}

type BroadcastStatsResult struct {
	Scalars                 BroadcastStats
	DailyStats              map[string]map[string]any
	Languages               json.RawMessage
	TopHours                json.RawMessage
	ReactionsByEmotion      json.RawMessage
	StoryReactionsByEmotion json.RawMessage
	RecentPosts             json.RawMessage
}

type ChannelInfoResponse struct {
	About                   string          `json:"about"`
	Subscribers             int             `json:"subscribers"`
	Languages               json.RawMessage `json:"languages,omitempty" swaggertype:"object"`
	TopHours                json.RawMessage `json:"top_hours,omitempty" swaggertype:"object"`
	ReactionsByEmotion      json.RawMessage `json:"reactions_by_emotion,omitempty" swaggertype:"object"`
	StoryReactionsByEmotion json.RawMessage `json:"story_reactions_by_emotion,omitempty" swaggertype:"object"`
	RecentPosts             json.RawMessage `json:"recent_posts,omitempty" swaggertype:"object"`
	FetchedAt               time.Time       `json:"fetched_at"`
}

type ChannelHistoricalStatsResponse struct {
	Stats []ChannelDailyStats `json:"stats"`
}

type ChannelDailyStats struct {
	Date string          `json:"date"`
	Data json.RawMessage `json:"data" swaggertype:"object"`
}
