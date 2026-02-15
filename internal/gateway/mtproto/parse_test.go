package mtproto

import (
	"encoding/json"
	"os"
	"sort"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/bpva/ad-marketplace/internal/entity"
)

func loadTestGraphs(t *testing.T) map[string]json.RawMessage {
	t.Helper()
	data, err := os.ReadFile("test_data.json")
	require.NoError(t, err)

	var graphs map[string]json.RawMessage
	require.NoError(t, json.Unmarshal(data, &graphs))
	return graphs
}

func TestParseLanguagesChart(t *testing.T) {
	graphs := loadTestGraphs(t)
	langs := parseLanguagesChart(graphs["languages"])

	require.Len(t, langs, 2)

	byLang := make(map[string]float64)
	for _, l := range langs {
		byLang[l.Language] = l.Percentage
	}

	assert.Equal(t, float64(4), byLang["English"])
	assert.Equal(t, float64(3), byLang["Russian"])
}

func TestParseTopHoursChart(t *testing.T) {
	graphs := loadTestGraphs(t)
	hours := parseTopHoursChart(graphs["top_hours"])

	require.Len(t, hours, 24, "should have one value per hour")

	assert.Equal(t, float64(15), hours[20], "peak at hour 20 (8 PM UTC)")
	assert.Equal(t, float64(2), hours[22], "secondary at hour 22")

	var totalActivity float64
	for _, v := range hours {
		totalActivity += v
	}
	assert.Equal(t, float64(17), totalActivity)
}

func TestParseSeriesTotals_Nil(t *testing.T) {
	graphs := loadTestGraphs(t)

	assert.Nil(t, parseSeriesTotals(graphs["reactions_by_emotion"]))
	assert.Nil(t, parseSeriesTotals(graphs["story_reactions_by_emotion"]))
	assert.Nil(t, parseSeriesTotals(graphs["iv_interactions"]))
}

func TestParseSeriesTotals_Interactions(t *testing.T) {
	graphs := loadTestGraphs(t)
	totals := parseSeriesTotals(graphs["interactions"])

	require.NotNil(t, totals)
	assert.Equal(t, int64(126), totals["Views"], "sum of all daily view interactions")
	assert.Equal(t, int64(3), totals["Shares"], "sum of all daily shares")
}

