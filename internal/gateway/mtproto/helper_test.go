package mtproto

import (
	"encoding/json"
	"fmt"
	"math"
	"math/rand/v2"
	"time"

	"github.com/gotd/td/tg"
)

// channelProfile holds parameters that define the channel's "personality".
type channelProfile struct {
	followers        float64 // current subscriber count
	followersGrowth  float64 // daily net growth (can be negative)
	avgViewsPerPost  float64
	avgSharesPerPost float64
	avgReactsPerPost float64
	avgViewsPerStory float64
	avgSharesStory   float64
	avgReactsStory   float64
	notifEnabledPct  float64 // 0..1
	postsPerDay      float64
}

func defaultProfile() channelProfile {
	return channelProfile{
		followers:        52_340,
		followersGrowth:  85,
		avgViewsPerPost:  12_450,
		avgSharesPerPost: 340,
		avgReactsPerPost: 215,
		avgViewsPerStory: 8_200,
		avgSharesStory:   120,
		avgReactsStory:   95,
		notifEnabledPct:  0.38,
		postsPerDay:      2.5,
	}
}

func generateBroadcastStats() *tg.StatsBroadcastStats {
	p := defaultProfile()
	now := time.Now().UTC()
	periodStart := now.AddDate(0, 0, -7)
	prevStart := periodStart.AddDate(0, 0, -7)

	interactionViews, interactionShares := interactionsSeries(p)
	storyViews, storyShares := storyInteractionsSeries(p)

	return &tg.StatsBroadcastStats{
		Period: tg.StatsDateRangeDays{
			MinDate: int(periodStart.Unix()),
			MaxDate: int(now.Unix()),
		},
		Followers:         absVal(p.followers, p.followers-p.followersGrowth*7*jitter(0.15)),
		ViewsPerPost:      absVal(p.avgViewsPerPost, p.avgViewsPerPost*jitter(0.12)),
		SharesPerPost:     absVal(p.avgSharesPerPost, p.avgSharesPerPost*jitter(0.18)),
		ReactionsPerPost:  absVal(p.avgReactsPerPost, p.avgReactsPerPost*jitter(0.15)),
		ViewsPerStory:     absVal(p.avgViewsPerStory, p.avgViewsPerStory*jitter(0.14)),
		SharesPerStory:    absVal(p.avgSharesStory, p.avgSharesStory*jitter(0.20)),
		ReactionsPerStory: absVal(p.avgReactsStory, p.avgReactsStory*jitter(0.18)),
		EnabledNotifications: tg.StatsPercentValue{
			Part:  math.Round(p.followers*p.notifEnabledPct*100) / 100,
			Total: p.followers,
		},

		GrowthGraph:                  makeLineGraph("Growth", periodStart, prevStart, now, growthSeries(p)),
		FollowersGraph:               makeLineGraph("Followers", periodStart, prevStart, now, followersDeltaSeries(p)),
		MuteGraph:                    makeLineGraph("Notifications", periodStart, prevStart, now, muteSeries(p)),
		TopHoursGraph:                makeBarGraph("Views by hour", topHoursSeries()),
		InteractionsGraph:            makeDoubleLineGraph("Interactions", periodStart, now, interactionViews, interactionShares),
		IvInteractionsGraph:          makeLineGraph("IV Interactions", periodStart, prevStart, now, ivInteractionsSeries(p)),
		ViewsBySourceGraph:           makeStackedGraph("Views by source", periodStart, now, viewsBySourceSeries(p)),
		NewFollowersBySourceGraph:    makeStackedGraph("New followers by source", periodStart, now, newFollowersBySourceSeries(p)),
		LanguagesGraph:               makePieGraph("Languages", languageSeries()),
		ReactionsByEmotionGraph:      makeStackedGraph("Reactions by emotion", periodStart, now, reactionsByEmotionSeries()),
		StoryInteractionsGraph:       makeDoubleLineGraph("Story interactions", periodStart, now, storyViews, storyShares),
		StoryReactionsByEmotionGraph: makeStackedGraph("Story reactions by emotion", periodStart, now, storyReactionsByEmotionSeries()),

		RecentPostsInteractions: recentPosts(p, now),
	}
}

