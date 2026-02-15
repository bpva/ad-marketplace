package mtproto

import (
	"context"
	"encoding/json"
	"math"
	"sort"
	"time"

	"github.com/gotd/td/tg"

	"github.com/bpva/ad-marketplace/internal/entity"
)

type telegramChart struct {
	Columns []json.RawMessage `json:"columns"`
	Names   map[string]string `json:"names"`
}

func (c *gateway) parseBroadcastStats(
	ctx context.Context,
	api *tg.Client,
	stats *tg.StatsBroadcastStats,
) *entity.BroadcastStats {
	g := func(graph tg.StatsGraphClass, metric string) json.RawMessage {
		return c.extractGraphJSON(ctx, api, graph, metric)
	}

	graphs := map[string]json.RawMessage{
		"growth":                     g(stats.GrowthGraph, "growth_graph"),
		"followers":                  g(stats.FollowersGraph, "followers_graph"),
		"mute":                       g(stats.MuteGraph, "mute_graph"),
		"interactions":               g(stats.InteractionsGraph, "interactions_graph"),
		"iv_interactions":            g(stats.IvInteractionsGraph, "iv_interactions_graph"),
		"views_by_source":            g(stats.ViewsBySourceGraph, "views_by_source_graph"),
		"followers_by_source":        g(stats.NewFollowersBySourceGraph, "new_followers_by_source_graph"),
		"story_interactions":         g(stats.StoryInteractionsGraph, "story_interactions_graph"),
		"languages":                  g(stats.LanguagesGraph, "languages_graph"),
		"top_hours":                  g(stats.TopHoursGraph, "top_hours_graph"),
		"reactions_by_emotion":       g(stats.ReactionsByEmotionGraph, "reactions_by_emotion_graph"),
		"story_reactions_by_emotion": g(stats.StoryReactionsByEmotionGraph, "story_reactions_by_emotion_graph"),
	}

	// dump graphs in data.json
	// data, _ := json.MarshalIndent(graphs, "", "  ")
	// _ = os.WriteFile("data.json", data, 0644)

	daily := make(map[time.Time]*entity.ChannelHistoricalDayData)

	c.mergeSingleSeries(
		daily,
		graphs["growth"],
		func(d *entity.ChannelHistoricalDayData, v float64) {
			value := roundToInt64(v)
			d.Subscribers = &value
		},
	)
	c.mergeSingleSeries(
		daily,
		graphs["followers"],
		func(d *entity.ChannelHistoricalDayData, v float64) {
			value := roundToInt64(v)
			d.NewFollowers = &value
		},
	)
	c.mergeSingleSeries(
		daily,
		graphs["mute"],
		func(d *entity.ChannelHistoricalDayData, v float64) {
			value := v
			d.MutePct = &value
		},
	)
	c.mergeSingleSeries(
		daily,
		graphs["interactions"],
		func(d *entity.ChannelHistoricalDayData, v float64) {
			value := roundToInt64(v)
			d.Interactions = &value
		},
	)
	c.mergeSingleSeries(
		daily,
		graphs["iv_interactions"],
		func(d *entity.ChannelHistoricalDayData, v float64) {
			value := roundToInt64(v)
			d.IVInteractions = &value
		},
	)

	c.mergeMultiSeries(
		daily,
		graphs["views_by_source"],
		func(d *entity.ChannelHistoricalDayData, key string, v float64) {
			if d.ViewsBySource == nil {
				d.ViewsBySource = make(map[string]int64)
			}
			d.ViewsBySource[key] = roundToInt64(v)
		},
	)
	c.mergeMultiSeries(
		daily,
		graphs["followers_by_source"],
		func(d *entity.ChannelHistoricalDayData, key string, v float64) {
			if d.FollowersBySource == nil {
				d.FollowersBySource = make(map[string]int64)
			}
			d.FollowersBySource[key] = roundToInt64(v)
		},
	)
	c.mergeMultiSeries(
		daily,
		graphs["story_interactions"],
		func(d *entity.ChannelHistoricalDayData, key string, v float64) {
			if d.StoryInteractions == nil {
				d.StoryInteractions = make(map[string]int64)
			}
			d.StoryInteractions[key] = roundToInt64(v)
		},
	)
	c.mergeMultiSeries(
		daily,
		graphs["reactions_by_emotion"],
		func(d *entity.ChannelHistoricalDayData, key string, v float64) {
			if d.ReactionsByEmotion == nil {
				d.ReactionsByEmotion = make(map[string]int64)
			}
			d.ReactionsByEmotion[key] = roundToInt64(v)
		},
	)
	c.mergeMultiSeries(
		daily,
		graphs["story_reactions_by_emotion"],
		func(d *entity.ChannelHistoricalDayData, key string, v float64) {
			if d.StoryReactionsByEmotion == nil {
				d.StoryReactionsByEmotion = make(map[string]int64)
			}
			d.StoryReactionsByEmotion[key] = roundToInt64(v)
		},
	)

	return &entity.BroadcastStats{
		DailyStats:              dailyToSlice(daily),
		Languages:               parseLanguagesChart(graphs["languages"]),
		TopHours:                parseTopHoursChart(graphs["top_hours"]),
		ReactionsByEmotion:      parseSeriesTotals(graphs["reactions_by_emotion"]),
		StoryReactionsByEmotion: parseSeriesTotals(graphs["story_reactions_by_emotion"]),
		RecentPosts:             parseRecentPosts(stats.RecentPostsInteractions),
	}
}

