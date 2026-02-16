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
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/bpva/ad-marketplace/internal/dto"
	"github.com/bpva/ad-marketplace/internal/entity"
)

type dealSetup struct {
	publisher    *entity.User
	advertiser   *entity.User
	channel      *entity.Channel
	adFormat     *entity.ChannelAdFormat
	templatePost *entity.Post
	advToken     string
	pubToken     string
}

func setupDeal(t *testing.T, ctx context.Context) *dealSetup {
	t.Helper()
	require.NoError(t, testTools.TruncateAll(ctx))

	publisher, err := testTools.CreateUser(ctx, 4001001, "Publisher")
	require.NoError(t, err)
	require.NoError(t, testTools.SetWalletAddress(ctx, publisher.ID, "UQPublisher000"))

	advertiser, err := testTools.CreateUser(ctx, 4001002, "Advertiser")
	require.NoError(t, err)
	require.NoError(t, testTools.SetWalletAddress(ctx, advertiser.ID, "UQAdvertiser000"))

	channel, err := testTools.CreateChannel(ctx, -1004001001, "Deal Channel", nil)
	require.NoError(t, err)
	require.NoError(t, testTools.UpdateChannelListing(ctx, channel.ID, true))

	_, err = testTools.CreateChannelRole(ctx, channel.ID, publisher.ID, entity.ChannelRoleTypeOwner)
	require.NoError(t, err)

	af, err := testTools.CreateAdFormat(
		ctx,
		channel.ID,
		entity.AdFormatTypePost,
		false,
		24,
		4,
		1000000000,
	)
	require.NoError(t, err)

	text := "Ad creative text"
	tmpl, err := testTools.CreatePost(
		ctx,
		entity.PostTypeTemplate,
		advertiser.ID,
		nil,
		nil,
		&text,
		nil,
		nil,
		nil,
	)
	require.NoError(t, err)

	advToken, err := testTools.GenerateToken(advertiser)
	require.NoError(t, err)
	pubToken, err := testTools.GenerateToken(publisher)
	require.NoError(t, err)

	return &dealSetup{
		publisher:    publisher,
		advertiser:   advertiser,
		channel:      channel,
		adFormat:     af,
		templatePost: tmpl,
		advToken:     "Bearer " + advToken,
		pubToken:     "Bearer " + pubToken,
	}
}

