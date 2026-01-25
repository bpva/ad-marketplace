//go:build integration

package bot_test

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	tele "gopkg.in/telebot.v4"

	"github.com/bpva/ad-marketplace/internal/dto"
	"github.com/bpva/ad-marketplace/internal/entity"
	bot_service "github.com/bpva/ad-marketplace/internal/service/bot"
)

func TestHandleBotAdded(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name         string
		setup        func(t *testing.T, mock *bot_service.MockTelebotClient)
		update       tele.Update
		checkUser    func(t *testing.T)
		checkChannel func(t *testing.T)
	}{
		{
			name: "bot added with post permission, new user",
			setup: func(t *testing.T, mock *bot_service.MockTelebotClient) {
				mock.EXPECT().AdminsOf(int64(-1001234567890)).Return([]dto.ChannelAdmin{
					{
						TelegramID: 111222333,
						FirstName:  "Channel",
						LastName:   "Owner",
						Role:       dto.RoleCreator,
					},
				}, nil)
			},
			update: createBotAddedUpdate(-1001234567890, "Test Channel", "testchannel", true),
			checkUser: func(t *testing.T) {
				user, err := testTools.GetUserByTelegramID(ctx, 111222333)
				require.NoError(t, err)
				assert.Equal(t, int64(111222333), user.TelegramID)
				assert.Equal(t, "Channel Owner", user.Name)
			},
			checkChannel: func(t *testing.T) {
				channel, err := testTools.GetChannelByTelegramID(ctx, -1001234567890)
				require.NoError(t, err)
				assert.Equal(t, int64(-1001234567890), channel.TelegramChannelID)
				assert.Equal(t, "Test Channel", channel.Title)
				assert.NotNil(t, channel.Username)
				assert.Equal(t, "testchannel", *channel.Username)
				assert.Nil(t, channel.DeletedAt)

				roles, err := testTools.GetChannelRolesByChannelID(ctx, channel.ID)
				require.NoError(t, err)
				require.Len(t, roles, 1)
				assert.Equal(t, entity.ChannelRoleTypeOwner, roles[0].Role)
			},
		},
		{
			name: "bot added with post permission, existing user",
			setup: func(t *testing.T, mock *bot_service.MockTelebotClient) {
				_, err := testTools.CreateUser(ctx, 111222333, "Existing User")
				require.NoError(t, err)

				mock.EXPECT().AdminsOf(int64(-1001234567890)).Return([]dto.ChannelAdmin{
					{
						TelegramID: 111222333,
						FirstName:  "Existing",
						LastName:   "User",
						Role:       dto.RoleCreator,
					},
				}, nil)
			},
			update: createBotAddedUpdate(-1001234567890, "Test Channel", "testchannel", true),
			checkUser: func(t *testing.T) {
				user, err := testTools.GetUserByTelegramID(ctx, 111222333)
				require.NoError(t, err)
				assert.Equal(t, "Existing User", user.Name)
			},
			checkChannel: func(t *testing.T) {
				channel, err := testTools.GetChannelByTelegramID(ctx, -1001234567890)
				require.NoError(t, err)
				assert.Equal(t, "Test Channel", channel.Title)

				user, err := testTools.GetUserByTelegramID(ctx, 111222333)
				require.NoError(t, err)

				roles, err := testTools.GetChannelRolesByChannelID(ctx, channel.ID)
				require.NoError(t, err)
				require.Len(t, roles, 1)
				assert.Equal(t, user.ID, roles[0].UserID)
			},
		},
		{
			name:   "bot added without post permission",
			setup:  func(t *testing.T, mock *bot_service.MockTelebotClient) {},
			update: createBotAddedUpdate(-1001234567890, "Test Channel", "testchannel", false),
			checkUser: func(t *testing.T) {
				_, err := testTools.GetUserByTelegramID(ctx, 111222333)
				assert.ErrorIs(t, err, pgx.ErrNoRows)
			},
			checkChannel: func(t *testing.T) {
				_, err := testTools.GetChannelByTelegramID(ctx, -1001234567890)
				assert.ErrorIs(t, err, pgx.ErrNoRows)
			},
		},
		{
			name:   "bot added to group (not channel)",
			setup:  func(t *testing.T, mock *bot_service.MockTelebotClient) {},
			update: createBotAddedToGroupUpdate(-1001234567890, "Test Group", true),
			checkUser: func(t *testing.T) {
				_, err := testTools.GetUserByTelegramID(ctx, 111222333)
				assert.ErrorIs(t, err, pgx.ErrNoRows)
			},
			checkChannel: func(t *testing.T) {
				_, err := testTools.GetChannelByTelegramID(ctx, -1001234567890)
				assert.ErrorIs(t, err, pgx.ErrNoRows)
			},
		},
		{
			name: "bot added but no creator in admins",
			setup: func(t *testing.T, mock *bot_service.MockTelebotClient) {
				mock.EXPECT().AdminsOf(int64(-1001234567890)).Return([]dto.ChannelAdmin{
					{
						TelegramID: 999,
						FirstName:  "Admin",
						Role:       dto.RoleAdministrator,
					},
				}, nil)
			},
			update: createBotAddedUpdate(-1001234567890, "Test Channel", "testchannel", true),
			checkUser: func(t *testing.T) {
				_, err := testTools.GetUserByTelegramID(ctx, 999)
				assert.ErrorIs(t, err, pgx.ErrNoRows)
			},
			checkChannel: func(t *testing.T) {
				_, err := testTools.GetChannelByTelegramID(ctx, -1001234567890)
				assert.ErrorIs(t, err, pgx.ErrNoRows)
			},
		},
		{
			name: "bot added but AdminsOf API fails",
			setup: func(t *testing.T, mock *bot_service.MockTelebotClient) {
				mock.EXPECT().
					AdminsOf(int64(-1001234567890)).
					Return(nil, errors.New("telegram api error"))
			},
			update: createBotAddedUpdate(-1001234567890, "Test Channel", "testchannel", true),
			checkUser: func(t *testing.T) {
				_, err := testTools.GetUserByTelegramID(ctx, 111222333)
				assert.ErrorIs(t, err, pgx.ErrNoRows)
			},
			checkChannel: func(t *testing.T) {
				_, err := testTools.GetChannelByTelegramID(ctx, -1001234567890)
				assert.ErrorIs(t, err, pgx.ErrNoRows)
			},
		},
		{
			name: "bot re-added to deleted channel",
			setup: func(t *testing.T, mock *bot_service.MockTelebotClient) {
				user, err := testTools.CreateUser(ctx, 111222333, "Owner")
				require.NoError(t, err)

				channel, err := testTools.CreateChannel(ctx, -1001234567890, "Old Title", nil)
				require.NoError(t, err)

				_, err = testTools.CreateChannelRole(
					ctx,
					channel.ID,
					user.ID,
					entity.ChannelRoleTypeOwner,
				)
				require.NoError(t, err)

				err = testTools.SoftDeleteChannel(ctx, -1001234567890)
				require.NoError(t, err)

				mock.EXPECT().AdminsOf(int64(-1001234567890)).Return([]dto.ChannelAdmin{
					{
						TelegramID: 111222333,
						FirstName:  "Owner",
						Role:       dto.RoleCreator,
					},
				}, nil)
			},
			update: createBotAddedUpdate(-1001234567890, "Updated Title", "newusername", true),
			checkUser: func(t *testing.T) {
				user, err := testTools.GetUserByTelegramID(ctx, 111222333)
				require.NoError(t, err)
				assert.Equal(t, "Owner", user.Name)
			},
			checkChannel: func(t *testing.T) {
				channel, err := testTools.GetChannelByTelegramID(ctx, -1001234567890)
				require.NoError(t, err)
				assert.Equal(t, "Updated Title", channel.Title)
				assert.NotNil(t, channel.Username)
				assert.Equal(t, "newusername", *channel.Username)
				assert.Nil(t, channel.DeletedAt)
			},
		},
		{
			name: "channel without username",
			setup: func(t *testing.T, mock *bot_service.MockTelebotClient) {
				mock.EXPECT().AdminsOf(int64(-1001234567890)).Return([]dto.ChannelAdmin{
					{
						TelegramID: 111222333,
						FirstName:  "Owner",
						Role:       dto.RoleCreator,
					},
				}, nil)
			},
			update: createBotAddedUpdate(-1001234567890, "Private Channel", "", true),
			checkUser: func(t *testing.T) {
				user, err := testTools.GetUserByTelegramID(ctx, 111222333)
				require.NoError(t, err)
				assert.Equal(t, "Owner", user.Name)
			},
			checkChannel: func(t *testing.T) {
				channel, err := testTools.GetChannelByTelegramID(ctx, -1001234567890)
				require.NoError(t, err)
				assert.Equal(t, "Private Channel", channel.Title)
				assert.Nil(t, channel.Username)
			},
		},
		{
			name: "creator with first name only",
			setup: func(t *testing.T, mock *bot_service.MockTelebotClient) {
				mock.EXPECT().AdminsOf(int64(-1001234567890)).Return([]dto.ChannelAdmin{
					{
						TelegramID: 111222333,
						FirstName:  "SingleName",
						Role:       dto.RoleCreator,
					},
				}, nil)
			},
			update: createBotAddedUpdate(-1001234567890, "Test Channel", "testchannel", true),
			checkUser: func(t *testing.T) {
				user, err := testTools.GetUserByTelegramID(ctx, 111222333)
				require.NoError(t, err)
				assert.Equal(t, "SingleName", user.Name)
			},
			checkChannel: func(t *testing.T) {
				channel, err := testTools.GetChannelByTelegramID(ctx, -1001234567890)
				require.NoError(t, err)
				assert.NotNil(t, channel)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.NoError(t, testTools.TruncateAll(ctx))

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mock := bot_service.NewMockTelebotClient(ctrl)
			mock.EXPECT().Handle(gomock.Any(), gomock.Any()).AnyTimes()

			botSvc := bot_service.New(mock, "http://localhost", log, testDB, channelRepo, userRepo)

			tt.setup(t, mock)

			mock.EXPECT().ProcessUpdate(tt.update).Do(func(upd tele.Update) {
				if upd.MyChatMember != nil {
					botSvc.HandleChatMemberUpdate(upd.MyChatMember)
				}
			})

			err := botSvc.ProcessUpdate(mustMarshal(tt.update))
			require.NoError(t, err)

			tt.checkUser(t)
			tt.checkChannel(t)
		})
	}
}

func mustMarshal(v any) []byte {
	data, _ := json.Marshal(v)
	return data
}

func createBotAddedUpdate(chatID int64, title, username string, canPostMessages bool) tele.Update {
	return tele.Update{
		MyChatMember: &tele.ChatMemberUpdate{
			Chat: &tele.Chat{
				ID:       chatID,
				Type:     tele.ChatChannel,
				Title:    title,
				Username: username,
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
				Role: tele.Administrator,
				Rights: tele.Rights{
					CanPostMessages: canPostMessages,
				},
			},
		},
	}
}

func createBotAddedToGroupUpdate(chatID int64, title string, canPostMessages bool) tele.Update {
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
				Role: tele.Administrator,
				Rights: tele.Rights{
					CanPostMessages: canPostMessages,
				},
			},
		},
	}
}