func tsToDate(ms int64) time.Time {
	return time.Unix(ms/1000, 0).UTC().Truncate(24 * time.Hour)
}

func ensureDaily(
	daily map[time.Time]*entity.ChannelHistoricalDayData,
	date time.Time,
) *entity.ChannelHistoricalDayData {
	if daily[date] == nil {
		daily[date] = &entity.ChannelHistoricalDayData{}
	}
	return daily[date]
}

func (c *gateway) mergeSingleSeries(
	daily map[time.Time]*entity.ChannelHistoricalDayData,
	raw json.RawMessage,
	assign func(*entity.ChannelHistoricalDayData, float64),
) {
	timestamps, values := parseSingleSeries(raw)
	if timestamps == nil {
		return
	}

	for i, ts := range timestamps {
		if i >= len(values) {
			break
		}
		assign(ensureDaily(daily, tsToDate(ts)), values[i])
	}
}

func (c *gateway) mergeMultiSeries(
	daily map[time.Time]*entity.ChannelHistoricalDayData,
	raw json.RawMessage,
	assign func(*entity.ChannelHistoricalDayData, string, float64),
) {
	timestamps, series := parseMultiSeries(raw)
	if timestamps == nil || len(series) == 0 {
		return
	}

	for key, values := range series {
		for i, ts := range timestamps {
			if i >= len(values) {
				break
			}
			assign(ensureDaily(daily, tsToDate(ts)), key, values[i])
		}
	}
}

func parseSingleSeries(raw json.RawMessage) ([]int64, []float64) {
	var chart telegramChart
	if err := json.Unmarshal(raw, &chart); err != nil || len(chart.Columns) < 2 {
		return nil, nil
	}

	timestamps := parseTimestampColumn(chart.Columns[0])
	if timestamps == nil {
		return nil, nil
	}

	_, values := parseNamedValueColumn(chart.Columns[1])
	if values == nil {
		return nil, nil
	}

	return timestamps, values
}

