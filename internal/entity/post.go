package entity

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

type MediaType string

const (
	MediaTypePhoto     MediaType = "photo"
	MediaTypeVideo     MediaType = "video"
	MediaTypeDocument  MediaType = "document"
	MediaTypeAnimation MediaType = "animation"
	MediaTypeAudio     MediaType = "audio"
	MediaTypeVoice     MediaType = "voice"
	MediaTypeVideoNote MediaType = "video_note"
	MediaTypeSticker   MediaType = "sticker"
)

func (m *MediaType) Scan(src any) error {
	switch v := src.(type) {
	case string:
		*m = MediaType(v)
	case []byte:
		return m.Scan(string(v))
	case nil:
		return nil
	default:
		return fmt.Errorf("cannot scan %T into MediaType", src)
	}
	return nil
}

type Post struct {
	ID                    uuid.UUID  `db:"id"`
	UserID                uuid.UUID  `db:"user_id"`
	MediaGroupID          *string    `db:"media_group_id"`
	Text                  *string    `db:"text"`
	Entities              []byte     `db:"entities"`
	MediaType             *MediaType `db:"media_type"`
	MediaFileID           *string    `db:"media_file_id"`
	HasMediaSpoiler       bool       `db:"has_media_spoiler"`
	ShowCaptionAboveMedia bool       `db:"show_caption_above_media"`
	CreatedAt             time.Time  `db:"created_at"`
	DeletedAt             *time.Time `db:"deleted_at"`
}
