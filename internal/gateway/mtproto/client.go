package mtproto

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/gotd/td/telegram"
	"github.com/gotd/td/tg"

	"github.com/bpva/ad-marketplace/internal/config"
	"github.com/bpva/ad-marketplace/internal/dto"
	"github.com/bpva/ad-marketplace/internal/logx"
)

type Client struct {
	client *telegram.Client
	api    *tg.Client
	log    *slog.Logger
}

func New(ctx context.Context, cfg config.Telegram, log *slog.Logger) (*Client, error) {
	if cfg.APIId == 0 || cfg.APIHash == "" {
		return nil, fmt.Errorf("TG_API_ID or TG_API_HASH not set")
	}

	log = log.With(logx.Service("mtproto"))
	client := telegram.NewClient(cfg.APIId, cfg.APIHash, telegram.Options{})

	c := &Client{
		client: client,
		log:    log,
	}

	ready := make(chan struct{})
	errCh := make(chan error, 1)
	go func() {
		errCh <- client.Run(ctx, func(ctx context.Context) error {
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
			close(ready)

			<-ctx.Done()
			return ctx.Err()
		})
	}()

	select {
	case <-ready:
		c.log.Info("connected")
		return c, nil
	case err := <-errCh:
		return nil, err
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

func (c *Client) Ping(ctx context.Context) error {
	_, err := c.client.API().HelpGetConfig(ctx)
	return err
}

func BotAPIToMTProto(botAPIID int64) int64 {
	if botAPIID < -1_000_000_000_000 {
		return -botAPIID - 1_000_000_000_000
	}
	return botAPIID
}

func (c *Client) ResolveChannel(ctx context.Context, channelID int64) (int64, error) {
	res, err := c.api.ChannelsGetChannels(ctx, []tg.InputChannelClass{
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

func (c *Client) GetChannelFull(
	ctx context.Context,
	channelID, accessHash int64,
) (*dto.ChannelFullInfo, error) {
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

func (c *Client) GetBroadcastStats(
	ctx context.Context,
	channelID, accessHash int64,
	statsDC int,
) (*dto.BroadcastStatsResult, error) {
	api := c.api

	if statsDC > 0 {
		dc, err := c.client.DC(ctx, statsDC, 0)
		if err != nil {
			return nil, fmt.Errorf("connect to stats DC %d: %w", statsDC, err)
		}
		api = tg.NewClient(dc)
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

	result := &dto.BroadcastStatsResult{
		Scalars: dto.BroadcastStats{
			Period: dto.StatsPeriod{
				MinDate: stats.Period.MinDate,
				MaxDate: stats.Period.MaxDate,
			},
			Followers:         absValueToDTO(stats.Followers),
			ViewsPerPost:      absValueToDTO(stats.ViewsPerPost),
			SharesPerPost:     absValueToDTO(stats.SharesPerPost),
			ReactionsPerPost:  absValueToDTO(stats.ReactionsPerPost),
			ViewsPerStory:     absValueToDTO(stats.ViewsPerStory),
			SharesPerStory:    absValueToDTO(stats.SharesPerStory),
			ReactionsPerStory: absValueToDTO(stats.ReactionsPerStory),
			EnabledNotifications: dto.StatsPercentage{
				Part:  stats.EnabledNotifications.Part,
				Total: stats.EnabledNotifications.Total,
			},
		},
		DailyStats: make(map[string]map[string]any),
	}

	timeSeriesGraphs := map[string]tg.StatsGraphClass{
		"subscribers":     stats.GrowthGraph,
		"new_followers":   stats.FollowersGraph,
		"mute_pct":        stats.MuteGraph,
		"interactions":    stats.InteractionsGraph,
		"iv_interactions": stats.IvInteractionsGraph,
	}
	for metric, graph := range timeSeriesGraphs {
		c.mergeTimeSeriesGraph(result.DailyStats, graph, metric)
	}

	multiSeriesGraphs := map[string]tg.StatsGraphClass{
		"views_by_source":     stats.ViewsBySourceGraph,
		"followers_by_source": stats.NewFollowersBySourceGraph,
	}
	for prefix, graph := range multiSeriesGraphs {
		c.mergeMultiSeriesGraph(result.DailyStats, graph, prefix)
	}

	result.Languages = extractGraphJSON(stats.LanguagesGraph)
	result.TopHours = extractGraphJSON(stats.TopHoursGraph)
	result.ReactionsByEmotion = extractGraphJSON(stats.ReactionsByEmotionGraph)
	result.StoryReactionsByEmotion = extractGraphJSON(stats.StoryReactionsByEmotionGraph)
	result.RecentPosts = marshalRecentPosts(stats.RecentPostsInteractions)

	return result, nil
}

func absValueToDTO(v tg.StatsAbsValueAndPrev) dto.StatsValue {
	return dto.StatsValue{Current: v.Current, Previous: v.Previous}
}

func extractGraphJSON(g tg.StatsGraphClass) json.RawMessage {
	sg, ok := g.(*tg.StatsGraph)
	if !ok || sg == nil {
		return nil
	}
	return json.RawMessage(sg.JSON.Data)
}

type chartData struct {
	Columns []json.RawMessage `json:"columns"`
}

func (c *Client) mergeTimeSeriesGraph(
	daily map[string]map[string]any,
	g tg.StatsGraphClass,
	metric string,
) {
	sg, ok := g.(*tg.StatsGraph)
	if !ok || sg == nil {
		return
	}

	var chart chartData
	if err := json.Unmarshal([]byte(sg.JSON.Data), &chart); err != nil {
		c.log.Warn("failed to parse graph JSON", "metric", metric, "error", err)
		return
	}

	if len(chart.Columns) < 2 {
		return
	}

	timestamps, values := parseTimestampValueColumns(chart.Columns[0], chart.Columns[1])
	if timestamps == nil {
		return
	}

	for i, ts := range timestamps {
		if i >= len(values) {
			break
		}
		date := time.Unix(ts/1000, 0).UTC().Format("2006-01-02")
		if daily[date] == nil {
			daily[date] = make(map[string]any)
		}
		daily[date][metric] = values[i]
	}
}

func (c *Client) mergeMultiSeriesGraph(
	daily map[string]map[string]any,
	g tg.StatsGraphClass,
	prefix string,
) {
	sg, ok := g.(*tg.StatsGraph)
	if !ok || sg == nil {
		return
	}

	var chart chartData
	if err := json.Unmarshal([]byte(sg.JSON.Data), &chart); err != nil {
		c.log.Warn("failed to parse graph JSON", "prefix", prefix, "error", err)
		return
	}

	if len(chart.Columns) < 2 {
		return
	}

	timestamps := parseTimestampColumn(chart.Columns[0])
	if timestamps == nil {
		return
	}

	for colIdx := 1; colIdx < len(chart.Columns); colIdx++ {
		name, vals := parseNamedValueColumn(chart.Columns[colIdx])
		if name == "" {
			continue
		}
		for i, ts := range timestamps {
			if i >= len(vals) {
				break
			}
			date := time.Unix(ts/1000, 0).UTC().Format("2006-01-02")
			if daily[date] == nil {
				daily[date] = make(map[string]any)
			}
			bySource, _ := daily[date][prefix].(map[string]any)
			if bySource == nil {
				bySource = make(map[string]any)
			}
			bySource[name] = vals[i]
			daily[date][prefix] = bySource
		}
	}
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

func parseTimestampValueColumns(tsRaw, valRaw json.RawMessage) ([]int64, []float64) {
	timestamps := parseTimestampColumn(tsRaw)
	if timestamps == nil {
		return nil, nil
	}
	_, values := parseNamedValueColumn(valRaw)
	return timestamps, values
}

func marshalRecentPosts(posts []tg.PostInteractionCountersClass) json.RawMessage {
	if len(posts) == 0 {
		return nil
	}

	type postData struct {
		Type      string `json:"type"`
		ID        int    `json:"id"`
		Views     int    `json:"views"`
		Forwards  int    `json:"forwards"`
		Reactions int    `json:"reactions"`
	}

	result := make([]postData, 0, len(posts))
	for _, p := range posts {
		switch v := p.(type) {
		case *tg.PostInteractionCountersMessage:
			result = append(result, postData{
				Type:      "message",
				ID:        v.MsgID,
				Views:     v.Views,
				Forwards:  v.Forwards,
				Reactions: v.Reactions,
			})
		case *tg.PostInteractionCountersStory:
			result = append(result, postData{
				Type:      "story",
				ID:        v.StoryID,
				Views:     v.Views,
				Forwards:  v.Forwards,
				Reactions: v.Reactions,
			})
		}
	}

	data, err := json.Marshal(result)
	if err != nil {
		return nil
	}
	return data
}
