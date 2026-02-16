package seeds

import (
	"context"
	"fmt"

	"github.com/bpva/ad-marketplace/internal/entity"
	post_repo "github.com/bpva/ad-marketplace/internal/repository/post"
	petname "github.com/dustinkirkland/golang-petname"
)

func (s *Seeder) seedPosts(ctx context.Context, users []seedUser) error {
	posts := post_repo.New(s.db)

	defs := []struct {
		userIdx  int
		text     string
		entities []byte
	}{
		{
			0,
			"NEW: AI Code Assistant v2.0 is here!\n\nThe fastest AI pair-programmer just got smarter. Now supports 40+ languages with real-time suggestions.\n\nWhat's new:\n- Context-aware completions\n- Built-in code review\n- One-click deployment\n\nTry it free for 30 days: Get Started\n\nBuilt by @techdaily team",
			[]byte(
				`[{"type":"bold","offset":0,"length":36},{"type":"italic","offset":38,"length":48},{"type":"bold","offset":143,"length":11},{"type":"text_link","offset":255,"length":11,"url":"https://example.com/try"},{"type":"mention","offset":277,"length":10}]`,
			),
		},
		{
			1,
			"MARKET UPDATE: TON breaks $8 resistance\n\nKey levels to watch:\n\n> Support: $7.40\n> Resistance: $8.50\n> Volume: 2.4x above average\n\nOur analysis suggests continued bullish momentum through Q1. The network activity metrics confirm growing adoption.\n\nFull report: https://example.com/ton-analysis\n\nNot financial advice. DYOR.",
			[]byte(
				`[{"type":"bold","offset":0,"length":39},{"type":"underline","offset":41,"length":20},{"type":"blockquote","offset":63,"length":65},{"type":"url","offset":260,"length":32},{"type":"italic","offset":294,"length":27}]`,
			),
		},
		{
			0,
			"Ship faster with our CLI toolkit\n\nInstall in seconds:\n\nnpm install -g @devtools/cli\n\nThen run:\n\ndevtools init --template react\ndevtools deploy --prod\n\nZero config. Works with Next.js, Remix, and Astro out of the box.\n\nDocs and examples: devtools.dev\n\n1200+ teams already use it daily",
			[]byte(
				`[{"type":"bold","offset":0,"length":32},{"type":"bold","offset":34,"length":19},{"type":"code","offset":55,"length":28},{"type":"pre","offset":96,"length":53,"language":"bash"},{"type":"text_link","offset":237,"length":12,"url":"https://devtools.dev"},{"type":"bold","offset":251,"length":32}]`,
			),
		},
		{
			2,
			"Join us at Web3 Dev Summit 2026!\n\nMarch 15-17 | Dubai\n\nEarly bird: $199 $149\n\nSpeakers include founders from @tonecosystem and leading DeFi protocols.\n\nTopics:\n- Smart contract security\n- Cross-chain bridges\n- Scalability solutions\n\nLimited seats available. Register now\n\nUse code TELEGRAM for 15% off",
			[]byte(
				`[{"type":"bold","offset":0,"length":32},{"type":"italic","offset":34,"length":19},{"type":"strikethrough","offset":67,"length":4},{"type":"bold","offset":72,"length":4},{"type":"mention","offset":109,"length":13},{"type":"bold","offset":152,"length":7},{"type":"text_link","offset":258,"length":12,"url":"https://example.com/register"},{"type":"code","offset":281,"length":8}]`,
			),
		},
		{
			0,
			"Premium Telegram analytics for channel owners.\n\nTrack growth, engagement, and revenue in real time.\n\nStart free trial",
			[]byte(
				`[{"type":"bold","offset":0,"length":46},{"type":"text_link","offset":101,"length":16,"url":"https://example.com/analytics"}]`,
			),
		},
		{
			2,
			"We tested 5 VPN services for Telegram users\n\nAfter a month of testing speed, privacy policies, and reliability, here is our verdict:\n\n> Best overall: FastVPN\n> Best budget: ShieldNet\n> Avoid: two services had serious logging issues\n\nThe winner surprised us. FastVPN delivered 94% speed retention with a strict no-logs policy verified by independent audit.\n\nRead the full comparison\n\nSponsored by FastVPN",
			[]byte(
				`[{"type":"bold","offset":0,"length":44},{"type":"italic","offset":46,"length":87},{"type":"blockquote","offset":135,"length":97},{"type":"spoiler","offset":193,"length":39},{"type":"bold","offset":259,"length":7},{"type":"text_link","offset":358,"length":24,"url":"https://example.com/vpn-review"},{"type":"italic","offset":384,"length":20}]`,
			),
		},
	}

	for _, d := range defs {
		text := d.text
		name := petname.Generate(2, " ")
		_, err := posts.Create(
			ctx,
			entity.PostTypeTemplate,
			users[d.userIdx].entity.ID,
			nil,
			&name,
			nil,
			&text,
			d.entities,
			nil,
			nil,
			false,
			false,
		)
		if err != nil {
			return fmt.Errorf("create post: %w", err)
		}
	}
	return nil
}