func parseMultiSeries(raw json.RawMessage) ([]int64, map[string][]float64) {
	var chart telegramChart
	if err := json.Unmarshal(raw, &chart); err != nil || len(chart.Columns) < 2 {
		return nil, nil
	}

	timestamps := parseTimestampColumn(chart.Columns[0])
	if timestamps == nil {
		return nil, nil
	}

	series := make(map[string][]float64)
	for colIdx := 1; colIdx < len(chart.Columns); colIdx++ {
		seriesID, values := parseNamedValueColumn(chart.Columns[colIdx])
		if seriesID == "" || values == nil {
			continue
		}
		name := chart.Names[seriesID]
		if name == "" {
			name = seriesID
		}
		series[name] = values
	}
	if len(series) == 0 {
		return nil, nil
	}

	return timestamps, series
}

func parseSeriesTotals(raw json.RawMessage) map[string]int64 {
	_, series := parseMultiSeries(raw)
	if len(series) == 0 {
		return nil
	}

	totals := make(map[string]int64, len(series))
	for name, values := range series {
		var total float64
		for _, v := range values {
			total += v
		}
		totals[name] = roundToInt64(total)
	}

	return totals
}

func parseLanguagesChart(raw json.RawMessage) []entity.LanguageShare {
	_, series := parseMultiSeries(raw)
	if len(series) == 0 {
		return nil
	}

	result := make([]entity.LanguageShare, 0, len(series))
	for lang, values := range series {
		if len(values) == 0 {
			continue
		}
		result = append(result, entity.LanguageShare{
			Language:   lang,
			Percentage: values[len(values)-1],
		})
	}
	return result
}

func parseTopHoursChart(raw json.RawMessage) []float64 {
	_, values := parseSingleSeries(raw)
	return values
}

func dailyToSlice(daily map[time.Time]*entity.ChannelHistoricalDayData) []entity.DailyMetrics {
	dates := make([]time.Time, 0, len(daily))
	for date := range daily {
		dates = append(dates, date)
	}
	sort.Slice(dates, func(i, j int) bool { return dates[i].Before(dates[j]) })

	result := make([]entity.DailyMetrics, 0, len(dates))
	for _, date := range dates {
		result = append(result, entity.DailyMetrics{
			Date: date,
			Data: *daily[date],
		})
	}
	return result
}

func parseTimestampColumn(raw json.RawMessage) []int64 {
	var col []json.RawMessage
	if err := json.Unmarshal(raw, &col); err != nil || len(col) < 2 {
		return nil
	}
	timestamps := make([]int64, 0, len(col)-1)
	for _, item := range col[1:] {
		var ts int64
		if err := json.Unmarshal(item, &ts); err != nil {
			return nil
		}
		timestamps = append(timestamps, ts)
	}
	return timestamps
}

func parseNamedValueColumn(raw json.RawMessage) (string, []float64) {
	var col []json.RawMessage
	if err := json.Unmarshal(raw, &col); err != nil || len(col) < 2 {
		return "", nil
	}
	var name string
	if err := json.Unmarshal(col[0], &name); err != nil {
		return "", nil
	}
	values := make([]float64, 0, len(col)-1)
	for _, item := range col[1:] {
		var v float64
		if err := json.Unmarshal(item, &v); err != nil {
			return "", nil
		}
		values = append(values, v)
	}
	return name, values
}

func roundToInt64(v float64) int64 {
	return int64(math.Round(v))
}

func parseRecentPosts(posts []tg.PostInteractionCountersClass) []entity.RecentPost {
	if len(posts) == 0 {
		return nil
	}

	result := make([]entity.RecentPost, 0, len(posts))
	for _, p := range posts {
		switch v := p.(type) {
		case *tg.PostInteractionCountersMessage:
			result = append(result, entity.RecentPost{
				Type:      "message",
				ID:        v.MsgID,
				Views:     v.Views,
				Forwards:  v.Forwards,
				Reactions: v.Reactions,
			})
		case *tg.PostInteractionCountersStory:
			result = append(result, entity.RecentPost{
				Type:      "story",
				ID:        v.StoryID,
				Views:     v.Views,
				Forwards:  v.Forwards,
				Reactions: v.Reactions,
			})
		}
	}
	return result
}