func TestHandleCreateDeal(t *testing.T) {
	ctx := context.Background()

	t.Run("happy path", func(t *testing.T) {
		s := setupDeal(t, ctx)

		body, _ := json.Marshal(dto.CreateDealRequest{
			TgChannelID:    s.channel.TgChannelID,
			FormatType:     entity.AdFormatTypePost,
			FeedHours:      24,
			TopHours:       4,
			PriceNanoTON:   1000000000,
			TemplatePostID: s.templatePost.ID.String(),
			ScheduledAt:    time.Now().Add(48 * time.Hour),
		})

		req, err := http.NewRequest(
			http.MethodPost,
			testServer.URL+"/api/v1/deals",
			bytes.NewReader(body),
		)
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", s.advToken)

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusCreated, resp.StatusCode)

		respBody, err := io.ReadAll(resp.Body)
		require.NoError(t, err)

		var dealResp dto.DealResponse
		require.NoError(t, json.Unmarshal(respBody, &dealResp))
		assert.NotEmpty(t, dealResp.ID)
		assert.Equal(t, s.channel.TgChannelID, dealResp.TgChannelID)
		assert.Equal(t, entity.DealStatusPendingPayment, dealResp.Status)
		assert.Equal(t, entity.AdFormatTypePost, dealResp.FormatType)
		assert.Equal(t, int64(1000000000), dealResp.PriceNanoTON)
		assert.NotNil(t, dealResp.Ad)
		assert.Equal(t, "Ad creative text", *dealResp.Ad.Text)
	})

	t.Run("price mismatch", func(t *testing.T) {
		s := setupDeal(t, ctx)

		body, _ := json.Marshal(dto.CreateDealRequest{
			TgChannelID:    s.channel.TgChannelID,
			FormatType:     entity.AdFormatTypePost,
			FeedHours:      24,
			TopHours:       4,
			PriceNanoTON:   999,
			TemplatePostID: s.templatePost.ID.String(),
			ScheduledAt:    time.Now().Add(48 * time.Hour),
		})

		req, err := http.NewRequest(
			http.MethodPost,
			testServer.URL+"/api/v1/deals",
			bytes.NewReader(body),
		)
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", s.advToken)

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusConflict, resp.StatusCode)
	})

	t.Run("past scheduled_at", func(t *testing.T) {
		s := setupDeal(t, ctx)

		body, _ := json.Marshal(dto.CreateDealRequest{
			TgChannelID:    s.channel.TgChannelID,
			FormatType:     entity.AdFormatTypePost,
			FeedHours:      24,
			TopHours:       4,
			PriceNanoTON:   1000000000,
			TemplatePostID: s.templatePost.ID.String(),
			ScheduledAt:    time.Now().Add(-time.Hour),
		})

		req, err := http.NewRequest(
			http.MethodPost,
			testServer.URL+"/api/v1/deals",
			bytes.NewReader(body),
		)
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", s.advToken)

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("template not owned", func(t *testing.T) {
		s := setupDeal(t, ctx)

		otherUser, err := testTools.CreateUser(ctx, 4001099, "Other")
		require.NoError(t, err)
		otherToken, err := testTools.GenerateToken(otherUser)
		require.NoError(t, err)

		body, _ := json.Marshal(dto.CreateDealRequest{
			TgChannelID:    s.channel.TgChannelID,
			FormatType:     entity.AdFormatTypePost,
			FeedHours:      24,
			TopHours:       4,
			PriceNanoTON:   1000000000,
			TemplatePostID: s.templatePost.ID.String(),
			ScheduledAt:    time.Now().Add(48 * time.Hour),
		})

		req, err := http.NewRequest(
			http.MethodPost,
			testServer.URL+"/api/v1/deals",
			bytes.NewReader(body),
		)
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+otherToken)

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusForbidden, resp.StatusCode)
	})

	t.Run("unauthorized", func(t *testing.T) {
		body, _ := json.Marshal(dto.CreateDealRequest{
			TgChannelID:    -1004001001,
			FormatType:     entity.AdFormatTypePost,
			FeedHours:      24,
			TopHours:       4,
			PriceNanoTON:   1000000000,
			TemplatePostID: uuid.New().String(),
			ScheduledAt:    time.Now().Add(48 * time.Hour),
		})

		req, err := http.NewRequest(
			http.MethodPost,
			testServer.URL+"/api/v1/deals",
			bytes.NewReader(body),
		)
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})
}

