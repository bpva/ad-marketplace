package app

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/bpva/ad-marketplace/internal/dto"
	"github.com/bpva/ad-marketplace/internal/http/respond"
	"github.com/bpva/ad-marketplace/internal/logx"
)

// HandleListTemplates returns user's saved post templates
//
//	@Summary		List templates
//	@Tags			posts
//	@Produce		json
//	@Security		BearerAuth
//	@Success		200	{object}	dto.TemplatesResponse
//	@Failure		401	{object}	dto.ErrorResponse
//	@Router			/posts [get]
func (a *App) HandleListTemplates() http.HandlerFunc {
	log := a.log.With(logx.Handler("/api/v1/posts"))

	return func(w http.ResponseWriter, r *http.Request) {
		templates, err := a.post.GetUserTemplates(r.Context())
		if err != nil {
			respond.Err(w, log, err)
			return
		}

		respond.OK(w, templates)
	}
}

func (a *App) HandleGetPostMedia() http.HandlerFunc {
	log := a.log.With(logx.Handler("/api/v1/posts/{postID}/media"))

	return func(w http.ResponseWriter, r *http.Request) {
		postID, err := uuid.Parse(chi.URLParam(r, "postID"))
		if err != nil {
			respond.Err(w, log, dto.ErrBadRequest)
			return
		}

		data, err := a.post.GetPostMedia(r.Context(), postID)
		if err != nil {
			respond.Err(w, log, err)
			return
		}

		w.Header().Set("Content-Type", "application/octet-stream")
		w.Header().Set("Cache-Control", "public, max-age=86400")
		w.WriteHeader(http.StatusOK)
		_, err = w.Write(data)
		if err != nil {
			log.Error("Failed to write post media", "error", err)
		}
	}
}
