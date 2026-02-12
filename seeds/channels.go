package seeds

import (
	"context"
	"fmt"

	"github.com/bpva/ad-marketplace/internal/entity"
	channel_repo "github.com/bpva/ad-marketplace/internal/repository/channel"
)

type seedChannel struct {
	entity      *entity.Channel
	subscribers int
}

func (s *Seeder) seedChannels(ctx context.Context, users []seedUser) ([]seedChannel, error) {
	channels := channel_repo.New(s.db)

	type fmtDef struct {
		typ       entity.AdFormatType
		native    bool
		feed, top int
		price     int64
	}
	type channelDef struct {
		tgID        int64
		title       string
		username    string
		ownerIdx    int
		managerIdx  int
		listed      bool
		subscribers int
		formats     []fmtDef
	}

	defs := []channelDef{
		{
			tgID: -1001234567890, title: "Tech Daily", username: "techdaily",
			ownerIdx: 0, managerIdx: -1, listed: true, subscribers: 45200,
			formats: []fmtDef{
				{entity.AdFormatTypePost, false, 24, 2, 5_000_000_000},
				{entity.AdFormatTypePost, true, 24, 4, 8_000_000_000},
			},
		},
		{
			tgID: -1009876543210, title: "Crypto Signals", username: "cryptosignals",
			ownerIdx: 1, managerIdx: 0, listed: true, subscribers: 128000,
			formats: []fmtDef{
				{entity.AdFormatTypePost, false, 24, 2, 15_000_000_000},
				{entity.AdFormatTypeRepost, false, 12, 2, 7_000_000_000},
				{entity.AdFormatTypeStory, false, 24, 2, 3_000_000_000},
			},
		},
		{
			tgID: -1001111111111, title: "Dev Memes", username: "devmemes",
			ownerIdx: 0, managerIdx: -1, listed: true, subscribers: 8900,
			formats: []fmtDef{
				{entity.AdFormatTypePost, true, 12, 2, 2_000_000_000},
			},
		},
		{
			tgID: -1002222222222, title: "AI Research Hub", username: "airesearchhub",
			ownerIdx: 1, managerIdx: -1, listed: true, subscribers: 67500,
			formats: []fmtDef{
				{entity.AdFormatTypePost, false, 24, 4, 12_000_000_000},
				{entity.AdFormatTypePost, true, 24, 2, 18_000_000_000},
				{entity.AdFormatTypeStory, false, 24, 2, 5_000_000_000},
			},
		},
		{
			tgID: -1003333333333, title: "TON Ecosystem", username: "tonecosystem",
			ownerIdx: 0, managerIdx: 1, listed: false, subscribers: 23100,
		},
	}

	result := make([]seedChannel, 0, len(defs))
	for _, d := range defs {
		username := d.username
		ch, err := channels.Create(ctx, d.tgID, d.title, &username)
		if err != nil {
			return nil, fmt.Errorf("create channel %s: %w", d.title, err)
		}

		if _, err := channels.CreateRole(
			ctx, ch.ID, users[d.ownerIdx].entity.ID, entity.ChannelRoleTypeOwner,
		); err != nil {
			return nil, fmt.Errorf("create owner role for %s: %w", d.title, err)
		}
		if d.managerIdx >= 0 {
			if _, err := channels.CreateRole(
				ctx, ch.ID, users[d.managerIdx].entity.ID, entity.ChannelRoleTypeManager,
			); err != nil {
				return nil, fmt.Errorf("create manager role for %s: %w", d.title, err)
			}
		}

		if d.listed {
			if err := channels.UpdateListing(ctx, ch.ID, true); err != nil {
				return nil, fmt.Errorf("list channel %s: %w", d.title, err)
			}
		}

		for _, f := range d.formats {
			if _, err := channels.CreateAdFormat(
				ctx, ch.ID, f.typ, f.native, f.feed, f.top, f.price,
			); err != nil {
				return nil, fmt.Errorf("create ad format for %s: %w", d.title, err)
			}
		}

		result = append(result, seedChannel{entity: ch, subscribers: d.subscribers})
	}
	return result, nil
}
