package dto

import (
	"encoding/json"
	"time"

	"github.com/bpva/ad-marketplace/internal/entity"
)

type PostMediaItem struct {
	PostID                string           `json:"post_id"`
	MediaType             entity.MediaType `json:"media_type"`
	HasMediaSpoiler       bool             `json:"has_media_spoiler"`
	ShowCaptionAboveMedia bool             `json:"show_caption_above_media"`
}

type TemplateResponse struct {
	ID        string          `json:"id"`
	Text      *string         `json:"text,omitempty"`
	Entities  json.RawMessage `json:"entities,omitempty"`
	Media     []PostMediaItem `json:"media,omitempty"`
	CreatedAt time.Time       `json:"created_at"`
}

type TemplatesResponse struct {
	Templates []TemplateResponse `json:"templates"`
}
