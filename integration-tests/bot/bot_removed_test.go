//go:build integration

package bot_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	tele "gopkg.in/telebot.v4"

	"github.com/bpva/ad-marketplace/internal/config"
	"github.com/bpva/ad-marketplace/internal/entity"
	"github.com/bpva/ad-marketplace/internal/service/bot"
)

func TestHandleBotRemoved(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name         string
		setup        func(t *testing.T)
		update       tele.Update
		checkChannel func(t *testing.T)
	}{
		{
			name: "bot removed (left) from existing channel",
			setup: func(t *testing.T) {
				user, err := testTools.CreateUser(ctx, 111222333, "Owner")
				require.NoError(t, err)

				channel, err := testTools.CreateChannel(ctx, -1001234567890, "Test Channel", nil)
				require.NoError(t, err)

				_, err = testTools.CreateChannelRole(
					ctx,
					channel.ID,
					user.ID,
					entity.ChannelRoleTypeOwner,
				)
				require.NoError(t, err)
			},
			update: createBotRemovedUpdate(-1001234567890, "Test Channel", tele.Left),
			checkChannel: func(t *testing.T) {
				deleted, err := testTools.IsChannelSoftDeleted(ctx, -1001234567890)
				require.NoError(t, err)
				assert.True(t, deleted)
			},
		},
		{
			name: "bot kicked from existing channel",
			setup: func(t *testing.T) {
				user, err := testTools.CreateUser(ctx, 111222333, "Owner")
				require.NoError(t, err)

				channel, err := testTools.CreateChannel(ctx, -1001234567890, "Test Channel", nil)
				require.NoError(t, err)

				_, err = testTools.CreateChannelRole(
					ctx,
					channel.ID,
					user.ID,
					entity.ChannelRoleTypeOwner,
				)
				require.NoError(t, err)
			},
			update: createBotRemovedUpdate(-1001234567890, "Test Channel", tele.Kicked),
			checkChannel: func(t *testing.T) {
				deleted, err := testTools.IsChannelSoftDeleted(ctx, -1001234567890)
				require.NoError(t, err)
				assert.True(t, deleted)
			},
		},
		{
			name:   "bot removed from non-existing channel",
			setup:  func(t *testing.T) {},
			update: createBotRemovedUpdate(-1001234567890, "Unknown Channel", tele.Left),
			checkChannel: func(t *testing.T) {
				_, err := testTools.GetChannelByTgID(ctx, -1001234567890)
				assert.ErrorIs(t, err, pgx.ErrNoRows)
			},
		},
		{
			name:   "bot removed from group (not channel)",
			setup:  func(t *testing.T) {},
			update: createBotRemovedFromGroupUpdate(-1001234567890, "Test Group", tele.Left),
			checkChannel: func(t *testing.T) {
				_, err := testTools.GetChannelByTgID(ctx, -1001234567890)
				assert.ErrorIs(t, err, pgx.ErrNoRows)
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

			botSvc := bot.New(mock, config.Telegram{}, log, testDB, channelRepo, userRepo, statsSvc)

			tt.setup(t)

			mock.EXPECT().ProcessUpdate(tt.update).Do(func(upd tele.Update) {
				if upd.MyChatMember != nil {
					botSvc.HandleChatMemberUpdate(upd.MyChatMember)
				}
			})

			data, _ := json.Marshal(tt.update)
			err := botSvc.ProcessUpdate(data)
			require.NoError(t, err)

			tt.checkChannel(t)
		})
	}
}

func createBotRemovedUpdate(chatID int64, title string, status tele.MemberStatus) tele.Update {
	return tele.Update{
		MyChatMember: &tele.ChatMemberUpdate{
			Chat: &tele.Chat{
				ID:    chatID,
				Type:  tele.ChatChannel,
				Title: title,
			},
			Sender: &tele.User{
				ID:        999999,
				FirstName: "ChannelOwner",
			},
			NewChatMember: &tele.ChatMember{
				User: &tele.User{
					ID:        12345,
					FirstName: "Bot",
				},
				Role: status,
			},
		},
	}
}

func createBotRemovedFromGroupUpdate(
	chatID int64,
	title string,
	status tele.MemberStatus,
) tele.Update {
	return tele.Update{
		MyChatMember: &tele.ChatMemberUpdate{
			Chat: &tele.Chat{
				ID:    chatID,
				Type:  tele.ChatGroup,
				Title: title,
			},
			Sender: &tele.User{
				ID:        999999,
				FirstName: "GroupOwner",
			},
			NewChatMember: &tele.ChatMember{
				User: &tele.User{
					ID:        12345,
					FirstName: "Bot",
				},
				Role: status,
			},
		},
	}
}
