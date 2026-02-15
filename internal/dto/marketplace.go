package dto

import "github.com/bpva/ad-marketplace/internal/entity"

type MarketplaceFilter struct {
	Name  string `json:"name"`
	Value any    `json:"value"`
}

type MarketplaceChannelsRequest struct {
	Filters   []MarketplaceFilter  `json:"filters,omitempty"`
	SortBy    entity.ChannelSortBy `json:"sort_by,omitempty"`
	SortOrder entity.SortOrder     `json:"sort_order,omitempty"`
	Page      int                  `json:"page,omitempty"`
}

type MarketplaceChannel struct {
	TgChannelID   int64  `json:"id"`
	Title         string `json:"title"`
	Username      string `json:"username,omitempty"`
	PhotoSmallURL string `json:"photo_small_url,omitempty"`
	About         string `json:"about,omitempty"`
	Subscribers   *int   `json:"subscribers,omitempty"`
	// ISO 639-1 codes: "en", "ru"
	Languages []entity.LanguageShare `json:"languages,omitempty"`
	// Views per hour (index 0–23, UTC)
	TopHours []float64 `json:"top_hours,omitempty"`
	// Keys are unicode emoji ("👍"), custom will be mapped to standard too
	ReactionsByEmotion      map[string]int     `json:"reactions_by_emotion,omitempty"`
	StoryReactionsByEmotion map[string]int     `json:"story_reactions_by_emotion,omitempty"`
	AdFormats               []AdFormat         `json:"ad_formats"`
	Categories              []CategoryResponse `json:"categories,omitempty"`
	AvgDailyViews1d         *int               `json:"avg_daily_views_1d,omitempty"`
	AvgDailyViews7d         *int               `json:"avg_daily_views_7d,omitempty"`
	AvgDailyViews30d        *int               `json:"avg_daily_views_30d,omitempty"`
	TotalViews7d            *int               `json:"total_views_7d,omitempty"`
	TotalViews30d           *int               `json:"total_views_30d,omitempty"`
	SubGrowth7d             *int               `json:"sub_growth_7d,omitempty"`
	SubGrowth30d            *int               `json:"sub_growth_30d,omitempty"`
	AvgInteractions7d       *int               `json:"avg_interactions_7d,omitempty"`
	AvgInteractions30d      *int               `json:"avg_interactions_30d,omitempty"`
	EngagementRate7d        *float64           `json:"engagement_rate_7d,omitempty"`
	EngagementRate30d       *float64           `json:"engagement_rate_30d,omitempty"`
}

type AdFormat struct {
	ID           string              `json:"-"`
	FormatType   entity.AdFormatType `json:"format_type"`
	IsNative     bool                `json:"is_native"`
	FeedHours    int                 `json:"feed_hours"`
	TopHours     int                 `json:"top_hours"`
	PriceNanoTON int64               `json:"price_nano_ton"`
}

type MarketplaceChannelsResponse struct {
	Channels []MarketplaceChannel `json:"channels"`
	Total    int                  `json:"total"`
}
