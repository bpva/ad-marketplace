package post

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
	tele "gopkg.in/telebot.v4"

	"github.com/bpva/ad-marketplace/internal/dto"
	"github.com/bpva/ad-marketplace/internal/entity"
	"github.com/bpva/ad-marketplace/internal/logx"
)

type PostRepository interface {
	GetTemplatesByOwner(ctx context.Context, ownerID uuid.UUID) ([]entity.Post, error)
	GetByID(ctx context.Context, id uuid.UUID) (*entity.Post, error)
	GetByMediaGroupID(ctx context.Context, mediaGroupID string) ([]entity.Post, error)
}

type TelebotClient interface {
	DownloadFile(fileID string) ([]byte, error)
	Send(to tele.Recipient, what any, opts ...any) (*tele.Message, error)
	SendAlbum(to tele.Recipient, a tele.Album, opts ...any) ([]tele.Message, error)
}

type svc struct {
	postRepo PostRepository
	bot      TelebotClient
	log      *slog.Logger
}

func New(postRepo PostRepository, bot TelebotClient, log *slog.Logger) *svc {
	log = log.With(logx.Service("PostService"))
	return &svc{
		postRepo: postRepo,
		bot:      bot,
		log:      log,
	}
}

func (s *svc) GetUserTemplates(ctx context.Context) (*dto.TemplatesResponse, error) {
	user, ok := dto.UserFromContext(ctx)
	if !ok {
		return nil, fmt.Errorf("get user templates: %w", dto.ErrForbidden)
	}

	posts, err := s.postRepo.GetTemplatesByOwner(ctx, user.ID)
	if err != nil {
		return nil, fmt.Errorf("get posts: %w", err)
	}

	templates := groupPosts(posts)
	return &dto.TemplatesResponse{Templates: templates}, nil
}

func (s *svc) GetPostMedia(ctx context.Context, postID uuid.UUID) ([]byte, error) {
	user, ok := dto.UserFromContext(ctx)
	if !ok {
		return nil, fmt.Errorf("get post media: %w", dto.ErrForbidden)
	}

	post, err := s.postRepo.GetByID(ctx, postID)
	if err != nil {
		return nil, fmt.Errorf("get post: %w", err)
	}

	if post.Type != entity.PostTypeTemplate || post.ExternalID != user.ID {
		return nil, fmt.Errorf("get post media: %w", dto.ErrForbidden)
	}

	if post.MediaFileID == nil {
		return nil, fmt.Errorf("get post media: %w", dto.ErrNotFound)
	}

	data, err := s.bot.DownloadFile(*post.MediaFileID)
	if err != nil {
		return nil, fmt.Errorf("download post media: %w", err)
	}

	return data, nil
}

func (s *svc) SendPreview(ctx context.Context, postID uuid.UUID) error {
	user, ok := dto.UserFromContext(ctx)
	if !ok {
		return fmt.Errorf("send preview: %w", dto.ErrForbidden)
	}

	post, err := s.postRepo.GetByID(ctx, postID)
	if err != nil {
		return fmt.Errorf("get post: %w", err)
	}

	if post.Type != entity.PostTypeTemplate || post.ExternalID != user.ID {
		return fmt.Errorf("send preview: %w", dto.ErrForbidden)
	}

	recipient := tele.ChatID(user.TgID)

	if post.MediaGroupID != nil {
		return s.sendAlbumPreview(ctx, post, recipient)
	}
	if post.MediaType != nil {
		return s.sendMediaPreview(post, recipient)
	}
	return s.sendTextPreview(post, recipient)
}

func (s *svc) sendTextPreview(post *entity.Post, to tele.Recipient) error {
	if post.Text == nil {
		return fmt.Errorf("send text preview: %w", dto.ErrNotFound)
	}

	var opts []any
	if sendOpts := buildSendOptions(post.Entities); sendOpts != nil {
		opts = append(opts, sendOpts)
	}
	_, err := s.bot.Send(to, *post.Text, opts...)
	if err != nil {
		return fmt.Errorf("send text preview: %w", err)
	}
	return nil
}

func (s *svc) sendMediaPreview(post *entity.Post, to tele.Recipient) error {
	media := buildSendable(post)
	if media == nil {
		return fmt.Errorf("send media preview: %w", dto.ErrNotFound)
	}

	var opts []any
	if sendOpts := buildSendOptions(post.Entities); sendOpts != nil {
		opts = append(opts, sendOpts)
	}
	_, err := s.bot.Send(to, media, opts...)
	if err != nil {
		return fmt.Errorf("send media preview: %w", err)
	}
	return nil
}

func (s *svc) sendAlbumPreview(ctx context.Context, post *entity.Post, to tele.Recipient) error {
	posts, err := s.postRepo.GetByMediaGroupID(ctx, *post.MediaGroupID)
	if err != nil {
		return fmt.Errorf("get album posts: %w", err)
	}

	var caption string
	var captionEntities []tele.MessageEntity
	for _, p := range posts {
		if p.Text != nil {
			caption = *p.Text
			captionEntities = parseEntities(p.Entities)
			break
		}
	}

	var album tele.Album
	for i := range posts {
		p := &posts[i]
		if p.MediaType == nil || p.MediaFileID == nil {
			continue
		}
		item := buildAlbumItem(p)
		if item == nil {
			continue
		}
		if i == 0 && caption != "" {
			setAlbumItemCaption(item, caption)
		}
		album = append(album, item)
	}

	if len(album) == 0 {
		return fmt.Errorf("send album preview: %w", dto.ErrNotFound)
	}

	var opts []any
	if len(captionEntities) > 0 {
		opts = append(opts, &tele.SendOptions{Entities: captionEntities})
	}
	_, err = s.bot.SendAlbum(to, album, opts...)
	if err != nil {
		return fmt.Errorf("send album preview: %w", err)
	}
	return nil
}

