//go:build integration

package http_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/bpva/ad-marketplace/internal/dto"
	"github.com/bpva/ad-marketplace/internal/entity"
)

func TestHandleListChannels(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name           string
		setup          func(t *testing.T) string
		expectedStatus int
		check          func(t *testing.T, body []byte)
	}{
		{
			name: "user with no channels",
			setup: func(t *testing.T) string {
				user, err := testTools.CreateUser(ctx, 1001001, "No Channels User")
				require.NoError(t, err)
				token, err := testTools.GenerateToken(user)
				require.NoError(t, err)
				return "Bearer " + token
			},
			expectedStatus: http.StatusOK,
			check: func(t *testing.T, body []byte) {
				var resp dto.ChannelsResponse
				require.NoError(t, json.Unmarshal(body, &resp))
				assert.Empty(t, resp.Channels)
			},
		},
		{
			name: "user with multiple channels",
			setup: func(t *testing.T) string {
				user, err := testTools.CreateUser(ctx, 1001002, "Multi Channel User")
				require.NoError(t, err)

				ch1, err := testTools.CreateChannel(ctx, -1001001002001, "Channel One", nil)
				require.NoError(t, err)
				_, err = testTools.CreateChannelRole(
					ctx,
					ch1.ID,
					user.ID,
					entity.ChannelRoleTypeOwner,
				)
				require.NoError(t, err)

				username := "channel_two"
				ch2, err := testTools.CreateChannel(ctx, -1001001002002, "Channel Two", &username)
				require.NoError(t, err)
				_, err = testTools.CreateChannelRole(
					ctx,
					ch2.ID,
					user.ID,
					entity.ChannelRoleTypeManager,
				)
				require.NoError(t, err)

				token, err := testTools.GenerateToken(user)
				require.NoError(t, err)
				return "Bearer " + token
			},
			expectedStatus: http.StatusOK,
			check: func(t *testing.T, body []byte) {
				var resp dto.ChannelsResponse
				require.NoError(t, json.Unmarshal(body, &resp))
				assert.Len(t, resp.Channels, 2)
			},
		},
		{
			name: "soft-deleted channel excluded",
			setup: func(t *testing.T) string {
				user, err := testTools.CreateUser(ctx, 1001003, "Deleted Channel User")
				require.NoError(t, err)

				ch, err := testTools.CreateChannel(ctx, -1001001003001, "Deleted Channel", nil)
				require.NoError(t, err)
				_, err = testTools.CreateChannelRole(
					ctx,
					ch.ID,
					user.ID,
					entity.ChannelRoleTypeOwner,
				)
				require.NoError(t, err)
				err = testTools.SoftDeleteChannel(ctx, ch.TgChannelID)
				require.NoError(t, err)

				token, err := testTools.GenerateToken(user)
				require.NoError(t, err)
				return "Bearer " + token
			},
			expectedStatus: http.StatusOK,
			check: func(t *testing.T, body []byte) {
				var resp dto.ChannelsResponse
				require.NoError(t, json.Unmarshal(body, &resp))
				assert.Empty(t, resp.Channels)
			},
		},
		{
			name:           "unauthenticated",
			setup:          func(t *testing.T) string { return "" },
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			authHeader := tt.setup(t)

			req, err := http.NewRequest(http.MethodGet, testServer.URL+"/api/v1/channels", nil)
			require.NoError(t, err)
			if authHeader != "" {
				req.Header.Set("Authorization", authHeader)
			}

			resp, err := http.DefaultClient.Do(req)
			require.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, tt.expectedStatus, resp.StatusCode)

			if tt.check != nil {
				body, err := io.ReadAll(resp.Body)
				require.NoError(t, err)
				tt.check(t, body)
			}
		})
	}
}