// --- Value helpers ---

func absVal(current, previous float64) tg.StatsAbsValueAndPrev {
	return tg.StatsAbsValueAndPrev{
		Current:  math.Round(current*100) / 100,
		Previous: math.Round(previous*100) / 100,
	}
}

// jitter returns a multiplier in [1-pct, 1+pct].
func jitter(pct float64) float64 {
	return 1 + (rand.Float64()*2-1)*pct
}

func jitterInt(base, pct float64) int {
	return int(math.Round(base * jitter(pct)))
}

// --- Graph JSON builders ---
// Telegram graphs are StatsGraph with a JSON field containing chart.js-like data.
// The JSON follows the format: {"columns": [...], "types": {...}, "names": {...}, "colors": {...}}

type graphJSON struct {
	Columns []any             `json:"columns"`
	Types   map[string]string `json:"types"`
	Names   map[string]string `json:"names"`
	Colors  map[string]string `json:"colors"`
}

func toStatsGraph(g graphJSON) *tg.StatsGraph {
	raw, _ := json.Marshal(g)
	return &tg.StatsGraph{
		JSON: tg.DataJSON{Data: string(raw)},
	}
}

func timestamps(start, end time.Time, step time.Duration) []any {
	var ts []any
	ts = append(ts, "x")
	for t := start; !t.After(end); t = t.Add(step) {
		ts = append(ts, t.UnixMilli())
	}
	return ts
}

func hours24() []any {
	h := []any{"x"}
	for i := 0; i < 24; i++ {
		h = append(h, i)
	}
	return h
}

// --- Series generators ---

func growthSeries(p channelProfile) []any {
	// Absolute subscriber count over 7 days.
	col := []any{"y0"}
	base := p.followers - p.followersGrowth*7
	for i := 0; i <= 7*24; i++ {
		hourGrowth := p.followersGrowth / 24
		base += hourGrowth * jitter(0.4)
		col = append(col, int(math.Round(base)))
	}
	return col
}

func followersDeltaSeries(p channelProfile) []any {
	// Net new followers per day.
	col := []any{"y0"}
	for i := 0; i <= 7*24; i++ {
		daily := p.followersGrowth / 24
		val := daily*jitter(0.6) - daily*0.15*jitter(0.5) // joins minus leaves
		col = append(col, int(math.Round(val)))
	}
	return col
}

func muteSeries(p channelProfile) []any {
	// Percentage of muted users over time (hovers around 1-notifPct).
	col := []any{"y0"}
	mutePct := 1 - p.notifEnabledPct
	for i := 0; i <= 7*24; i++ {
		col = append(col, math.Round((mutePct+rand.Float64()*0.02-0.01)*10000)/100)
	}
	return col
}

func topHoursSeries() []any {
	// Views distribution by hour of day.
	col := []any{"y0"}
	// Peak at 10-14 and 19-22 (UTC), low at 2-6.
	hourWeights := []float64{
		0.15, 0.08, 0.05, 0.04, 0.03, 0.04, 0.08, 0.18,
		0.35, 0.52, 0.68, 0.72, 0.70, 0.65, 0.55, 0.48,
		0.50, 0.58, 0.65, 0.78, 0.82, 0.75, 0.55, 0.30,
	}
	for _, w := range hourWeights {
		col = append(col, math.Round(w*jitter(0.10)*10000)/100)
	}
	return col
}

func interactionsSeries(p channelProfile) (views, shares []any) {
	views = []any{"y0"}
	shares = []any{"y1"}
	for i := 0; i <= 7; i++ {
		views = append(views, jitterInt(p.avgViewsPerPost*p.postsPerDay, 0.25))
		shares = append(shares, jitterInt(p.avgSharesPerPost*p.postsPerDay, 0.30))
	}
	return
}