func TestDailyStats(t *testing.T) {
	graphs := loadTestGraphs(t)
	g := &gateway{}

	daily := make(map[time.Time]*entity.ChannelHistoricalDayData)

	g.mergeSingleSeries(
		daily,
		graphs["growth"],
		func(d *entity.ChannelHistoricalDayData, v float64) {
			value := roundToInt64(v)
			d.Subscribers = &value
		},
	)
	g.mergeSingleSeries(
		daily,
		graphs["followers"],
		func(d *entity.ChannelHistoricalDayData, v float64) {
			value := roundToInt64(v)
			d.NewFollowers = &value
		},
	)
	g.mergeSingleSeries(daily, graphs["mute"], func(d *entity.ChannelHistoricalDayData, v float64) {
		value := v
		d.MutePct = &value
	})
	g.mergeSingleSeries(
		daily,
		graphs["interactions"],
		func(d *entity.ChannelHistoricalDayData, v float64) {
			value := roundToInt64(v)
			d.Interactions = &value
		},
	)
	g.mergeMultiSeries(
		daily,
		graphs["views_by_source"],
		func(d *entity.ChannelHistoricalDayData, key string, v float64) {
			if d.ViewsBySource == nil {
				d.ViewsBySource = make(map[string]int64)
			}
			d.ViewsBySource[key] = roundToInt64(v)
		},
	)
	g.mergeMultiSeries(
		daily,
		graphs["followers_by_source"],
		func(d *entity.ChannelHistoricalDayData, key string, v float64) {
			if d.FollowersBySource == nil {
				d.FollowersBySource = make(map[string]int64)
			}
			d.FollowersBySource[key] = roundToInt64(v)
		},
	)
	g.mergeMultiSeries(
		daily,
		graphs["story_interactions"],
		func(d *entity.ChannelHistoricalDayData, key string, v float64) {
			if d.StoryInteractions == nil {
				d.StoryInteractions = make(map[string]int64)
			}
			d.StoryInteractions[key] = roundToInt64(v)
		},
	)

	result := dailyToSlice(daily)

	t.Run("sorted_by_date", func(t *testing.T) {
		require.True(t, len(result) > 0)
		assert.True(t, sort.SliceIsSorted(result, func(i, j int) bool {
			return result[i].Date.Before(result[j].Date)
		}))
	})

	t.Run("latest_subscribers_599", func(t *testing.T) {
		last := result[len(result)-1]
		require.NotNil(t, last.Data.Subscribers)
		assert.Equal(t, int64(599), *last.Data.Subscribers)
	})

	t.Run("second_to_last_subscribers_603", func(t *testing.T) {
		prev := result[len(result)-2]
		require.NotNil(t, prev.Data.Subscribers)
		assert.Equal(t, int64(603), *prev.Data.Subscribers)
	})

	t.Run("first_day_subscribers_119", func(t *testing.T) {
		first := result[0]
		require.NotNil(t, first.Data.Subscribers)
		assert.Equal(t, int64(119), *first.Data.Subscribers)
	})

	t.Run("latest_day_has_followers_by_source", func(t *testing.T) {
		feb14 := tsToDate(1771027200000)
		var day *entity.ChannelHistoricalDayData
		for i := range result {
			if result[i].Date.Equal(feb14) {
				day = &result[i].Data
				break
			}
		}
		require.NotNil(t, day)
		assert.Equal(t, int64(364), day.FollowersBySource["URL"])
		assert.Equal(t, int64(127), day.FollowersBySource["Search"])
	})

	t.Run("latest_day_joined_492", func(t *testing.T) {
		feb14 := tsToDate(1771027200000)
		var day *entity.ChannelHistoricalDayData
		for i := range result {
			if result[i].Date.Equal(feb14) {
				day = &result[i].Data
				break
			}
		}
		require.NotNil(t, day)
		require.NotNil(t, day.NewFollowers)
		assert.Equal(t, int64(492), *day.NewFollowers)
	})

	t.Run("latest_day_mute_pct", func(t *testing.T) {
		feb14 := tsToDate(1771027200000)
		var day *entity.ChannelHistoricalDayData
		for i := range result {
			if result[i].Date.Equal(feb14) {
				day = &result[i].Data
				break
			}
		}
		require.NotNil(t, day)
		require.NotNil(t, day.MutePct)
		assert.Equal(t, float64(2), *day.MutePct)
	})

	t.Run("latest_day_views_by_source", func(t *testing.T) {
		feb14 := tsToDate(1771027200000)
		var day *entity.ChannelHistoricalDayData
		for i := range result {
			if result[i].Date.Equal(feb14) {
				day = &result[i].Data
				break
			}
		}
		require.NotNil(t, day)
		assert.Equal(t, int64(52), day.ViewsBySource["Followers"])
		assert.Equal(t, int64(32), day.ViewsBySource["Groups"])
	})

	t.Run("story_interactions_feb14", func(t *testing.T) {
		feb14 := tsToDate(1771027200000)
		var day *entity.ChannelHistoricalDayData
		for i := range result {
			if result[i].Date.Equal(feb14) {
				day = &result[i].Data
				break
			}
		}
		require.NotNil(t, day)
		assert.Equal(t, int64(1), day.StoryInteractions["Views"])
		assert.Equal(t, int64(1), day.StoryInteractions["Shares"])
	})

	t.Run("interactions_feb14", func(t *testing.T) {
		feb14 := tsToDate(1771027200000)
		var day *entity.ChannelHistoricalDayData
		for i := range result {
			if result[i].Date.Equal(feb14) {
				day = &result[i].Data
				break
			}
		}
		require.NotNil(t, day)
		require.NotNil(t, day.Interactions)
		assert.Equal(t, int64(84), *day.Interactions)
	})
}