func TestHandleGetChannel(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name           string
		setup          func(t *testing.T) (string, int64)
		expectedStatus int
		check          func(t *testing.T, body []byte)
	}{
		{
			name: "valid channel authorized user",
			setup: func(t *testing.T) (string, int64) {
				user, err := testTools.CreateUser(ctx, 2001001, "Owner User")
				require.NoError(t, err)

				username := "test_channel"
				ch, err := testTools.CreateChannel(ctx, -1002001001001, "Test Channel", &username)
				require.NoError(t, err)
				_, err = testTools.CreateChannelRole(
					ctx,
					ch.ID,
					user.ID,
					entity.ChannelRoleTypeOwner,
				)
				require.NoError(t, err)

				token, err := testTools.GenerateToken(user)
				require.NoError(t, err)
				return "Bearer " + token, ch.TgChannelID
			},
			expectedStatus: http.StatusOK,
			check: func(t *testing.T, body []byte) {
				var resp dto.ChannelResponse
				require.NoError(t, json.Unmarshal(body, &resp))
				assert.Equal(t, int64(-1002001001001), resp.TgChannelID)
				assert.Equal(t, "Test Channel", resp.Title)
				assert.Equal(t, "test_channel", resp.Username)
			},
		},
		{
			name: "invalid channel id",
			setup: func(t *testing.T) (string, int64) {
				user, err := testTools.CreateUser(ctx, 2001002, "Some User")
				require.NoError(t, err)
				token, err := testTools.GenerateToken(user)
				require.NoError(t, err)
				return "Bearer " + token, 0
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "non-existent channel",
			setup: func(t *testing.T) (string, int64) {
				user, err := testTools.CreateUser(ctx, 2001003, "Some User")
				require.NoError(t, err)
				token, err := testTools.GenerateToken(user)
				require.NoError(t, err)
				return "Bearer " + token, -1002001003999
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			name: "user without role in channel",
			setup: func(t *testing.T) (string, int64) {
				owner, err := testTools.CreateUser(ctx, 2001004, "Owner")
				require.NoError(t, err)
				ch, err := testTools.CreateChannel(ctx, -1002001004001, "Private Channel", nil)
				require.NoError(t, err)
				_, err = testTools.CreateChannelRole(
					ctx,
					ch.ID,
					owner.ID,
					entity.ChannelRoleTypeOwner,
				)
				require.NoError(t, err)

				outsider, err := testTools.CreateUser(ctx, 2001005, "Outsider")
				require.NoError(t, err)
				token, err := testTools.GenerateToken(outsider)
				require.NoError(t, err)
				return "Bearer " + token, ch.TgChannelID
			},
			expectedStatus: http.StatusForbidden,
		},
		{
			name: "unauthenticated",
			setup: func(t *testing.T) (string, int64) {
				return "", -1002001006001
			},
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			authHeader, channelID := tt.setup(t)

			var url string
			if channelID == 0 {
				url = testServer.URL + "/api/v1/channels/invalid"
			} else {
				url = fmt.Sprintf("%s/api/v1/channels/%d", testServer.URL, channelID)
			}

			req, err := http.NewRequest(http.MethodGet, url, nil)
			require.NoError(t, err)
			if authHeader != "" {
				req.Header.Set("Authorization", authHeader)
			}

			resp, err := http.DefaultClient.Do(req)
			require.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, tt.expectedStatus, resp.StatusCode)

			if tt.check != nil {
				body, err := io.ReadAll(resp.Body)
				require.NoError(t, err)
				tt.check(t, body)
			}
		})
	}
}

func TestHandleGetChannelAdmins(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name           string
		setup          func(t *testing.T) (string, int64)
		expectedStatus int
		check          func(t *testing.T, body []byte)
	}{
		{
			name: "invalid channel id",
			setup: func(t *testing.T) (string, int64) {
				user, err := testTools.CreateUser(ctx, 3001002, "Some User")
				require.NoError(t, err)
				token, err := testTools.GenerateToken(user)
				require.NoError(t, err)
				return "Bearer " + token, 0
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "non-existent channel",
			setup: func(t *testing.T) (string, int64) {
				user, err := testTools.CreateUser(ctx, 3001003, "Some User")
				require.NoError(t, err)
				token, err := testTools.GenerateToken(user)
				require.NoError(t, err)
				return "Bearer " + token, -1003001003999
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			name: "user without role",
			setup: func(t *testing.T) (string, int64) {
				owner, err := testTools.CreateUser(ctx, 3001004, "Owner")
				require.NoError(t, err)
				ch, err := testTools.CreateChannel(ctx, -1003001004001, "Private", nil)
				require.NoError(t, err)
				_, err = testTools.CreateChannelRole(
					ctx,
					ch.ID,
					owner.ID,
					entity.ChannelRoleTypeOwner,
				)
				require.NoError(t, err)

				outsider, err := testTools.CreateUser(ctx, 3001005, "Outsider")
				require.NoError(t, err)
				token, err := testTools.GenerateToken(outsider)
				require.NoError(t, err)
				return "Bearer " + token, ch.TgChannelID
			},
			expectedStatus: http.StatusForbidden,
		},
		{
			name: "unauthenticated",
			setup: func(t *testing.T) (string, int64) {
				return "", -1003001006001
			},
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			authHeader, channelID := tt.setup(t)

			var url string
			if channelID == 0 {
				url = testServer.URL + "/api/v1/channels/invalid/admins"
			} else {
				url = fmt.Sprintf("%s/api/v1/channels/%d/admins", testServer.URL, channelID)
			}

			req, err := http.NewRequest(http.MethodGet, url, nil)
			require.NoError(t, err)
			if authHeader != "" {
				req.Header.Set("Authorization", authHeader)
			}

			resp, err := http.DefaultClient.Do(req)
			require.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, tt.expectedStatus, resp.StatusCode)

			if tt.check != nil {
				body, err := io.ReadAll(resp.Body)
				require.NoError(t, err)
				tt.check(t, body)
			}
		})
	}
}