func ivInteractionsSeries(p channelProfile) []any {
	// Instant View interactions — usually much lower than regular views.
	col := []any{"y0"}
	ivBase := p.avgViewsPerPost * 0.05
	for i := 0; i <= 7*24; i++ {
		col = append(col, jitterInt(ivBase/24, 0.5))
	}
	return col
}

func viewsBySourceSeries(p channelProfile) map[string][]any {
	sources := map[string]float64{
		"Follows":  0.45,
		"Others":   0.20,
		"Channels": 0.18,
		"Search":   0.10,
		"Forwards": 0.07,
	}
	result := make(map[string][]any)
	totalDaily := p.avgViewsPerPost * p.postsPerDay
	for name, frac := range sources {
		col := []any{name}
		for i := 0; i <= 7; i++ {
			col = append(col, jitterInt(totalDaily*frac, 0.25))
		}
		result[name] = col
	}
	return result
}

func newFollowersBySourceSeries(p channelProfile) map[string][]any {
	sources := map[string]float64{
		"Search":   0.30,
		"Channels": 0.25,
		"Contacts": 0.15,
		"Mentions": 0.15,
		"Other":    0.15,
	}
	result := make(map[string][]any)
	for name, frac := range sources {
		col := []any{name}
		for i := 0; i <= 7; i++ {
			col = append(col, jitterInt(p.followersGrowth*frac, 0.35))
		}
		result[name] = col
	}
	return result
}

func languageSeries() map[string]float64 {
	return map[string]float64{
		"English": 42.5,
		"Russian": 18.3,
		"Spanish": 8.7,
		"Arabic":  6.2,
		"German":  5.1,
		"French":  4.8,
		"Other":   14.4,
	}
}

func reactionsByEmotionSeries() map[string][]any {
	emotions := map[string]float64{
		"\U0001f44d":   0.40,
		"\u2764\ufe0f": 0.25,
		"\U0001f525":   0.15,
		"\U0001f389":   0.10,
		"\U0001f914":   0.10,
	}
	result := make(map[string][]any)
	for emoji, frac := range emotions {
		col := []any{emoji}
		for i := 0; i <= 7; i++ {
			col = append(col, jitterInt(200*frac, 0.30))
		}
		result[emoji] = col
	}
	return result
}

func storyInteractionsSeries(p channelProfile) (views, shares []any) {
	views = []any{"y0"}
	shares = []any{"y1"}
	for i := 0; i <= 7; i++ {
		views = append(views, jitterInt(p.avgViewsPerStory, 0.25))
		shares = append(shares, jitterInt(p.avgSharesStory, 0.30))
	}
	return
}

func storyReactionsByEmotionSeries() map[string][]any {
	emotions := map[string]float64{
		"\U0001f44d":   0.35,
		"\u2764\ufe0f": 0.30,
		"\U0001f525":   0.20,
		"\U0001f602":   0.15,
	}
	result := make(map[string][]any)
	for emoji, frac := range emotions {
		col := []any{emoji}
		for i := 0; i <= 7; i++ {
			col = append(col, jitterInt(90*frac, 0.35))
		}
		result[emoji] = col
	}
	return result
}

// --- Graph constructors ---

func makeLineGraph(name string, start, _, end time.Time, ySeries []any) *tg.StatsGraph {
	g := graphJSON{
		Columns: []any{
			timestamps(start, end, time.Hour),
			ySeries,
		},
		Types:  map[string]string{"y0": "line", "x": "x"},
		Names:  map[string]string{"y0": name},
		Colors: map[string]string{"y0": "#3DC23F"},
	}
	return toStatsGraph(g)
}

