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

func TestHandleGetAdFormats(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name           string
		setup          func(t *testing.T) (string, int64)
		expectedStatus int
		check          func(t *testing.T, body []byte)
	}{
		{
			name: "owner gets ad formats",
			setup: func(t *testing.T) (string, int64) {
				owner, err := testTools.CreateUser(ctx, 8001001, "Owner")
				require.NoError(t, err)

				ch, err := testTools.CreateChannel(ctx, -1008001001001, "Ad Format Channel", nil)
				require.NoError(t, err)
				_, err = testTools.CreateChannelRole(
					ctx,
					ch.ID,
					owner.ID,
					entity.ChannelRoleTypeOwner,
				)
				require.NoError(t, err)

				_, err = testTools.CreateAdFormat(
					ctx,
					ch.ID,
					entity.AdFormatTypePost,
					false,
					12,
					2,
					1000000000,
				)
				require.NoError(t, err)
				_, err = testTools.CreateAdFormat(
					ctx,
					ch.ID,
					entity.AdFormatTypePost,
					true,
					24,
					4,
					2000000000,
				)
				require.NoError(t, err)

				token, err := testTools.GenerateToken(owner)
				require.NoError(t, err)
				return "Bearer " + token, ch.TgChannelID
			},
			expectedStatus: http.StatusOK,
			check: func(t *testing.T, body []byte) {
				var resp dto.AdFormatsResponse
				require.NoError(t, json.Unmarshal(body, &resp))
				assert.Len(t, resp.AdFormats, 2)
			},
		},
		{
			name: "manager gets ad formats",
			setup: func(t *testing.T) (string, int64) {
				owner, err := testTools.CreateUser(ctx, 8001002, "Owner")
				require.NoError(t, err)
				manager, err := testTools.CreateUser(ctx, 8001003, "Manager")
				require.NoError(t, err)

				ch, err := testTools.CreateChannel(ctx, -1008001002001, "Manager Channel", nil)
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

				_, err = testTools.CreateAdFormat(
					ctx,
					ch.ID,
					entity.AdFormatTypePost,
					false,
					12,
					2,
					1000000000,
				)
				require.NoError(t, err)

				token, err := testTools.GenerateToken(manager)
				require.NoError(t, err)
				return "Bearer " + token, ch.TgChannelID
			},
			expectedStatus: http.StatusOK,
			check: func(t *testing.T, body []byte) {
				var resp dto.AdFormatsResponse
				require.NoError(t, json.Unmarshal(body, &resp))
				assert.Len(t, resp.AdFormats, 1)
			},
		},
		{
			name: "empty ad formats",
			setup: func(t *testing.T) (string, int64) {
				owner, err := testTools.CreateUser(ctx, 8001004, "Owner")
				require.NoError(t, err)

				ch, err := testTools.CreateChannel(ctx, -1008001004001, "Empty Channel", nil)
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
				var resp dto.AdFormatsResponse
				require.NoError(t, json.Unmarshal(body, &resp))
				assert.Empty(t, resp.AdFormats)
			},
		},
		{
			name: "non-member forbidden",
			setup: func(t *testing.T) (string, int64) {
				owner, err := testTools.CreateUser(ctx, 8001005, "Owner")
				require.NoError(t, err)
				outsider, err := testTools.CreateUser(ctx, 8001006, "Outsider")
				require.NoError(t, err)

				ch, err := testTools.CreateChannel(ctx, -1008001005001, "Private Channel", nil)
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
				return "Bearer " + token, ch.TgChannelID
			},
			expectedStatus: http.StatusForbidden,
		},
		{
			name: "unauthenticated",
			setup: func(t *testing.T) (string, int64) {
				return "", -1008001007001
			},
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			authHeader, channelID := tt.setup(t)

			url := fmt.Sprintf("%s/api/v1/channels/%d/ad-formats", testServer.URL, channelID)
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

func TestHandleAddAdFormat(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name           string
		setup          func(t *testing.T) (string, int64, string)
		expectedStatus int
		check          func(t *testing.T)
	}{
		{
			name: "owner adds ad format",
			setup: func(t *testing.T) (string, int64, string) {
				owner, err := testTools.CreateUser(ctx, 8002001, "Owner")
				require.NoError(t, err)

				ch, err := testTools.CreateChannel(ctx, -1008002001001, "Add Format Channel", nil)
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
				body := `{"format_type": "post", "is_native": false, "feed_hours": 12, "top_hours": 2, "price_nano_ton": 1000000000}`
				return "Bearer " + token, ch.TgChannelID, body
			},
			expectedStatus: http.StatusNoContent,
			check: func(t *testing.T) {
				ch, err := testTools.GetChannelByTgID(ctx, -1008002001001)
				require.NoError(t, err)
				formats, err := testTools.GetAdFormatsByChannelID(ctx, ch.ID)
				require.NoError(t, err)
				assert.Len(t, formats, 1)
				assert.Equal(t, entity.AdFormatTypePost, formats[0].FormatType)
				assert.False(t, formats[0].IsNative)
				assert.Equal(t, 12, formats[0].FeedHours)
				assert.Equal(t, 2, formats[0].TopHours)
				assert.Equal(t, int64(1000000000), formats[0].PriceNanoTON)
			},
		},
		{
			name: "owner adds native ad format",
			setup: func(t *testing.T) (string, int64, string) {
				owner, err := testTools.CreateUser(ctx, 8002002, "Owner")
				require.NoError(t, err)

				ch, err := testTools.CreateChannel(
					ctx,
					-1008002002001,
					"Native Format Channel",
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

				token, err := testTools.GenerateToken(owner)
				require.NoError(t, err)
				body := `{"format_type": "post", "is_native": true, "feed_hours": 24, "top_hours": 4, "price_nano_ton": 2000000000}`
				return "Bearer " + token, ch.TgChannelID, body
			},
			expectedStatus: http.StatusNoContent,
			check: func(t *testing.T) {
				ch, err := testTools.GetChannelByTgID(ctx, -1008002002001)
				require.NoError(t, err)
				formats, err := testTools.GetAdFormatsByChannelID(ctx, ch.ID)
				require.NoError(t, err)
				assert.Len(t, formats, 1)
				assert.True(t, formats[0].IsNative)
			},
		},
		{
			name: "duplicate ad format conflict",
			setup: func(t *testing.T) (string, int64, string) {
				owner, err := testTools.CreateUser(ctx, 8002003, "Owner")
				require.NoError(t, err)

				ch, err := testTools.CreateChannel(ctx, -1008002003001, "Duplicate Channel", nil)
				require.NoError(t, err)
				_, err = testTools.CreateChannelRole(
					ctx,
					ch.ID,
					owner.ID,
					entity.ChannelRoleTypeOwner,
				)
				require.NoError(t, err)
				_, err = testTools.CreateAdFormat(
					ctx,
					ch.ID,
					entity.AdFormatTypePost,
					false,
					12,
					2,
					1000000000,
				)
				require.NoError(t, err)

				token, err := testTools.GenerateToken(owner)
				require.NoError(t, err)
				body := `{"format_type": "post", "is_native": false, "feed_hours": 12, "top_hours": 2, "price_nano_ton": 500000000}`
				return "Bearer " + token, ch.TgChannelID, body
			},
			expectedStatus: http.StatusConflict,
		},
		{
			name: "repost format type not allowed",
			setup: func(t *testing.T) (string, int64, string) {
				owner, err := testTools.CreateUser(ctx, 8002004, "Owner")
				require.NoError(t, err)

				ch, err := testTools.CreateChannel(ctx, -1008002004001, "Repost Channel", nil)
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
				body := `{"format_type": "repost", "is_native": false, "feed_hours": 12, "top_hours": 2, "price_nano_ton": 1000000000}`
				return "Bearer " + token, ch.TgChannelID, body
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "story format type not allowed",
			setup: func(t *testing.T) (string, int64, string) {
				owner, err := testTools.CreateUser(ctx, 8002005, "Owner")
				require.NoError(t, err)

				ch, err := testTools.CreateChannel(ctx, -1008002005001, "Story Channel", nil)
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
				body := `{"format_type": "story", "is_native": false, "feed_hours": 12, "top_hours": 2, "price_nano_ton": 1000000000}`
				return "Bearer " + token, ch.TgChannelID, body
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "manager tries to add format",
			setup: func(t *testing.T) (string, int64, string) {
				owner, err := testTools.CreateUser(ctx, 8002006, "Owner")
				require.NoError(t, err)
				manager, err := testTools.CreateUser(ctx, 8002007, "Manager")
				require.NoError(t, err)

				ch, err := testTools.CreateChannel(ctx, -1008002006001, "Manager Channel", nil)
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
				body := `{"format_type": "post", "is_native": false, "feed_hours": 12, "top_hours": 2, "price_nano_ton": 1000000000}`
				return "Bearer " + token, ch.TgChannelID, body
			},
			expectedStatus: http.StatusForbidden,
		},
		{
			name: "unauthenticated",
			setup: func(t *testing.T) (string, int64, string) {
				return "", -1008002008001, `{"format_type": "post", "is_native": false, "feed_hours": 12, "top_hours": 2, "price_nano_ton": 1000000000}`
			},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name: "invalid feed_hours value",
			setup: func(t *testing.T) (string, int64, string) {
				owner, err := testTools.CreateUser(ctx, 8002009, "Owner")
				require.NoError(t, err)

				ch, err := testTools.CreateChannel(ctx, -1008002009001, "Invalid Feed Channel", nil)
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
				body := `{"format_type": "post", "is_native": false, "feed_hours": 10, "top_hours": 2, "price_nano_ton": 1000000000}`
				return "Bearer " + token, ch.TgChannelID, body
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "invalid top_hours value",
			setup: func(t *testing.T) (string, int64, string) {
				owner, err := testTools.CreateUser(ctx, 8002010, "Owner")
				require.NoError(t, err)

				ch, err := testTools.CreateChannel(ctx, -1008002010001, "Invalid Top Channel", nil)
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
				body := `{"format_type": "post", "is_native": false, "feed_hours": 12, "top_hours": 3, "price_nano_ton": 1000000000}`
				return "Bearer " + token, ch.TgChannelID, body
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "zero price_nano_ton",
			setup: func(t *testing.T) (string, int64, string) {
				owner, err := testTools.CreateUser(ctx, 8002011, "Owner")
				require.NoError(t, err)

				ch, err := testTools.CreateChannel(ctx, -1008002011001, "Zero Price Channel", nil)
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
				body := `{"format_type": "post", "is_native": false, "feed_hours": 12, "top_hours": 2, "price_nano_ton": 0}`
				return "Bearer " + token, ch.TgChannelID, body
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "negative price_nano_ton",
			setup: func(t *testing.T) (string, int64, string) {
				owner, err := testTools.CreateUser(ctx, 8002012, "Owner")
				require.NoError(t, err)

				ch, err := testTools.CreateChannel(
					ctx,
					-1008002012001,
					"Negative Price Channel",
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

				token, err := testTools.GenerateToken(owner)
				require.NoError(t, err)
				body := `{"format_type": "post", "is_native": false, "feed_hours": 12, "top_hours": 2, "price_nano_ton": -100}`
				return "Bearer " + token, ch.TgChannelID, body
			},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			authHeader, channelID, reqBody := tt.setup(t)

			url := fmt.Sprintf("%s/api/v1/channels/%d/ad-formats", testServer.URL, channelID)
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

func TestHandleRemoveAdFormat(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name           string
		setup          func(t *testing.T) (string, int64, string)
		expectedStatus int
		check          func(t *testing.T)
	}{
		{
			name: "owner removes ad format",
			setup: func(t *testing.T) (string, int64, string) {
				owner, err := testTools.CreateUser(ctx, 8003001, "Owner")
				require.NoError(t, err)

				ch, err := testTools.CreateChannel(
					ctx,
					-1008003001001,
					"Remove Format Channel",
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

				af, err := testTools.CreateAdFormat(
					ctx,
					ch.ID,
					entity.AdFormatTypePost,
					false,
					12,
					2,
					1000000000,
				)
				require.NoError(t, err)

				token, err := testTools.GenerateToken(owner)
				require.NoError(t, err)
				return "Bearer " + token, ch.TgChannelID, af.ID.String()
			},
			expectedStatus: http.StatusNoContent,
			check: func(t *testing.T) {
				ch, err := testTools.GetChannelByTgID(ctx, -1008003001001)
				require.NoError(t, err)
				formats, err := testTools.GetAdFormatsByChannelID(ctx, ch.ID)
				require.NoError(t, err)
				assert.Empty(t, formats)
			},
		},
		{
			name: "owner removes format from different channel",
			setup: func(t *testing.T) (string, int64, string) {
				owner1, err := testTools.CreateUser(ctx, 8003002, "Owner1")
				require.NoError(t, err)
				owner2, err := testTools.CreateUser(ctx, 8003003, "Owner2")
				require.NoError(t, err)

				ch1, err := testTools.CreateChannel(ctx, -1008003002001, "Channel1", nil)
				require.NoError(t, err)
				_, err = testTools.CreateChannelRole(
					ctx,
					ch1.ID,
					owner1.ID,
					entity.ChannelRoleTypeOwner,
				)
				require.NoError(t, err)

				ch2, err := testTools.CreateChannel(ctx, -1008003002002, "Channel2", nil)
				require.NoError(t, err)
				_, err = testTools.CreateChannelRole(
					ctx,
					ch2.ID,
					owner2.ID,
					entity.ChannelRoleTypeOwner,
				)
				require.NoError(t, err)

				af, err := testTools.CreateAdFormat(
					ctx,
					ch2.ID,
					entity.AdFormatTypePost,
					false,
					12,
					2,
					1000000000,
				)
				require.NoError(t, err)

				token, err := testTools.GenerateToken(owner1)
				require.NoError(t, err)
				return "Bearer " + token, ch1.TgChannelID, af.ID.String()
			},
			expectedStatus: http.StatusForbidden,
		},
		{
			name: "manager tries to remove format",
			setup: func(t *testing.T) (string, int64, string) {
				owner, err := testTools.CreateUser(ctx, 8003004, "Owner")
				require.NoError(t, err)
				manager, err := testTools.CreateUser(ctx, 8003005, "Manager")
				require.NoError(t, err)

				ch, err := testTools.CreateChannel(ctx, -1008003004001, "Manager Channel", nil)
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

				af, err := testTools.CreateAdFormat(
					ctx,
					ch.ID,
					entity.AdFormatTypePost,
					false,
					12,
					2,
					1000000000,
				)
				require.NoError(t, err)

				token, err := testTools.GenerateToken(manager)
				require.NoError(t, err)
				return "Bearer " + token, ch.TgChannelID, af.ID.String()
			},
			expectedStatus: http.StatusForbidden,
		},
		{
			name: "non-existent format",
			setup: func(t *testing.T) (string, int64, string) {
				owner, err := testTools.CreateUser(ctx, 8003006, "Owner")
				require.NoError(t, err)

				ch, err := testTools.CreateChannel(ctx, -1008003006001, "Empty Channel", nil)
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
				return "Bearer " + token, ch.TgChannelID, "01942f8a-0000-7000-0000-000000000000"
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			name: "invalid format id",
			setup: func(t *testing.T) (string, int64, string) {
				owner, err := testTools.CreateUser(ctx, 8003007, "Owner")
				require.NoError(t, err)

				ch, err := testTools.CreateChannel(ctx, -1008003007001, "Invalid ID Channel", nil)
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
				return "Bearer " + token, ch.TgChannelID, "invalid-uuid"
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "unauthenticated",
			setup: func(t *testing.T) (string, int64, string) {
				return "", -1008003008001, "01942f8a-0000-7000-0000-000000000000"
			},
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			authHeader, channelID, formatID := tt.setup(t)

			url := fmt.Sprintf(
				"%s/api/v1/channels/%d/ad-formats/%s",
				testServer.URL,
				channelID,
				formatID,
			)
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