func TestHandleGetChannelManagers(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name           string
		setup          func(t *testing.T) (string, int64)
		expectedStatus int
		check          func(t *testing.T, body []byte)
	}{
		{
			name: "channel with managers",
			setup: func(t *testing.T) (string, int64) {
				owner, err := testTools.CreateUser(ctx, 4001001, "Owner")
				require.NoError(t, err)
				manager1, err := testTools.CreateUser(ctx, 4001002, "Manager One")
				require.NoError(t, err)
				manager2, err := testTools.CreateUser(ctx, 4001003, "Manager Two")
				require.NoError(t, err)

				ch, err := testTools.CreateChannel(ctx, -1004001001001, "Managed Channel", nil)
				require.NoError(t, err)
				_, err = testTools.CreateChannelRole(
					ctx,
					ch.ID,
					owner.ID,
					entity.ChannelRoleTypeOwner,
				)
				require.NoError(t, err)
				_, err = testTools.CreateChannelRole(
					ctx,
					ch.ID,
					manager1.ID,
					entity.ChannelRoleTypeManager,
				)
				require.NoError(t, err)
				_, err = testTools.CreateChannelRole(
					ctx,
					ch.ID,
					manager2.ID,
					entity.ChannelRoleTypeManager,
				)
				require.NoError(t, err)

				token, err := testTools.GenerateToken(owner)
				require.NoError(t, err)
				return "Bearer " + token, ch.TgChannelID
			},
			expectedStatus: http.StatusOK,
			check: func(t *testing.T, body []byte) {
				var resp dto.ChannelManagersResponse
				require.NoError(t, json.Unmarshal(body, &resp))
				assert.Len(t, resp.Managers, 3)
			},
		},
		{
			name: "channel with only owner",
			setup: func(t *testing.T) (string, int64) {
				owner, err := testTools.CreateUser(ctx, 4001004, "Solo Owner")
				require.NoError(t, err)

				ch, err := testTools.CreateChannel(ctx, -1004001004001, "Solo Channel", nil)
				require.NoError(t, err)
				_, err = testTools.CreateChannelRole(
					ctx,
					ch.ID,
					owner.ID,
					entity.ChannelRoleTypeOwner,
				)
				require.NoError(t, err)

				token, err := testTools.GenerateToken(owner)
				require.NoError(t, err)
				return "Bearer " + token, ch.TgChannelID
			},
			expectedStatus: http.StatusOK,
			check: func(t *testing.T, body []byte) {
				var resp dto.ChannelManagersResponse
				require.NoError(t, json.Unmarshal(body, &resp))
				assert.Len(t, resp.Managers, 1)
			},
		},
		{
			name: "invalid channel id",
			setup: func(t *testing.T) (string, int64) {
				user, err := testTools.CreateUser(ctx, 4001005, "Some User")
				require.NoError(t, err)
				token, err := testTools.GenerateToken(user)
				require.NoError(t, err)
				return "Bearer " + token, 0
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "non-existent channel",
			setup: func(t *testing.T) (string, int64) {
				user, err := testTools.CreateUser(ctx, 4001006, "Some User")
				require.NoError(t, err)
				token, err := testTools.GenerateToken(user)
				require.NoError(t, err)
				return "Bearer " + token, -1004001006999
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			name: "user without role",
			setup: func(t *testing.T) (string, int64) {
				owner, err := testTools.CreateUser(ctx, 4001007, "Owner")
				require.NoError(t, err)
				ch, err := testTools.CreateChannel(ctx, -1004001007001, "Private", nil)
				require.NoError(t, err)
				_, err = testTools.CreateChannelRole(
					ctx,
					ch.ID,
					owner.ID,
					entity.ChannelRoleTypeOwner,
				)
				require.NoError(t, err)

				outsider, err := testTools.CreateUser(ctx, 4001008, "Outsider")
				require.NoError(t, err)
				token, err := testTools.GenerateToken(outsider)
				require.NoError(t, err)
				return "Bearer " + token, ch.TgChannelID
			},
			expectedStatus: http.StatusForbidden,
		},
		{
			name: "unauthenticated",
			setup: func(t *testing.T) (string, int64) {
				return "", -1004001009001
			},
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			authHeader, channelID := tt.setup(t)

			var url string
			if channelID == 0 {
				url = testServer.URL + "/api/v1/channels/invalid/managers"
			} else {
				url = fmt.Sprintf("%s/api/v1/channels/%d/managers", testServer.URL, channelID)
			}

			req, err := http.NewRequest(http.MethodGet, url, nil)
			require.NoError(t, err)
			if authHeader != "" {
				req.Header.Set("Authorization", authHeader)
			}

			resp, err := http.DefaultClient.Do(req)
			require.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, tt.expectedStatus, resp.StatusCode)

			if tt.check != nil {
				body, err := io.ReadAll(resp.Body)
				require.NoError(t, err)
				tt.check(t, body)
			}
		})
	}
}