func buildSendOptions(rawEntities []byte) *tele.SendOptions {
	entities := parseEntities(rawEntities)
	if len(entities) == 0 {
		return nil
	}
	return &tele.SendOptions{Entities: entities}
}

func parseEntities(raw []byte) []tele.MessageEntity {
	if len(raw) == 0 || string(raw) == "null" {
		return nil
	}
	var entities []tele.MessageEntity
	if err := json.Unmarshal(raw, &entities); err != nil {
		return nil
	}
	return entities
}

func buildSendable(p *entity.Post) tele.Sendable {
	if p.MediaType == nil || p.MediaFileID == nil {
		return nil
	}
	file := tele.File{FileID: *p.MediaFileID}
	caption := ""
	if p.Text != nil {
		caption = *p.Text
	}

	switch *p.MediaType {
	case entity.MediaTypePhoto:
		return &tele.Photo{
			File:         file,
			Caption:      caption,
			HasSpoiler:   p.HasMediaSpoiler,
			CaptionAbove: p.ShowCaptionAboveMedia,
		}
	case entity.MediaTypeVideo:
		return &tele.Video{
			File:         file,
			Caption:      caption,
			HasSpoiler:   p.HasMediaSpoiler,
			CaptionAbove: p.ShowCaptionAboveMedia,
		}
	case entity.MediaTypeDocument:
		return &tele.Document{File: file, Caption: caption}
	case entity.MediaTypeAnimation:
		return &tele.Animation{File: file, Caption: caption, HasSpoiler: p.HasMediaSpoiler}
	case entity.MediaTypeAudio:
		return &tele.Audio{File: file, Caption: caption}
	case entity.MediaTypeVoice:
		return &tele.Voice{File: file, Caption: caption}
	case entity.MediaTypeVideoNote:
		return &tele.VideoNote{File: file}
	case entity.MediaTypeSticker:
		return &tele.Sticker{File: file}
	default:
		return nil
	}
}

func buildAlbumItem(p *entity.Post) tele.Inputtable {
	file := tele.File{FileID: *p.MediaFileID}
	switch *p.MediaType {
	case entity.MediaTypePhoto:
		return &tele.Photo{
			File:         file,
			HasSpoiler:   p.HasMediaSpoiler,
			CaptionAbove: p.ShowCaptionAboveMedia,
		}
	case entity.MediaTypeVideo:
		return &tele.Video{
			File:         file,
			HasSpoiler:   p.HasMediaSpoiler,
			CaptionAbove: p.ShowCaptionAboveMedia,
		}
	case entity.MediaTypeDocument:
		return &tele.Document{File: file}
	case entity.MediaTypeAudio:
		return &tele.Audio{File: file}
	default:
		return nil
	}
}

func setAlbumItemCaption(item tele.Inputtable, caption string) {
	switch v := item.(type) {
	case *tele.Photo:
		v.Caption = caption
	case *tele.Video:
		v.Caption = caption
	case *tele.Document:
		v.Caption = caption
	case *tele.Audio:
		v.Caption = caption
	}
}

func groupPosts(posts []entity.Post) []dto.TemplateResponse {
	var templates []dto.TemplateResponse
	grouped := make(map[string]int)

	for i := range posts {
		p := &posts[i]

		if p.MediaGroupID != nil {
			idx, exists := grouped[*p.MediaGroupID]
			if exists {
				if p.MediaType != nil {
					templates[idx].Media = append(templates[idx].Media, dto.PostMediaItem{
						PostID:                p.ID.String(),
						MediaType:             *p.MediaType,
						HasMediaSpoiler:       p.HasMediaSpoiler,
						ShowCaptionAboveMedia: p.ShowCaptionAboveMedia,
					})
				}
				if p.Text != nil && templates[idx].Text == nil {
					templates[idx].Text = p.Text
					if len(p.Entities) > 0 && string(p.Entities) != "null" {
						templates[idx].Entities = json.RawMessage(p.Entities)
					}
				}
				continue
			}
		}

		tmpl := dto.TemplateResponse{
			ID:        p.ID.String(),
			Name:      p.Name,
			Text:      p.Text,
			CreatedAt: p.CreatedAt,
		}
		if len(p.Entities) > 0 && string(p.Entities) != "null" {
			tmpl.Entities = json.RawMessage(p.Entities)
		}
		if p.MediaType != nil {
			tmpl.Media = []dto.PostMediaItem{{
				PostID:                p.ID.String(),
				MediaType:             *p.MediaType,
				HasMediaSpoiler:       p.HasMediaSpoiler,
				ShowCaptionAboveMedia: p.ShowCaptionAboveMedia,
			}}
		}

		if p.MediaGroupID != nil {
			grouped[*p.MediaGroupID] = len(templates)
		}
		templates = append(templates, tmpl)
	}

	return templates
}