func TestHandleListDeals(t *testing.T) {
	ctx := context.Background()

	t.Run("advertiser deals", func(t *testing.T) {
		s := setupDeal(t, ctx)

		_, err := testTools.CreateDeal(ctx, s.channel.ID, s.advertiser.ID,
			entity.DealStatusPendingPayment, time.Now().Add(48*time.Hour),
			entity.AdFormatTypePost, false, 24, 4, 1000000000)
		require.NoError(t, err)

		req, err := http.NewRequest(
			http.MethodGet,
			testServer.URL+"/api/v1/deals?role=advertiser",
			nil,
		)
		require.NoError(t, err)
		req.Header.Set("Authorization", s.advToken)

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		respBody, err := io.ReadAll(resp.Body)
		require.NoError(t, err)

		var dealsResp dto.DealsResponse
		require.NoError(t, json.Unmarshal(respBody, &dealsResp))
		assert.Equal(t, 1, dealsResp.Total)
		require.Len(t, dealsResp.Deals, 1)
		assert.Equal(t, s.channel.TgChannelID, dealsResp.Deals[0].TgChannelID)
	})

	t.Run("publisher deals", func(t *testing.T) {
		s := setupDeal(t, ctx)

		_, err := testTools.CreateDeal(ctx, s.channel.ID, s.advertiser.ID,
			entity.DealStatusPendingReview, time.Now().Add(48*time.Hour),
			entity.AdFormatTypePost, false, 24, 4, 1000000000)
		require.NoError(t, err)

		url := fmt.Sprintf(
			"%s/api/v1/deals?role=publisher&channel_id=%d",
			testServer.URL,
			s.channel.TgChannelID,
		)
		req, err := http.NewRequest(http.MethodGet, url, nil)
		require.NoError(t, err)
		req.Header.Set("Authorization", s.pubToken)

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		respBody, err := io.ReadAll(resp.Body)
		require.NoError(t, err)

		var dealsResp dto.DealsResponse
		require.NoError(t, json.Unmarshal(respBody, &dealsResp))
		assert.Equal(t, 1, dealsResp.Total)
		require.Len(t, dealsResp.Deals, 1)
	})

	t.Run("empty results", func(t *testing.T) {
		s := setupDeal(t, ctx)

		req, err := http.NewRequest(http.MethodGet, testServer.URL+"/api/v1/deals", nil)
		require.NoError(t, err)
		req.Header.Set("Authorization", s.advToken)

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		respBody, err := io.ReadAll(resp.Body)
		require.NoError(t, err)

		var dealsResp dto.DealsResponse
		require.NoError(t, json.Unmarshal(respBody, &dealsResp))
		assert.Equal(t, 0, dealsResp.Total)
		assert.Empty(t, dealsResp.Deals)
	})
}

