//go:build integration

package bot_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	tele "gopkg.in/telebot.v4"

	"github.com/bpva/ad-marketplace/internal/config"
	"github.com/bpva/ad-marketplace/internal/entity"
	"github.com/bpva/ad-marketplace/internal/service/bot"
)

func TestHandleAddPromo(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name       string
		updates    []tele.Update
		checkPosts func(t *testing.T)
	}{
		{
			name: "text-only post saved",
			updates: []tele.Update{
				createCommandUpdate(111222333, "/add_promo"),
				createTextUpdate(111222333, "Check out our amazing product!"),
			},
			checkPosts: func(t *testing.T) {
				user, err := testTools.GetUserByTgID(ctx, 111222333)
				require.NoError(t, err)

				posts, err := testTools.GetPostsByUserID(ctx, user.ID)
				require.NoError(t, err)
				require.Len(t, posts, 1)
				require.NotNil(t, posts[0].Text)
				assert.Equal(t, "Check out our amazing product!", *posts[0].Text)
				assert.Nil(t, posts[0].MediaType)
				assert.Nil(t, posts[0].MediaFileID)
			},
		},
		{
			name: "photo with caption saved",
			updates: []tele.Update{
				createCommandUpdate(111222333, "/add_promo"),
				createPhotoUpdate(111222333, "photo-file-123", "Look at this!", ""),
			},
			checkPosts: func(t *testing.T) {
				user, err := testTools.GetUserByTgID(ctx, 111222333)
				require.NoError(t, err)

				posts, err := testTools.GetPostsByUserID(ctx, user.ID)
				require.NoError(t, err)
				require.Len(t, posts, 1)
				require.NotNil(t, posts[0].Text)
				assert.Equal(t, "Look at this!", *posts[0].Text)
				require.NotNil(t, posts[0].MediaType)
				assert.Equal(t, entity.MediaTypePhoto, *posts[0].MediaType)
				require.NotNil(t, posts[0].MediaFileID)
				assert.Equal(t, "photo-file-123", *posts[0].MediaFileID)
			},
		},
		{
			name: "message without /add_promo creates no post",
			updates: []tele.Update{
				createTextUpdate(111222333, "hello"),
			},
			checkPosts: func(t *testing.T) {
				user, err := testTools.GetUserByTgID(ctx, 111222333)
				if err != nil {
					return
				}
				posts, err := testTools.GetPostsByUserID(ctx, user.ID)
				require.NoError(t, err)
				assert.Empty(t, posts)
			},
		},
		{
			name: "album saved as multiple rows",
			updates: []tele.Update{
				createCommandUpdate(111222333, "/add_promo"),
				createPhotoUpdate(111222333, "photo-1", "Album caption", "album-123"),
				createPhotoUpdate(111222333, "photo-2", "", "album-123"),
				createPhotoUpdate(111222333, "photo-3", "", "album-123"),
			},
			checkPosts: func(t *testing.T) {
				user, err := testTools.GetUserByTgID(ctx, 111222333)
				require.NoError(t, err)

				posts, err := testTools.GetPostsByUserID(ctx, user.ID)
				require.NoError(t, err)
				require.Len(t, posts, 3)

				for _, p := range posts {
					require.NotNil(t, p.MediaGroupID)
					assert.Equal(t, "album-123", *p.MediaGroupID)
					require.NotNil(t, p.MediaType)
					assert.Equal(t, entity.MediaTypePhoto, *p.MediaType)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.NoError(t, testTools.TruncateAll(ctx))

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mock := bot.NewMockTelebotClient(ctrl)
			mock.EXPECT().Handle(gomock.Any(), gomock.Any()).AnyTimes()

			botSvc := bot.New(
				mock,
				config.Telegram{},
				log,
				testDB,
				channelRepo,
				userRepo,
				statsSvc,
				postRepo,
			)

			for _, upd := range tt.updates {
				botSvc.HandleUpdate(upd)
			}

			tt.checkPosts(t)
		})
	}
}

func createCommandUpdate(senderID int64, command string) tele.Update {
	return tele.Update{
		Message: &tele.Message{
			Chat: &tele.Chat{
				ID:   senderID,
				Type: tele.ChatPrivate,
			},
			Sender: &tele.User{
				ID:        senderID,
				FirstName: "TestUser",
			},
			Text: command,
		},
	}
}

func createTextUpdate(senderID int64, text string) tele.Update {
	return tele.Update{
		Message: &tele.Message{
			Chat: &tele.Chat{
				ID:   senderID,
				Type: tele.ChatPrivate,
			},
			Sender: &tele.User{
				ID:        senderID,
				FirstName: "TestUser",
			},
			Text: text,
		},
	}
}

func createPhotoUpdate(senderID int64, fileID, caption, albumID string) tele.Update {
	return tele.Update{
		Message: &tele.Message{
			Chat: &tele.Chat{
				ID:   senderID,
				Type: tele.ChatPrivate,
			},
			Sender: &tele.User{
				ID:        senderID,
				FirstName: "TestUser",
			},
			Photo: &tele.Photo{
				File: tele.File{
					FileID: fileID,
				},
			},
			Caption: caption,
			AlbumID: albumID,
		},
	}
}
