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
)

func TestHandleMe(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name           string
		setup          func(t *testing.T) string
		expectedStatus int
		check          func(t *testing.T, body []byte)
	}{
		{
			name: "success",
			setup: func(t *testing.T) string {
				user, err := testTools.CreateUser(ctx, 123456789, "John Doe")
				require.NoError(t, err)
				token, err := testTools.GenerateToken(user)
				require.NoError(t, err)
				return "Bearer " + token
			},
			expectedStatus: http.StatusOK,
			check: func(t *testing.T, body []byte) {
				var resp dto.UserResponse
				require.NoError(t, json.Unmarshal(body, &resp))
				assert.Equal(t, int64(123456789), resp.TgID)
				assert.Equal(t, "John Doe", resp.Name)
			},
		},
		{
			name:           "no auth header",
			setup:          func(t *testing.T) string { return "" },
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "invalid format",
			setup:          func(t *testing.T) string { return "InvalidToken" },
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name: "expired token",
			setup: func(t *testing.T) string {
				user, err := testTools.CreateUser(ctx, 987654321, "Jane")
				require.NoError(t, err)
				token, err := testTools.GenerateExpiredToken(user)
				require.NoError(t, err)
				return "Bearer " + token
			},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name: "user deleted",
			setup: func(t *testing.T) string {
				user, err := testTools.CreateUser(ctx, 444555666, "Deleted")
				require.NoError(t, err)
				_, err = testPool.Exec(
					ctx,
					"UPDATE users SET deleted_at = NOW() WHERE id = $1",
					user.ID,
				)
				require.NoError(t, err)
				token, err := testTools.GenerateToken(user)
				require.NoError(t, err)
				return "Bearer " + token
			},
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.NoError(t, testTools.TruncateAll(ctx))

			authHeader := tt.setup(t)

			req, err := http.NewRequest(http.MethodGet, testServer.URL+"/api/v1/me", nil)
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