func TestHandleAddManager(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name           string
		setup          func(t *testing.T) (string, int64, string)
		expectedStatus int
		check          func(t *testing.T)
	}{
		{
			name: "owner adds existing user",
			setup: func(t *testing.T) (string, int64, string) {
				owner, err := testTools.CreateUser(ctx, 5001001, "Owner")
				require.NoError(t, err)
				target, err := testTools.CreateUser(ctx, 5001002, "Target User")
				require.NoError(t, err)

				ch, err := testTools.CreateChannel(ctx, -1005001001001, "Add Manager Channel", nil)
				require.NoError(t, err)
				_, err = testTools.CreateChannelRole(
					ctx,
					ch.ID,
					owner.ID,
					entity.ChannelRoleTypeOwner,
				)
				require.NoError(t, err)

				token, err := testTools.GenerateToken(owner)
				require.NoError(t, err)
				body := fmt.Sprintf(`{"telegram_id": %d}`, target.TgID)
				return "Bearer " + token, ch.TgChannelID, body
			},
			expectedStatus: http.StatusNoContent,
			check: func(t *testing.T) {
				ch, err := testTools.GetChannelByTgID(ctx, -1005001001001)
				require.NoError(t, err)
				roles, err := testTools.GetChannelRolesByChannelID(ctx, ch.ID)
				require.NoError(t, err)
				assert.Len(t, roles, 2)
			},
		},
		{
			name: "owner adds non-existent user",
			setup: func(t *testing.T) (string, int64, string) {
				owner, err := testTools.CreateUser(ctx, 5001003, "Owner")
				require.NoError(t, err)

				ch, err := testTools.CreateChannel(ctx, -1005001003001, "Auto Create Channel", nil)
				require.NoError(t, err)
				_, err = testTools.CreateChannelRole(
					ctx,
					ch.ID,
					owner.ID,
					entity.ChannelRoleTypeOwner,
				)
				require.NoError(t, err)

				token, err := testTools.GenerateToken(owner)
				require.NoError(t, err)
				body := `{"telegram_id": 5001004}`
				return "Bearer " + token, ch.TgChannelID, body
			},
			expectedStatus: http.StatusNoContent,
			check: func(t *testing.T) {
				user, err := testTools.GetUserByTgID(ctx, 5001004)
				require.NoError(t, err)
				assert.NotNil(t, user)
			},
		},
		{
			name: "manager tries to add",
			setup: func(t *testing.T) (string, int64, string) {
				owner, err := testTools.CreateUser(ctx, 5001005, "Owner")
				require.NoError(t, err)
				manager, err := testTools.CreateUser(ctx, 5001006, "Manager")
				require.NoError(t, err)

				ch, err := testTools.CreateChannel(ctx, -1005001005001, "Manager Add Channel", nil)
				require.NoError(t, err)
				_, err = testTools.CreateChannelRole(
					ctx,
					ch.ID,
					owner.ID,
					entity.ChannelRoleTypeOwner,
				)
				require.NoError(t, err)
				_, err = testTools.CreateChannelRole(
					ctx,
					ch.ID,
					manager.ID,
					entity.ChannelRoleTypeManager,
				)
				require.NoError(t, err)

				token, err := testTools.GenerateToken(manager)
				require.NoError(t, err)
				body := `{"telegram_id": 5001007}`
				return "Bearer " + token, ch.TgChannelID, body
			},
			expectedStatus: http.StatusForbidden,
		},
		{
			name: "non-existent channel",
			setup: func(t *testing.T) (string, int64, string) {
				user, err := testTools.CreateUser(ctx, 5001008, "Some User")
				require.NoError(t, err)
				token, err := testTools.GenerateToken(user)
				require.NoError(t, err)
				return "Bearer " + token, -1005001008999, `{"telegram_id": 123}`
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			name: "invalid channel id",
			setup: func(t *testing.T) (string, int64, string) {
				user, err := testTools.CreateUser(ctx, 5001009, "Some User")
				require.NoError(t, err)
				token, err := testTools.GenerateToken(user)
				require.NoError(t, err)
				return "Bearer " + token, 0, `{"telegram_id": 123}`
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "missing telegram_id in body",
			setup: func(t *testing.T) (string, int64, string) {
				owner, err := testTools.CreateUser(ctx, 5001010, "Owner")
				require.NoError(t, err)

				ch, err := testTools.CreateChannel(ctx, -1005001010001, "Missing ID Channel", nil)
				require.NoError(t, err)
				_, err = testTools.CreateChannelRole(
					ctx,
					ch.ID,
					owner.ID,
					entity.ChannelRoleTypeOwner,
				)
				require.NoError(t, err)

				token, err := testTools.GenerateToken(owner)
				require.NoError(t, err)
				return "Bearer " + token, ch.TgChannelID, `{}`
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "unauthenticated",
			setup: func(t *testing.T) (string, int64, string) {
				return "", -1005001011001, `{"telegram_id": 123}`
			},
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			authHeader, channelID, reqBody := tt.setup(t)

			var url string
			if channelID == 0 {
				url = testServer.URL + "/api/v1/channels/invalid/managers"
			} else {
				url = fmt.Sprintf("%s/api/v1/channels/%d/managers", testServer.URL, channelID)
			}

			req, err := http.NewRequest(http.MethodPost, url, bytes.NewBufferString(reqBody))
			require.NoError(t, err)
			req.Header.Set("Content-Type", "application/json")
			if authHeader != "" {
				req.Header.Set("Authorization", authHeader)
			}

			resp, err := http.DefaultClient.Do(req)
			require.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, tt.expectedStatus, resp.StatusCode)

			if tt.check != nil {
				tt.check(t)
			}
		})
	}
}

