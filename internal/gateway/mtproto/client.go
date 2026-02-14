package mtproto

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"math"
	"sort"
	"time"

	"github.com/gotd/td/session"
	"github.com/gotd/td/telegram"
	"github.com/gotd/td/tg"

	"github.com/bpva/ad-marketplace/internal/config"
	"github.com/bpva/ad-marketplace/internal/dto"
	"github.com/bpva/ad-marketplace/internal/entity"
	"github.com/bpva/ad-marketplace/internal/logx"
)

type telegramChart struct {
	Columns []json.RawMessage `json:"columns"`
	Names   map[string]string `json:"names"`
}

type gateway struct {
	client     *telegram.Client
	api        *tg.Client
	userClient *telegram.Client
	userAPI    *tg.Client
	log        *slog.Logger
}

func New(ctx context.Context, cfg config.Telegram, log *slog.Logger) (*gateway, error) {
	if cfg.APIId == 0 || cfg.APIHash == "" {
		return nil, fmt.Errorf("TG_API_ID or TG_API_HASH not set")
	}

	log = log.With(logx.Service("mtproto"))
	client := telegram.NewClient(cfg.APIId, cfg.APIHash, telegram.Options{})
	userClient := telegram.NewClient(cfg.APIId, cfg.APIHash, telegram.Options{
		SessionStorage: &session.FileStorage{Path: ".session.json"},
	})

	c := &gateway{
		client:     client,
		userClient: userClient,
		log:        log,
	}

	botReady := make(chan struct{})
	botErrCh := make(chan error, 1)
	go func() {
		botErrCh <- client.Run(ctx, func(ctx context.Context) error {
			status, err := client.Auth().Status(ctx)
			if err != nil {
				return fmt.Errorf("auth status: %w", err)
			}

			if !status.Authorized {
				if _, err := client.Auth().Bot(ctx, cfg.BotToken); err != nil {
					return fmt.Errorf("bot auth: %w", err)
				}
			}

			c.api = client.API()
			close(botReady)

			<-ctx.Done()
			return ctx.Err()
		})
	}()

	userReady := make(chan struct{})
	userErrCh := make(chan error, 1)
	go func() {
		userErrCh <- userClient.Run(ctx, func(ctx context.Context) error {
			status, err := userClient.Auth().Status(ctx)
			if err != nil {
				return fmt.Errorf("user auth status: %w", err)
			}
			if !status.Authorized {
				return fmt.Errorf("user session is not authorized in .session.json")
			}

			c.userAPI = userClient.API()
			close(userReady)

			<-ctx.Done()
			return ctx.Err()
		})
	}()

	botConnected := false
	userConnected := false
	for !botConnected || !userConnected {
		select {
		case <-botReady:
			botConnected = true
		case <-userReady:
			userConnected = true
		case err := <-botErrCh:
			return nil, err
		case err := <-userErrCh:
			return nil, err
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}

	c.log.Info("connected")
	return c, nil
}

func (c *gateway) Ping(ctx context.Context) error {
	_, err := c.client.API().HelpGetConfig(ctx)
	return err
}

func BotAPIToMTProto(botAPIID int64) int64 {
	if botAPIID < -1_000_000_000_000 {
		return -botAPIID - 1_000_000_000_000
	}
	return botAPIID
}

func (c *gateway) resolveChannel(
	ctx context.Context,
	api *tg.Client,
	channelID int64,
) (int64, error) {
	res, err := api.ChannelsGetChannels(ctx, []tg.InputChannelClass{
		&tg.InputChannel{ChannelID: channelID, AccessHash: 0},
	})
	if err != nil {
		return 0, fmt.Errorf("resolve channel %d: %w", channelID, err)
	}

	chats, ok := res.(*tg.MessagesChats)
	if !ok {
		return 0, fmt.Errorf("unexpected response type: %T", res)
	}

	for _, ch := range chats.Chats {
		if channel, ok := ch.(*tg.Channel); ok && channel.ID == channelID {
			return channel.AccessHash, nil
		}
	}
	return 0, fmt.Errorf("channel %d not found", channelID)
}

func (c *gateway) GetChannelFull(
	ctx context.Context,
	channelID int64,
) (*dto.ChannelFullInfo, error) {
	accessHash, err := c.resolveChannel(ctx, c.api, channelID)
	if err != nil {
		return nil, err
	}

	full, err := c.api.ChannelsGetFullChannel(ctx, &tg.InputChannel{
		ChannelID:  channelID,
		AccessHash: accessHash,
	})
	if err != nil {
		return nil, fmt.Errorf("get full channel: %w", err)
	}

	channelFull, ok := full.FullChat.(*tg.ChannelFull)
	if !ok {
		return nil, fmt.Errorf("unexpected chat type: %T", full.FullChat)
	}

	info := &dto.ChannelFullInfo{
		About:        channelFull.GetAbout(),
		CanViewStats: channelFull.CanViewStats,
	}

	if v, ok := channelFull.GetParticipantsCount(); ok {
		info.ParticipantsCount = v
	}
	if v, ok := channelFull.GetLinkedChatID(); ok {
		info.LinkedChatID = v
	}
	if v, ok := channelFull.GetAdminsCount(); ok {
		info.AdminsCount = v
	}
	if v, ok := channelFull.GetOnlineCount(); ok {
		info.OnlineCount = v
	}
	if v, ok := channelFull.GetStatsDC(); ok {
		info.StatsDC = v
	}

	return info, nil
}

func (c *gateway) GetBroadcastStats(
	ctx context.Context,
	channelID int64,
	statsDC int,
) (*entity.BroadcastStats, error) {
	api := c.userAPI

	if statsDC > 0 {
		dc, err := c.userClient.DC(ctx, statsDC, 0)
		if err != nil {
			return nil, fmt.Errorf("connect to stats DC %d: %w", statsDC, err)
		}
		api = tg.NewClient(dc)
	}

	accessHash, err := c.resolveChannel(ctx, api, channelID)
	if err != nil {
		return nil, err
	}

	stats, err := api.StatsGetBroadcastStats(ctx, &tg.StatsGetBroadcastStatsRequest{
		Channel: &tg.InputChannel{
			ChannelID:  channelID,
			AccessHash: accessHash,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("get broadcast stats: %w", err)
	}

	graphJSON := map[string]json.RawMessage{
		"growth": c.extractGraphJSON(ctx, api, stats.GrowthGraph, "growth_graph"),
		"followers": c.extractGraphJSON(
			ctx,
			api,
			stats.FollowersGraph,
			"followers_graph",
		),
		"mute": c.extractGraphJSON(ctx, api, stats.MuteGraph, "mute_graph"),
		"interactions": c.extractGraphJSON(
			ctx,
			api,
			stats.InteractionsGraph,
			"interactions_graph",
		),
		"iv_interactions": c.extractGraphJSON(
			ctx,
			api,
			stats.IvInteractionsGraph,
			"iv_interactions_graph",
		),
		"views_by_source": c.extractGraphJSON(
			ctx,
			api,
			stats.ViewsBySourceGraph,
			"views_by_source_graph",
		),
		"followers_by_source": c.extractGraphJSON(
			ctx,
			api,
			stats.NewFollowersBySourceGraph,
			"new_followers_by_source_graph",
		),
		"story_interactions": c.extractGraphJSON(
			ctx,
			api,
			stats.StoryInteractionsGraph,
			"story_interactions_graph",
		),
		"languages": c.extractGraphJSON(
			ctx,
			api,
			stats.LanguagesGraph,
			"languages_graph",
		),
		"top_hours": c.extractGraphJSON(
			ctx,
			api,
			stats.TopHoursGraph,
			"top_hours_graph",
		),
		"reactions_by_emotion": c.extractGraphJSON(
			ctx,
			api,
			stats.ReactionsByEmotionGraph,
			"reactions_by_emotion_graph",
		),
		"story_reactions_by_emotion": c.extractGraphJSON(
			ctx,
			api,
			stats.StoryReactionsByEmotionGraph,
			"story_reactions_by_emotion_graph",
		),
	}

	daily := make(map[time.Time]*entity.ChannelHistoricalDayData)

	c.mergeSingleSeries(
		daily,
		graphJSON["growth"],
		func(d *entity.ChannelHistoricalDayData, v float64) {
			value := roundToInt64(v)
			d.Subscribers = &value
		},
	)
	c.mergeSingleSeries(
		daily,
		graphJSON["followers"],
		func(d *entity.ChannelHistoricalDayData, v float64) {
			value := roundToInt64(v)
			d.NewFollowers = &value
		},
	)
	c.mergeSingleSeries(
		daily,
		graphJSON["mute"],
		func(d *entity.ChannelHistoricalDayData, v float64) {
			value := v
			d.MutePct = &value
		},
	)
	c.mergeSingleSeries(
		daily,
		graphJSON["interactions"],
		func(d *entity.ChannelHistoricalDayData, v float64) {
			value := roundToInt64(v)
			d.Interactions = &value
		},
	)
	c.mergeSingleSeries(
		daily,
		graphJSON["iv_interactions"],
		func(d *entity.ChannelHistoricalDayData, v float64) {
			value := roundToInt64(v)
			d.IVInteractions = &value
		},
	)

	c.mergeMultiSeries(
		daily,
		graphJSON["views_by_source"],
		func(d *entity.ChannelHistoricalDayData, key string, v float64) {
			if d.ViewsBySource == nil {
				d.ViewsBySource = make(map[string]int64)
			}
			d.ViewsBySource[key] = roundToInt64(v)
		},
	)
	c.mergeMultiSeries(
		daily,
		graphJSON["followers_by_source"],
		func(d *entity.ChannelHistoricalDayData, key string, v float64) {
			if d.FollowersBySource == nil {
				d.FollowersBySource = make(map[string]int64)
			}
			d.FollowersBySource[key] = roundToInt64(v)
		},
	)
	c.mergeMultiSeries(
		daily,
		graphJSON["story_interactions"],
		func(d *entity.ChannelHistoricalDayData, key string, v float64) {
			if d.StoryInteractions == nil {
				d.StoryInteractions = make(map[string]int64)
			}
			d.StoryInteractions[key] = roundToInt64(v)
		},
	)
	c.mergeMultiSeries(
		daily,
		graphJSON["reactions_by_emotion"],
		func(d *entity.ChannelHistoricalDayData, key string, v float64) {
			if d.ReactionsByEmotion == nil {
				d.ReactionsByEmotion = make(map[string]int64)
			}
			d.ReactionsByEmotion[key] = roundToInt64(v)
		},
	)
	c.mergeMultiSeries(
		daily,
		graphJSON["story_reactions_by_emotion"],
		func(d *entity.ChannelHistoricalDayData, key string, v float64) {
			if d.StoryReactionsByEmotion == nil {
				d.StoryReactionsByEmotion = make(map[string]int64)
			}
			d.StoryReactionsByEmotion[key] = roundToInt64(v)
		},
	)

	return &entity.BroadcastStats{
		DailyStats:              dailyToSlice(daily),
		Languages:               parseLanguagesChart(graphJSON["languages"]),
		TopHours:                parseTopHoursChart(graphJSON["top_hours"]),
		ReactionsByEmotion:      parseSeriesTotals(graphJSON["reactions_by_emotion"]),
		StoryReactionsByEmotion: parseSeriesTotals(graphJSON["story_reactions_by_emotion"]),
		RecentPosts:             parseRecentPosts(stats.RecentPostsInteractions),
	}, nil
}

func (c *gateway) extractGraphJSON(
	ctx context.Context,
	api *tg.Client,
	g tg.StatsGraphClass,
	metric string,
) json.RawMessage {
	const maxAsyncDepth = 3
	for depth := 0; depth <= maxAsyncDepth; depth++ {
		switch v := g.(type) {
		case *tg.StatsGraph:
			if v == nil {
				return nil
			}
			return json.RawMessage(v.JSON.Data)
		case *tg.StatsGraphAsync:
			if v == nil || v.Token == "" {
				return nil
			}
			loaded, err := api.StatsLoadAsyncGraph(
				ctx,
				&tg.StatsLoadAsyncGraphRequest{Token: v.Token},
			)
			if err != nil {
				c.log.Warn("failed to load async graph", "metric", metric, "error", err)
				return nil
			}
			g = loaded
		case *tg.StatsGraphError:
			if v == nil {
				return nil
			}
			c.log.Warn("graph is unavailable", "metric", metric, "error", v.Error)
			return nil
		default:
			return nil
		}
	}

	c.log.Warn("async graph depth exceeded", "metric", metric)
	return nil
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
