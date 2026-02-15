package post

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/google/uuid"

	"github.com/bpva/ad-marketplace/internal/dto"
	"github.com/bpva/ad-marketplace/internal/entity"
	"github.com/bpva/ad-marketplace/internal/logx"
)

type PostRepository interface {
	GetByUserID(ctx context.Context, userID uuid.UUID) ([]entity.Post, error)
	GetByID(ctx context.Context, id uuid.UUID) (*entity.Post, error)
}

type TelebotClient interface {
	DownloadFile(fileID string) ([]byte, error)
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

	posts, err := s.postRepo.GetByUserID(ctx, user.ID)
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

	if post.UserID != user.ID {
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