func makeDoubleLineGraph(name string, start, end time.Time, y0, y1 []any) *tg.StatsGraph {
	g := graphJSON{
		Columns: []any{
			timestamps(start, end, 24*time.Hour),
			y0, y1,
		},
		Types:  map[string]string{"y0": "line", "y1": "line", "x": "x"},
		Names:  map[string]string{"y0": "Views", "y1": "Shares"},
		Colors: map[string]string{"y0": "#3DC23F", "y1": "#F34C44"},
	}
	return toStatsGraph(g)
}

func makeBarGraph(name string, ySeries []any) *tg.StatsGraph {
	g := graphJSON{
		Columns: []any{
			hours24(),
			ySeries,
		},
		Types:  map[string]string{"y0": "bar", "x": "x"},
		Names:  map[string]string{"y0": name},
		Colors: map[string]string{"y0": "#4681BB"},
	}
	return toStatsGraph(g)
}

func makeStackedGraph(name string, start, end time.Time, series map[string][]any) *tg.StatsGraph {
	ts := timestamps(start, end, 24*time.Hour)
	columns := []any{ts}
	types := map[string]string{"x": "x"}
	names := map[string]string{}
	colors := map[string]string{}
	palette := []string{"#3DC23F", "#F34C44", "#4681BB", "#FFA841", "#AB7FD4", "#E667AF"}
	i := 0
	for k, col := range series {
		id := fmt.Sprintf("y%d", i)
		// Replace the human-readable first element with the series id.
		col[0] = id
		columns = append(columns, col)
		types[id] = "area"
		names[id] = k
		colors[id] = palette[i%len(palette)]
		i++
	}
	g := graphJSON{
		Columns: columns,
		Types:   types,
		Names:   names,
		Colors:  colors,
	}
	return toStatsGraph(g)
}

func makePieGraph(_ string, slices map[string]float64) *tg.StatsGraph {
	columns := []any{}
	types := map[string]string{}
	names := map[string]string{}
	colors := map[string]string{}
	palette := []string{"#4681BB", "#F34C44", "#FFA841", "#3DC23F", "#AB7FD4", "#E667AF", "#999999"}
	i := 0
	for lang, pct := range slices {
		id := fmt.Sprintf("y%d", i)
		columns = append(columns, []any{id, pct})
		types[id] = "pie"
		names[id] = lang
		colors[id] = palette[i%len(palette)]
		i++
	}
	g := graphJSON{
		Columns: columns,
		Types:   types,
		Names:   names,
		Colors:  colors,
	}
	return toStatsGraph(g)
}

// --- Recent posts ---

func recentPosts(p channelProfile, now time.Time) []tg.PostInteractionCountersClass {
	// Generate ~15-20 recent messages and a couple of stories.
	count := 15 + rand.IntN(6)
	result := make([]tg.PostInteractionCountersClass, 0, count+2)

	baseMsgID := 4200 + rand.IntN(500)
	for i := 0; i < count; i++ {
		msgID := baseMsgID + i*rand.IntN(3) + 1
		// Newer posts have fewer views (less time to accumulate).
		ageFactor := 1.0 - float64(i)*0.03
		if ageFactor < 0.4 {
			ageFactor = 0.4
		}
		result = append(result, &tg.PostInteractionCountersMessage{
			MsgID:     msgID,
			Views:     jitterInt(p.avgViewsPerPost*ageFactor, 0.20),
			Forwards:  jitterInt(p.avgSharesPerPost*ageFactor, 0.25),
			Reactions: jitterInt(p.avgReactsPerPost*ageFactor, 0.30),
		})
		baseMsgID = msgID
	}

	// A couple of recent stories.
	for i := 0; i < 2; i++ {
		result = append(result, &tg.PostInteractionCountersStory{
			StoryID:   100 + i + rand.IntN(20),
			Views:     jitterInt(p.avgViewsPerStory, 0.20),
			Forwards:  jitterInt(p.avgSharesStory, 0.25),
			Reactions: jitterInt(p.avgReactsStory, 0.30),
		})
	}

	return result
}
