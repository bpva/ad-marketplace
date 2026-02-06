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

	"github.com/bpva/ad-marketplace/internal/entity"
)

func TestHandleUpdateListing(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name           string
		setup          func(t *testing.T) (string, int64, string)
		expectedStatus int
		check          func(t *testing.T)
	}{
		{
			name: "owner unlists channel",
			setup: func(t *testing.T) (string, int64, string) {
				owner, err := testTools.CreateUser(ctx, 7001001, "Owner")
				require.NoError(t, err)

				ch, err := testTools.CreateChannel(ctx, -1007001001001, "Listed Channel", nil)
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
				return "Bearer " + token, ch.TgChannelID, `{"is_listed": false}`
			},
			expectedStatus: http.StatusNoContent,
			check: func(t *testing.T) {
				ch, err := testTools.GetChannelByTgID(ctx, -1007001001001)
				require.NoError(t, err)
				assert.False(t, ch.IsListed)
			},
		},
		{
			name: "owner lists channel",
			setup: func(t *testing.T) (string, int64, string) {
				owner, err := testTools.CreateUser(ctx, 7001002, "Owner")
				require.NoError(t, err)

				ch, err := testTools.CreateChannel(ctx, -1007001002001, "Unlisted Channel", nil)
				require.NoError(t, err)
				err = testTools.UpdateChannelListing(ctx, ch.ID, false)
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
				return "Bearer " + token, ch.TgChannelID, `{"is_listed": true}`
			},
			expectedStatus: http.StatusNoContent,
			check: func(t *testing.T) {
				ch, err := testTools.GetChannelByTgID(ctx, -1007001002001)
				require.NoError(t, err)
				assert.True(t, ch.IsListed)
			},
		},
		{
			name: "manager tries to update listing",
			setup: func(t *testing.T) (string, int64, string) {
				owner, err := testTools.CreateUser(ctx, 7001003, "Owner")
				require.NoError(t, err)
				manager, err := testTools.CreateUser(ctx, 7001004, "Manager")
				require.NoError(t, err)

				ch, err := testTools.CreateChannel(ctx, -1007001003001, "Manager Channel", nil)
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
				return "Bearer " + token, ch.TgChannelID, `{"is_listed": false}`
			},
			expectedStatus: http.StatusForbidden,
		},
		{
			name: "non-member tries to update listing",
			setup: func(t *testing.T) (string, int64, string) {
				owner, err := testTools.CreateUser(ctx, 7001005, "Owner")
				require.NoError(t, err)
				outsider, err := testTools.CreateUser(ctx, 7001006, "Outsider")
				require.NoError(t, err)

				ch, err := testTools.CreateChannel(ctx, -1007001005001, "Private Channel", nil)
				require.NoError(t, err)
				_, err = testTools.CreateChannelRole(
					ctx,
					ch.ID,
					owner.ID,
					entity.ChannelRoleTypeOwner,
				)
				require.NoError(t, err)

				token, err := testTools.GenerateToken(outsider)
				require.NoError(t, err)
				return "Bearer " + token, ch.TgChannelID, `{"is_listed": false}`
			},
			expectedStatus: http.StatusForbidden,
		},
		{
			name: "non-existent channel",
			setup: func(t *testing.T) (string, int64, string) {
				user, err := testTools.CreateUser(ctx, 7001007, "User")
				require.NoError(t, err)
				token, err := testTools.GenerateToken(user)
				require.NoError(t, err)
				return "Bearer " + token, -1007001007999, `{"is_listed": false}`
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			name: "invalid channel id",
			setup: func(t *testing.T) (string, int64, string) {
				user, err := testTools.CreateUser(ctx, 7001008, "User")
				require.NoError(t, err)
				token, err := testTools.GenerateToken(user)
				require.NoError(t, err)
				return "Bearer " + token, 0, `{"is_listed": false}`
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "unauthenticated",
			setup: func(t *testing.T) (string, int64, string) {
				return "", -1007001009001, `{"is_listed": false}`
			},
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			authHeader, channelID, reqBody := tt.setup(t)

			var url string
			if channelID == 0 {
				url = testServer.URL + "/api/v1/channels/invalid/listing"
			} else {
				url = fmt.Sprintf("%s/api/v1/channels/%d/listing", testServer.URL, channelID)
			}

			req, err := http.NewRequest(http.MethodPatch, url, bytes.NewBufferString(reqBody))
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

func TestChannelResponseIncludesIsListed(t *testing.T) {
	ctx := context.Background()

	owner, err := testTools.CreateUser(ctx, 7002001, "Owner")
	require.NoError(t, err)

	ch, err := testTools.CreateChannel(ctx, -1007002001001, "Test Channel", nil)
	require.NoError(t, err)
	_, err = testTools.CreateChannelRole(ctx, ch.ID, owner.ID, entity.ChannelRoleTypeOwner)
	require.NoError(t, err)

	token, err := testTools.GenerateToken(owner)
	require.NoError(t, err)

	url := fmt.Sprintf("%s/api/v1/channels/%d", testServer.URL, ch.TgChannelID)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	var result map[string]any
	require.NoError(t, json.Unmarshal(body, &result))

	assert.Contains(t, result, "is_listed")
	assert.True(t, result["is_listed"].(bool))
}