func TestHandleGetDeal(t *testing.T) {
	ctx := context.Background()

	t.Run("as advertiser", func(t *testing.T) {
		s := setupDeal(t, ctx)

		deal, err := testTools.CreateDeal(ctx, s.channel.ID, s.advertiser.ID,
			entity.DealStatusPendingReview, time.Now().Add(48*time.Hour),
			entity.AdFormatTypePost, false, 24, 4, 1000000000)
		require.NoError(t, err)

		req, err := http.NewRequest(
			http.MethodGet,
			testServer.URL+"/api/v1/deals/"+deal.ID.String(),
			nil,
		)
		require.NoError(t, err)
		req.Header.Set("Authorization", s.advToken)

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		respBody, err := io.ReadAll(resp.Body)
		require.NoError(t, err)

		var dealResp dto.DealResponse
		require.NoError(t, json.Unmarshal(respBody, &dealResp))
		assert.Equal(t, deal.ID.String(), dealResp.ID)
		assert.Equal(t, s.channel.TgChannelID, dealResp.TgChannelID)
		assert.Equal(t, entity.DealStatusPendingReview, dealResp.Status)
	})

	t.Run("as publisher", func(t *testing.T) {
		s := setupDeal(t, ctx)

		deal, err := testTools.CreateDeal(ctx, s.channel.ID, s.advertiser.ID,
			entity.DealStatusPendingReview, time.Now().Add(48*time.Hour),
			entity.AdFormatTypePost, false, 24, 4, 1000000000)
		require.NoError(t, err)

		req, err := http.NewRequest(
			http.MethodGet,
			testServer.URL+"/api/v1/deals/"+deal.ID.String(),
			nil,
		)
		require.NoError(t, err)
		req.Header.Set("Authorization", s.pubToken)

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("forbidden for outsider", func(t *testing.T) {
		s := setupDeal(t, ctx)

		deal, err := testTools.CreateDeal(ctx, s.channel.ID, s.advertiser.ID,
			entity.DealStatusPendingReview, time.Now().Add(48*time.Hour),
			entity.AdFormatTypePost, false, 24, 4, 1000000000)
		require.NoError(t, err)

		outsider, err := testTools.CreateUser(ctx, 4001050, "Outsider")
		require.NoError(t, err)
		outsiderToken, err := testTools.GenerateToken(outsider)
		require.NoError(t, err)

		req, err := http.NewRequest(
			http.MethodGet,
			testServer.URL+"/api/v1/deals/"+deal.ID.String(),
			nil,
		)
		require.NoError(t, err)
		req.Header.Set("Authorization", "Bearer "+outsiderToken)

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusForbidden, resp.StatusCode)
	})

	t.Run("not found", func(t *testing.T) {
		s := setupDeal(t, ctx)

		req, err := http.NewRequest(
			http.MethodGet,
			testServer.URL+"/api/v1/deals/"+uuid.New().String(),
			nil,
		)
		require.NoError(t, err)
		req.Header.Set("Authorization", s.advToken)

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	})
}

func TestHandleApproveDeal(t *testing.T) {
	ctx := context.Background()

	t.Run("happy path", func(t *testing.T) {
		s := setupDeal(t, ctx)

		deal, err := testTools.CreateDeal(ctx, s.channel.ID, s.advertiser.ID,
			entity.DealStatusPendingReview, time.Now().Add(48*time.Hour),
			entity.AdFormatTypePost, false, 24, 4, 1000000000)
		require.NoError(t, err)

		req, err := http.NewRequest(
			http.MethodPost,
			testServer.URL+"/api/v1/deals/"+deal.ID.String()+"/approve",
			nil,
		)
		require.NoError(t, err)
		req.Header.Set("Authorization", s.pubToken)

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusNoContent, resp.StatusCode)
	})

	t.Run("invalid transition", func(t *testing.T) {
		s := setupDeal(t, ctx)

		deal, err := testTools.CreateDeal(ctx, s.channel.ID, s.advertiser.ID,
			entity.DealStatusPendingPayment, time.Now().Add(48*time.Hour),
			entity.AdFormatTypePost, false, 24, 4, 1000000000)
		require.NoError(t, err)

		req, err := http.NewRequest(
			http.MethodPost,
			testServer.URL+"/api/v1/deals/"+deal.ID.String()+"/approve",
			nil,
		)
		require.NoError(t, err)
		req.Header.Set("Authorization", s.pubToken)

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("forbidden for advertiser", func(t *testing.T) {
		s := setupDeal(t, ctx)

		deal, err := testTools.CreateDeal(ctx, s.channel.ID, s.advertiser.ID,
			entity.DealStatusPendingReview, time.Now().Add(48*time.Hour),
			entity.AdFormatTypePost, false, 24, 4, 1000000000)
		require.NoError(t, err)

		req, err := http.NewRequest(
			http.MethodPost,
			testServer.URL+"/api/v1/deals/"+deal.ID.String()+"/approve",
			nil,
		)
		require.NoError(t, err)
		req.Header.Set("Authorization", s.advToken)

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusForbidden, resp.StatusCode)
	})
}

func TestHandleRejectDeal(t *testing.T) {
	ctx := context.Background()

	t.Run("with reason", func(t *testing.T) {
		s := setupDeal(t, ctx)

		deal, err := testTools.CreateDeal(ctx, s.channel.ID, s.advertiser.ID,
			entity.DealStatusPendingReview, time.Now().Add(48*time.Hour),
			entity.AdFormatTypePost, false, 24, 4, 1000000000)
		require.NoError(t, err)

		body, _ := json.Marshal(dto.RejectRequest{Reason: strPtr("bad quality")})
		req, err := http.NewRequest(
			http.MethodPost,
			testServer.URL+"/api/v1/deals/"+deal.ID.String()+"/reject",
			bytes.NewReader(body),
		)
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", s.pubToken)

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusNoContent, resp.StatusCode)
	})

	t.Run("without reason", func(t *testing.T) {
		s := setupDeal(t, ctx)

		deal, err := testTools.CreateDeal(ctx, s.channel.ID, s.advertiser.ID,
			entity.DealStatusPendingReview, time.Now().Add(48*time.Hour),
			entity.AdFormatTypePost, false, 24, 4, 1000000000)
		require.NoError(t, err)

		req, err := http.NewRequest(
			http.MethodPost,
			testServer.URL+"/api/v1/deals/"+deal.ID.String()+"/reject",
			nil,
		)
		require.NoError(t, err)
		req.Header.Set("Authorization", s.pubToken)

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusNoContent, resp.StatusCode)
	})

	t.Run("invalid transition", func(t *testing.T) {
		s := setupDeal(t, ctx)

		deal, err := testTools.CreateDeal(ctx, s.channel.ID, s.advertiser.ID,
			entity.DealStatusApproved, time.Now().Add(48*time.Hour),
			entity.AdFormatTypePost, false, 24, 4, 1000000000)
		require.NoError(t, err)

		req, err := http.NewRequest(
			http.MethodPost,
			testServer.URL+"/api/v1/deals/"+deal.ID.String()+"/reject",
			nil,
		)
		require.NoError(t, err)
		req.Header.Set("Authorization", s.pubToken)

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})
}

