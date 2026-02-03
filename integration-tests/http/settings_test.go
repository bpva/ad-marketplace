//go:build integration

package http_test

import (
	"bytes"
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHandleUpdateSettings(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name           string
		body           string
		expectedStatus int
	}{
		{
			name:           "valid language en",
			body:           `{"language":"en"}`,
			expectedStatus: http.StatusNoContent,
		},
		{
			name:           "valid language ru",
			body:           `{"language":"ru"}`,
			expectedStatus: http.StatusNoContent,
		},
		{
			name:           "invalid language",
			body:           `{"language":"xx"}`,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "valid preferred_mode publisher",
			body:           `{"preferred_mode":"publisher"}`,
			expectedStatus: http.StatusNoContent,
		},
		{
			name:           "valid preferred_mode advertiser",
			body:           `{"preferred_mode":"advertiser"}`,
			expectedStatus: http.StatusNoContent,
		},
		{
			name:           "invalid preferred_mode",
			body:           `{"preferred_mode":"invalid"}`,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "valid theme light",
			body:           `{"theme":"light"}`,
			expectedStatus: http.StatusNoContent,
		},
		{
			name:           "valid theme dark",
			body:           `{"theme":"dark"}`,
			expectedStatus: http.StatusNoContent,
		},
		{
			name:           "valid theme auto",
			body:           `{"theme":"auto"}`,
			expectedStatus: http.StatusNoContent,
		},
		{
			name:           "invalid theme",
			body:           `{"theme":"invalid"}`,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "empty body",
			body:           `{}`,
			expectedStatus: http.StatusNoContent,
		},
		{
			name:           "multiple valid fields",
			body:           `{"language":"ru","theme":"dark","preferred_mode":"advertiser"}`,
			expectedStatus: http.StatusNoContent,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.NoError(t, testTools.TruncateAll(ctx))

			user, err := testTools.CreateUser(ctx, 123456789, "Test User")
			require.NoError(t, err)
			token, err := testTools.GenerateToken(user)
			require.NoError(t, err)

			req, err := http.NewRequest(
				http.MethodPatch,
				testServer.URL+"/api/v1/user/settings",
				bytes.NewBufferString(tt.body),
			)
			require.NoError(t, err)
			req.Header.Set("Authorization", "Bearer "+token)
			req.Header.Set("Content-Type", "application/json")

			resp, err := http.DefaultClient.Do(req)
			require.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, tt.expectedStatus, resp.StatusCode)
		})
	}
}
