//go:build integration

package http_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/bpva/ad-marketplace/internal/dto"
	"github.com/bpva/ad-marketplace/internal/entity"
)

func TestHandleListTemplates(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name           string
		setup          func(t *testing.T) string
		expectedStatus int
		check          func(t *testing.T, body []byte)
	}{
		{
			name: "empty list",
			setup: func(t *testing.T) string {
				user, err := testTools.CreateUser(ctx, 2001001, "Empty Posts User")
				require.NoError(t, err)
				token, err := testTools.GenerateToken(user)
				require.NoError(t, err)
				return "Bearer " + token
			},
			expectedStatus: http.StatusOK,
			check: func(t *testing.T, body []byte) {
				var resp dto.TemplatesResponse
				require.NoError(t, json.Unmarshal(body, &resp))
				assert.Empty(t, resp.Templates)
			},
		},
		{
			name: "text-only post",
			setup: func(t *testing.T) string {
				user, err := testTools.CreateUser(ctx, 2001002, "Text Post User")
				require.NoError(t, err)

				text := "Hello **world**"
				entities := []byte(`[{"type":"bold","offset":6,"length":9}]`)
				_, err = testTools.CreatePost(ctx, user.ID, nil, &text, entities, nil, nil)
				require.NoError(t, err)

				token, err := testTools.GenerateToken(user)
				require.NoError(t, err)
				return "Bearer " + token
			},
			expectedStatus: http.StatusOK,
			check: func(t *testing.T, body []byte) {
				var resp dto.TemplatesResponse
				require.NoError(t, json.Unmarshal(body, &resp))
				require.Len(t, resp.Templates, 1)
				assert.NotEmpty(t, resp.Templates[0].ID)
				assert.Equal(t, "Hello **world**", *resp.Templates[0].Text)
				assert.NotNil(t, resp.Templates[0].Entities)
				assert.Empty(t, resp.Templates[0].Media)
			},
		},
		{
			name: "album grouping",
			setup: func(t *testing.T) string {
				user, err := testTools.CreateUser(ctx, 2001003, "Album User")
				require.NoError(t, err)

				groupID := "group123"
				caption := "Album caption"
				photoType := entity.MediaTypePhoto
				fileID1 := "file1"
				fileID2 := "file2"

				_, err = testTools.CreatePost(
					ctx,
					user.ID,
					&groupID,
					&caption,
					nil,
					&photoType,
					&fileID1,
				)
				require.NoError(t, err)
				_, err = testTools.CreatePost(
					ctx,
					user.ID,
					&groupID,
					nil,
					nil,
					&photoType,
					&fileID2,
				)
				require.NoError(t, err)

				token, err := testTools.GenerateToken(user)
				require.NoError(t, err)
				return "Bearer " + token
			},
			expectedStatus: http.StatusOK,
			check: func(t *testing.T, body []byte) {
				var resp dto.TemplatesResponse
				require.NoError(t, json.Unmarshal(body, &resp))
				require.Len(t, resp.Templates, 1)
				assert.Equal(t, "Album caption", *resp.Templates[0].Text)
				assert.Len(t, resp.Templates[0].Media, 2)
			},
		},
		{
			name: "ownership isolation",
			setup: func(t *testing.T) string {
				user1, err := testTools.CreateUser(ctx, 2001004, "Owner One")
				require.NoError(t, err)
				user2, err := testTools.CreateUser(ctx, 2001005, "Owner Two")
				require.NoError(t, err)

				text1 := "User 1 post"
				text2 := "User 2 post"
				_, err = testTools.CreatePost(ctx, user1.ID, nil, &text1, nil, nil, nil)
				require.NoError(t, err)
				_, err = testTools.CreatePost(ctx, user2.ID, nil, &text2, nil, nil, nil)
				require.NoError(t, err)

				token, err := testTools.GenerateToken(user1)
				require.NoError(t, err)
				return "Bearer " + token
			},
			expectedStatus: http.StatusOK,
			check: func(t *testing.T, body []byte) {
				var resp dto.TemplatesResponse
				require.NoError(t, json.Unmarshal(body, &resp))
				require.Len(t, resp.Templates, 1)
				assert.Equal(t, "User 1 post", *resp.Templates[0].Text)
			},
		},
		{
			name: "unauthorized",
			setup: func(t *testing.T) string {
				return ""
			},
			expectedStatus: http.StatusUnauthorized,
			check:          func(t *testing.T, body []byte) {},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.NoError(t, testTools.TruncateAll(ctx))

			authHeader := tt.setup(t)

			req, err := http.NewRequest(http.MethodGet, testServer.URL+"/api/v1/posts", nil)
			require.NoError(t, err)
			if authHeader != "" {
				req.Header.Set("Authorization", authHeader)
			}

			resp, err := http.DefaultClient.Do(req)
			require.NoError(t, err)
			defer resp.Body.Close()

			body, err := io.ReadAll(resp.Body)
			require.NoError(t, err)

			assert.Equal(t, tt.expectedStatus, resp.StatusCode)
			tt.check(t, body)
		})
	}
}

func TestHandleGetPostMedia(t *testing.T) {
	ctx := context.Background()

	t.Run("returns media bytes", func(t *testing.T) {
		require.NoError(t, testTools.TruncateAll(ctx))

		user, err := testTools.CreateUser(ctx, 2002001, "Media User")
		require.NoError(t, err)

		photoType := entity.MediaTypePhoto
		fileID := "test_file_id"
		post, err := testTools.CreatePost(ctx, user.ID, nil, nil, nil, &photoType, &fileID)
		require.NoError(t, err)

		token, err := testTools.GenerateToken(user)
		require.NoError(t, err)

		req, err := http.NewRequest(
			http.MethodGet,
			testServer.URL+"/api/v1/posts/"+post.ID.String()+"/media",
			nil,
		)
		require.NoError(t, err)
		req.Header.Set("Authorization", "Bearer "+token)

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.Equal(t, "public, max-age=86400", resp.Header.Get("Cache-Control"))
	})

	t.Run("forbidden for other user", func(t *testing.T) {
		require.NoError(t, testTools.TruncateAll(ctx))

		user1, err := testTools.CreateUser(ctx, 2002002, "Post Owner")
		require.NoError(t, err)
		user2, err := testTools.CreateUser(ctx, 2002003, "Other User")
		require.NoError(t, err)

		photoType := entity.MediaTypePhoto
		fileID := "test_file_id"
		post, err := testTools.CreatePost(ctx, user1.ID, nil, nil, nil, &photoType, &fileID)
		require.NoError(t, err)

		token, err := testTools.GenerateToken(user2)
		require.NoError(t, err)

		req, err := http.NewRequest(
			http.MethodGet,
			testServer.URL+"/api/v1/posts/"+post.ID.String()+"/media",
			nil,
		)
		require.NoError(t, err)
		req.Header.Set("Authorization", "Bearer "+token)

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusForbidden, resp.StatusCode)
	})
}