func TestHandleRequestChanges(t *testing.T) {
	ctx := context.Background()

	t.Run("happy path", func(t *testing.T) {
		s := setupDeal(t, ctx)

		deal, err := testTools.CreateDeal(ctx, s.channel.ID, s.advertiser.ID,
			entity.DealStatusPendingReview, time.Now().Add(48*time.Hour),
			entity.AdFormatTypePost, false, 24, 4, 1000000000)
		require.NoError(t, err)

		body, _ := json.Marshal(dto.RequestChangesRequest{Note: "please fix the text"})
		req, err := http.NewRequest(
			http.MethodPost,
			testServer.URL+"/api/v1/deals/"+deal.ID.String()+"/request-changes",
			bytes.NewReader(body),
		)
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", s.pubToken)

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusNoContent, resp.StatusCode)
	})

	t.Run("missing note", func(t *testing.T) {
		s := setupDeal(t, ctx)

		deal, err := testTools.CreateDeal(ctx, s.channel.ID, s.advertiser.ID,
			entity.DealStatusPendingReview, time.Now().Add(48*time.Hour),
			entity.AdFormatTypePost, false, 24, 4, 1000000000)
		require.NoError(t, err)

		body, _ := json.Marshal(map[string]string{})
		req, err := http.NewRequest(
			http.MethodPost,
			testServer.URL+"/api/v1/deals/"+deal.ID.String()+"/request-changes",
			bytes.NewReader(body),
		)
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", s.pubToken)

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})
}

func TestHandleCancelDeal(t *testing.T) {
	ctx := context.Background()

	t.Run("happy path", func(t *testing.T) {
		s := setupDeal(t, ctx)

		deal, err := testTools.CreateDeal(ctx, s.channel.ID, s.advertiser.ID,
			entity.DealStatusPendingPayment, time.Now().Add(48*time.Hour),
			entity.AdFormatTypePost, false, 24, 4, 1000000000)
		require.NoError(t, err)

		req, err := http.NewRequest(
			http.MethodPost,
			testServer.URL+"/api/v1/deals/"+deal.ID.String()+"/cancel",
			nil,
		)
		require.NoError(t, err)
		req.Header.Set("Authorization", s.advToken)

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusNoContent, resp.StatusCode)
	})

	t.Run("invalid transition after approval", func(t *testing.T) {
		s := setupDeal(t, ctx)

		deal, err := testTools.CreateDeal(ctx, s.channel.ID, s.advertiser.ID,
			entity.DealStatusApproved, time.Now().Add(48*time.Hour),
			entity.AdFormatTypePost, false, 24, 4, 1000000000)
		require.NoError(t, err)

		req, err := http.NewRequest(
			http.MethodPost,
			testServer.URL+"/api/v1/deals/"+deal.ID.String()+"/cancel",
			nil,
		)
		require.NoError(t, err)
		req.Header.Set("Authorization", s.advToken)

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})
}

func strPtr(s string) *string { return &s }
