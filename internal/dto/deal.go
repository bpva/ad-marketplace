package dto

import (
	"encoding/json"
	"time"

	"github.com/bpva/ad-marketplace/internal/entity"
)

type CreateDealRequest struct {
	TgChannelID    int64               `json:"channel_id" validate:"required"`
	FormatType     entity.AdFormatType `json:"format_type" validate:"required"`
	IsNative       bool                `json:"is_native"`
	FeedHours      int                 `json:"feed_hours" validate:"required,gt=0"`
	TopHours       int                 `json:"top_hours" validate:"required,gt=0"`
	PriceNanoTON   int64               `json:"price_nano_ton" validate:"required,gt=0"`
	TemplatePostID string              `json:"template_post_id" validate:"required,uuid"`
	ScheduledAt    time.Time           `json:"scheduled_at" validate:"required"`
}

type RejectRequest struct {
	Reason *string `json:"reason,omitempty"`
}

type RequestChangesRequest struct {
	Note string `json:"note" validate:"required"`
}

type DealResponse struct {
	ID            string              `json:"id"`
	TgChannelID   int64               `json:"channel_id"`
	Status        entity.DealStatus   `json:"status"`
	ScheduledAt   time.Time           `json:"scheduled_at"`
	PublisherNote *string             `json:"publisher_note,omitempty"`
	FormatType    entity.AdFormatType `json:"format_type"`
	IsNative      bool                `json:"is_native"`
	FeedHours     int                 `json:"feed_hours"`
	TopHours      int                 `json:"top_hours"`
	PriceNanoTON  int64               `json:"price_nano_ton"`
	Ad            *TemplateResponse   `json:"ad,omitempty"`
	CreatedAt     time.Time           `json:"created_at"`
}

type DealsResponse struct {
	Deals []DealResponse `json:"deals"`
	Total int            `json:"total"`
}

type DealListItem struct {
	entity.Deal
	TgChannelID int64
}

func DealResponseFrom(deal *entity.Deal, posts []entity.Post, tgChannelID int64) DealResponse {
	resp := DealResponse{
		ID:            deal.ID.String(),
		TgChannelID:   tgChannelID,
		Status:        deal.Status,
		ScheduledAt:   deal.ScheduledAt,
		PublisherNote: deal.PublisherNote,
		FormatType:    deal.FormatType,
		IsNative:      deal.IsNative,
		FeedHours:     deal.FeedHours,
		TopHours:      deal.TopHours,
		PriceNanoTON:  deal.PriceNanoTON,
		CreatedAt:     deal.CreatedAt,
	}

	if len(posts) > 0 {
		ad := buildAdResponse(posts)
		resp.Ad = &ad
	}

	return resp
}

func DealListResponseFrom(item DealListItem) DealResponse {
	return DealResponse{
		ID:            item.ID.String(),
		TgChannelID:   item.TgChannelID,
		Status:        item.Status,
		ScheduledAt:   item.ScheduledAt,
		PublisherNote: item.PublisherNote,
		FormatType:    item.FormatType,
		IsNative:      item.IsNative,
		FeedHours:     item.FeedHours,
		TopHours:      item.TopHours,
		PriceNanoTON:  item.PriceNanoTON,
		CreatedAt:     item.CreatedAt,
	}
}

func buildAdResponse(posts []entity.Post) TemplateResponse {
	resp := TemplateResponse{
		ID:        posts[0].ID.String(),
		CreatedAt: posts[0].CreatedAt,
	}

	for i := range posts {
		p := &posts[i]
		if p.Text != nil && resp.Text == nil {
			resp.Text = p.Text
			if len(p.Entities) > 0 && string(p.Entities) != "null" {
				resp.Entities = json.RawMessage(p.Entities)
			}
		}
		if p.MediaType != nil {
			resp.Media = append(resp.Media, PostMediaItem{
				PostID:                p.ID.String(),
				MediaType:             *p.MediaType,
				HasMediaSpoiler:       p.HasMediaSpoiler,
				ShowCaptionAboveMedia: p.ShowCaptionAboveMedia,
			})
		}
	}

	return resp
}
