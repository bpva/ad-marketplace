package mtproto

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/gotd/td/tg"

	"github.com/bpva/ad-marketplace/internal/entity"
)

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

	return c.parseBroadcastStats(ctx, api, stats), nil
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