func TestHandleRemoveManager(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name           string
		setup          func(t *testing.T) (string, int64, int64)
		expectedStatus int
		check          func(t *testing.T)
	}{
		{
			name: "owner removes manager",
			setup: func(t *testing.T) (string, int64, int64) {
				owner, err := testTools.CreateUser(ctx, 6001001, "Owner")
				require.NoError(t, err)
				manager, err := testTools.CreateUser(ctx, 6001002, "Manager")
				require.NoError(t, err)

				ch, err := testTools.CreateChannel(
					ctx,
					-1006001001001,
					"Remove Manager Channel",
					nil,
				)
				require.NoError(t, err)
				_, err = testTools.CreateChannelRole(
					ctx,
					ch.ID,
					owner.ID,
					entity.ChannelRoleTypeOwner,
				)
				require.NoError(t, err)
				_, err = testTools.CreateChannelRole(
					ctx,
					ch.ID,
					manager.ID,
					entity.ChannelRoleTypeManager,
				)
				require.NoError(t, err)

				token, err := testTools.GenerateToken(owner)
				require.NoError(t, err)
				return "Bearer " + token, ch.TgChannelID, manager.TgID
			},
			expectedStatus: http.StatusNoContent,
			check: func(t *testing.T) {
				ch, err := testTools.GetChannelByTgID(ctx, -1006001001001)
				require.NoError(t, err)
				roles, err := testTools.GetChannelRolesByChannelID(ctx, ch.ID)
				require.NoError(t, err)
				assert.Len(t, roles, 1)
				assert.Equal(t, entity.ChannelRoleTypeOwner, roles[0].Role)
			},
		},
		{
			name: "owner tries to remove owner",
			setup: func(t *testing.T) (string, int64, int64) {
				owner, err := testTools.CreateUser(ctx, 6001003, "Owner")
				require.NoError(t, err)

				ch, err := testTools.CreateChannel(ctx, -1006001003001, "Self Remove Channel", nil)
				require.NoError(t, err)
				_, err = testTools.CreateChannelRole(
					ctx,
					ch.ID,
					owner.ID,
					entity.ChannelRoleTypeOwner,
				)
				require.NoError(t, err)

				token, err := testTools.GenerateToken(owner)
				require.NoError(t, err)
				return "Bearer " + token, ch.TgChannelID, owner.TgID
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "manager tries to remove",
			setup: func(t *testing.T) (string, int64, int64) {
				owner, err := testTools.CreateUser(ctx, 6001004, "Owner")
				require.NoError(t, err)
				manager1, err := testTools.CreateUser(ctx, 6001005, "Manager1")
				require.NoError(t, err)
				manager2, err := testTools.CreateUser(ctx, 6001006, "Manager2")
				require.NoError(t, err)

				ch, err := testTools.CreateChannel(
					ctx,
					-1006001004001,
					"Manager Remove Channel",
					nil,
				)
				require.NoError(t, err)
				_, err = testTools.CreateChannelRole(
					ctx,
					ch.ID,
					owner.ID,
					entity.ChannelRoleTypeOwner,
				)
				require.NoError(t, err)
				_, err = testTools.CreateChannelRole(
					ctx,
					ch.ID,
					manager1.ID,
					entity.ChannelRoleTypeManager,
				)
				require.NoError(t, err)
				_, err = testTools.CreateChannelRole(
					ctx,
					ch.ID,
					manager2.ID,
					entity.ChannelRoleTypeManager,
				)
				require.NoError(t, err)

				token, err := testTools.GenerateToken(manager1)
				require.NoError(t, err)
				return "Bearer " + token, ch.TgChannelID, manager2.TgID
			},
			expectedStatus: http.StatusForbidden,
		},
		{
			name: "target user not in channel",
			setup: func(t *testing.T) (string, int64, int64) {
				owner, err := testTools.CreateUser(ctx, 6001007, "Owner")
				require.NoError(t, err)

				ch, err := testTools.CreateChannel(ctx, -1006001007001, "Not In Channel", nil)
				require.NoError(t, err)
				_, err = testTools.CreateChannelRole(
					ctx,
					ch.ID,
					owner.ID,
					entity.ChannelRoleTypeOwner,
				)
				require.NoError(t, err)

				token, err := testTools.GenerateToken(owner)
				require.NoError(t, err)
				return "Bearer " + token, ch.TgChannelID, 6001008
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			name: "invalid channel id",
			setup: func(t *testing.T) (string, int64, int64) {
				user, err := testTools.CreateUser(ctx, 6001009, "Some User")
				require.NoError(t, err)
				token, err := testTools.GenerateToken(user)
				require.NoError(t, err)
				return "Bearer " + token, 0, 123
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "invalid target tgID",
			setup: func(t *testing.T) (string, int64, int64) {
				owner, err := testTools.CreateUser(ctx, 6001010, "Owner")
				require.NoError(t, err)

				ch, err := testTools.CreateChannel(ctx, -1006001010001, "Invalid Target", nil)
				require.NoError(t, err)
				_, err = testTools.CreateChannelRole(
					ctx,
					ch.ID,
					owner.ID,
					entity.ChannelRoleTypeOwner,
				)
				require.NoError(t, err)

				token, err := testTools.GenerateToken(owner)
				require.NoError(t, err)
				return "Bearer " + token, ch.TgChannelID, 0
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "unauthenticated",
			setup: func(t *testing.T) (string, int64, int64) {
				return "", -1006001011001, 123
			},
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			authHeader, channelID, targetTgID := tt.setup(t)

			var url string
			if channelID == 0 {
				url = fmt.Sprintf(
					"%s/api/v1/channels/invalid/managers/%d",
					testServer.URL,
					targetTgID,
				)
			} else if targetTgID == 0 {
				url = fmt.Sprintf(
					"%s/api/v1/channels/%d/managers/invalid",
					testServer.URL,
					channelID,
				)
			} else {
				url = fmt.Sprintf(
					"%s/api/v1/channels/%d/managers/%d",
					testServer.URL,
					channelID,
					targetTgID,
				)
			}

			req, err := http.NewRequest(http.MethodDelete, url, nil)
			require.NoError(t, err)
			if authHeader != "" {
				req.Header.Set("Authorization", authHeader)
			}

			resp, err := http.DefaultClient.Do(req)
			require.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, tt.expectedStatus, resp.StatusCode)

			if tt.check != nil {
				tt.check(t)
			}
		})
	}
}
