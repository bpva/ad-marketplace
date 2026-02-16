package deal

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/bpva/ad-marketplace/internal/dto"
	"github.com/bpva/ad-marketplace/internal/entity"
)

func newTestService(t *testing.T) (
	*svc,
	*MockDealRepository,
	*MockChannelRepository,
	*MockPostRepository,
	*MockUserRepository,
	*MockTransactor,
) {
	ctrl := gomock.NewController(t)
	dealRepo := NewMockDealRepository(ctrl)
	channelRepo := NewMockChannelRepository(ctrl)
	postRepo := NewMockPostRepository(ctrl)
	userRepo := NewMockUserRepository(ctrl)
	tx := NewMockTransactor(ctrl)
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	s := New(dealRepo, channelRepo, postRepo, userRepo, tx, log)
	return s, dealRepo, channelRepo, postRepo, userRepo, tx
}

func ctxWithUser(id uuid.UUID, tgID int64) context.Context {
	return dto.ContextWithUser(context.Background(), dto.UserContext{ID: id, TgID: tgID})
}

// --- State Machine ---

func TestCanTransition(t *testing.T) {
	valid := []struct {
		from entity.DealStatus
		to   entity.DealStatus
	}{
		{entity.DealStatusPendingPayment, entity.DealStatusPendingReview},
		{entity.DealStatusPendingPayment, entity.DealStatusHoldFailed},
		{entity.DealStatusPendingPayment, entity.DealStatusCancelled},
		{entity.DealStatusPendingReview, entity.DealStatusApproved},
		{entity.DealStatusPendingReview, entity.DealStatusRejected},
		{entity.DealStatusPendingReview, entity.DealStatusChangesRequested},
		{entity.DealStatusPendingReview, entity.DealStatusCancelled},
		{entity.DealStatusChangesRequested, entity.DealStatusPendingReview},
		{entity.DealStatusChangesRequested, entity.DealStatusCancelled},
		{entity.DealStatusApproved, entity.DealStatusPosted},
		{entity.DealStatusPosted, entity.DealStatusCompleted},
		{entity.DealStatusPosted, entity.DealStatusDispute},
	}
	for _, tt := range valid {
		assert.True(t, canTransition(tt.from, tt.to), "%s → %s should be valid", tt.from, tt.to)
	}

	invalid := []struct {
		from entity.DealStatus
		to   entity.DealStatus
	}{
		{entity.DealStatusPendingPayment, entity.DealStatusApproved},
		{entity.DealStatusPendingPayment, entity.DealStatusPosted},
		{entity.DealStatusPendingReview, entity.DealStatusPosted},
		{entity.DealStatusPendingReview, entity.DealStatusCompleted},
		{entity.DealStatusApproved, entity.DealStatusCancelled},
		{entity.DealStatusApproved, entity.DealStatusRejected},
		{entity.DealStatusRejected, entity.DealStatusPendingReview},
		{entity.DealStatusCancelled, entity.DealStatusPendingReview},
		{entity.DealStatusCompleted, entity.DealStatusDispute},
		{entity.DealStatusPosted, entity.DealStatusCancelled},
		{entity.DealStatusChangesRequested, entity.DealStatusApproved},
		{entity.DealStatusHoldFailed, entity.DealStatusPendingReview},
	}
	for _, tt := range invalid {
		assert.False(t, canTransition(tt.from, tt.to), "%s → %s should be invalid", tt.from, tt.to)
	}
}

// --- CreateDeal ---

var (
	userID    = uuid.Must(uuid.NewV7())
	channelID = uuid.Must(uuid.NewV7())
	dealID    = uuid.Must(uuid.NewV7())
	postID    = uuid.Must(uuid.NewV7())
)

func defaultCreateParams() CreateDealParams {
	return CreateDealParams{
		TgChannelID:    -1001234567890,
		FormatType:     entity.AdFormatTypePost,
		IsNative:       false,
		FeedHours:      24,
		TopHours:       2,
		PriceNanoTON:   5000000000,
		TemplatePostID: postID,
		ScheduledAt:    time.Now().Add(24 * time.Hour),
	}
}

func defaultChannel() *entity.Channel {
	return &entity.Channel{
		ID:          channelID,
		TgChannelID: -1001234567890,
		IsListed:    true,
	}
}

func defaultAdFormats() []entity.ChannelAdFormat {
	return []entity.ChannelAdFormat{
		{
			ID:           uuid.Must(uuid.NewV7()),
			ChannelID:    channelID,
			FormatType:   entity.AdFormatTypePost,
			IsNative:     false,
			FeedHours:    24,
			TopHours:     2,
			PriceNanoTON: 5000000000,
		},
	}
}

func defaultTemplatePost() *entity.Post {
	return &entity.Post{
		ID:         postID,
		Type:       entity.PostTypeTemplate,
		ExternalID: userID,
	}
}

func defaultUser() *entity.User {
	wallet := "UQBtest"
	return &entity.User{
		ID:            userID,
		TgID:          123456,
		WalletAddress: &wallet,
	}
}

func TestCreateDeal_NoContext(t *testing.T) {
	s, _, _, _, _, _ := newTestService(t)
	_, _, err := s.CreateDeal(context.Background(), defaultCreateParams())
	require.Error(t, err)
	assert.True(t, errors.Is(err, dto.ErrForbidden))
}

func TestCreateDeal_ChannelNotFound(t *testing.T) {
	s, _, channelRepo, _, _, _ := newTestService(t)
	ctx := ctxWithUser(userID, 123456)
	params := defaultCreateParams()

	channelRepo.EXPECT().
		GetByTgChannelID(ctx, params.TgChannelID).
		Return(nil, fmt.Errorf("get channel: %w", dto.ErrNotFound))

	_, _, err := s.CreateDeal(ctx, params)
	require.Error(t, err)
	assert.True(t, errors.Is(err, dto.ErrNotFound))
}

func TestCreateDeal_ChannelNotListed(t *testing.T) {
	s, _, channelRepo, _, _, _ := newTestService(t)
	ctx := ctxWithUser(userID, 123456)
	params := defaultCreateParams()

	ch := defaultChannel()
	ch.IsListed = false
	channelRepo.EXPECT().GetByTgChannelID(ctx, params.TgChannelID).Return(ch, nil)

	_, _, err := s.CreateDeal(ctx, params)
	require.Error(t, err)
	assert.True(t, errors.Is(err, dto.ErrChannelNotListed))
}

func TestCreateDeal_NoMatchingFormat(t *testing.T) {
	s, _, channelRepo, _, _, _ := newTestService(t)
	ctx := ctxWithUser(userID, 123456)
	params := defaultCreateParams()
	params.FeedHours = 12

	channelRepo.EXPECT().GetByTgChannelID(ctx, params.TgChannelID).Return(defaultChannel(), nil)
	channelRepo.EXPECT().GetAdFormatsByChannelID(ctx, channelID).Return(defaultAdFormats(), nil)

	_, _, err := s.CreateDeal(ctx, params)
	require.Error(t, err)
	assert.True(t, errors.Is(err, dto.ErrNotFound))
}

func TestCreateDeal_PriceMismatch(t *testing.T) {
	s, _, channelRepo, _, _, _ := newTestService(t)
	ctx := ctxWithUser(userID, 123456)
	params := defaultCreateParams()
	params.PriceNanoTON = 1

	channelRepo.EXPECT().GetByTgChannelID(ctx, params.TgChannelID).Return(defaultChannel(), nil)
	channelRepo.EXPECT().GetAdFormatsByChannelID(ctx, channelID).Return(defaultAdFormats(), nil)

	_, _, err := s.CreateDeal(ctx, params)
	require.Error(t, err)
	assert.True(t, errors.Is(err, dto.ErrPriceMismatch))
}

func TestCreateDeal_ScheduledInPast(t *testing.T) {
	s, _, channelRepo, _, _, _ := newTestService(t)
	ctx := ctxWithUser(userID, 123456)
	params := defaultCreateParams()
	params.ScheduledAt = time.Now().Add(-time.Hour)

	channelRepo.EXPECT().GetByTgChannelID(ctx, params.TgChannelID).Return(defaultChannel(), nil)
	channelRepo.EXPECT().GetAdFormatsByChannelID(ctx, channelID).Return(defaultAdFormats(), nil)

	_, _, err := s.CreateDeal(ctx, params)
	require.Error(t, err)
	requireAPIError(t, err, "invalid_request")
}

func TestCreateDeal_TemplateNotOwned(t *testing.T) {
	s, _, channelRepo, postRepo, _, _ := newTestService(t)
	ctx := ctxWithUser(userID, 123456)
	params := defaultCreateParams()

	otherUserPost := &entity.Post{
		ID:         postID,
		Type:       entity.PostTypeTemplate,
		ExternalID: uuid.Must(uuid.NewV7()),
	}

	channelRepo.EXPECT().GetByTgChannelID(ctx, params.TgChannelID).Return(defaultChannel(), nil)
	channelRepo.EXPECT().GetAdFormatsByChannelID(ctx, channelID).Return(defaultAdFormats(), nil)
	postRepo.EXPECT().GetByID(ctx, params.TemplatePostID).Return(otherUserPost, nil)

	_, _, err := s.CreateDeal(ctx, params)
	require.Error(t, err)
	assert.True(t, errors.Is(err, dto.ErrForbidden))
}

func TestCreateDeal_AdPostNotTemplate(t *testing.T) {
	s, _, channelRepo, postRepo, _, _ := newTestService(t)
	ctx := ctxWithUser(userID, 123456)
	params := defaultCreateParams()

	adPost := &entity.Post{
		ID:         postID,
		Type:       entity.PostTypeAd,
		ExternalID: userID,
	}

	channelRepo.EXPECT().GetByTgChannelID(ctx, params.TgChannelID).Return(defaultChannel(), nil)
	channelRepo.EXPECT().GetAdFormatsByChannelID(ctx, channelID).Return(defaultAdFormats(), nil)
	postRepo.EXPECT().GetByID(ctx, params.TemplatePostID).Return(adPost, nil)

	_, _, err := s.CreateDeal(ctx, params)
	require.Error(t, err)
	assert.True(t, errors.Is(err, dto.ErrForbidden))
}

func TestCreateDeal_Success(t *testing.T) {
	s, dealRepo, channelRepo, postRepo, userRepo, tx := newTestService(t)
	ctx := ctxWithUser(userID, 123456)
	params := defaultCreateParams()

	payoutWallet := "UQBpayout"
	createdDeal := &entity.Deal{
		ID:                      dealID,
		ChannelID:               channelID,
		AdvertiserID:            userID,
		Status:                  entity.DealStatusPendingPayment,
		ScheduledAt:             params.ScheduledAt,
		AdvertiserWalletAddress: defaultUser().WalletAddress,
		PayoutWalletAddress:     &payoutWallet,
		FormatType:              entity.AdFormatTypePost,
		IsNative:                false,
		FeedHours:               24,
		TopHours:                2,
		PriceNanoTON:            5000000000,
	}
	copiedPosts := []entity.Post{{ID: uuid.Must(uuid.NewV7()), Type: entity.PostTypeAd, ExternalID: dealID}}

	channelRepo.EXPECT().GetByTgChannelID(ctx, params.TgChannelID).Return(defaultChannel(), nil)
	channelRepo.EXPECT().GetAdFormatsByChannelID(ctx, channelID).Return(defaultAdFormats(), nil)
	postRepo.EXPECT().GetByID(ctx, params.TemplatePostID).Return(defaultTemplatePost(), nil)
	userRepo.EXPECT().GetByID(ctx, userID).Return(defaultUser(), nil)
	channelRepo.EXPECT().GetOwnerWalletAddress(ctx, channelID).Return(&payoutWallet, nil)

	tx.EXPECT().WithTx(ctx, gomock.Any()).DoAndReturn(
		func(ctx context.Context, f func(context.Context) error) error {
			return f(ctx)
		},
	)
	dealRepo.EXPECT().Create(ctx, gomock.Any()).DoAndReturn(
		func(_ context.Context, d *entity.Deal) (*entity.Deal, error) {
			assert.Equal(t, entity.DealStatusPendingPayment, d.Status)
			assert.Equal(t, channelID, d.ChannelID)
			assert.Equal(t, userID, d.AdvertiserID)
			assert.Equal(t, entity.AdFormatTypePost, d.FormatType)
			assert.Equal(t, int64(5000000000), d.PriceNanoTON)
			assert.Equal(t, defaultUser().WalletAddress, d.AdvertiserWalletAddress)
			assert.Equal(t, &payoutWallet, d.PayoutWalletAddress)
			return createdDeal, nil
		},
	)
	postRepo.EXPECT().CopyAsAd(ctx, params.TemplatePostID, dealID, 1).Return(copiedPosts, nil)

	deal, posts, err := s.CreateDeal(ctx, params)
	require.NoError(t, err)
	assert.Equal(t, dealID, deal.ID)
	assert.Equal(t, entity.DealStatusPendingPayment, deal.Status)
	assert.Len(t, posts, 1)
	assert.Equal(t, entity.PostTypeAd, posts[0].Type)
}

// --- Approve ---

func TestApprove_NoContext(t *testing.T) {
	s, _, _, _, _, _ := newTestService(t)
	err := s.Approve(context.Background(), dealID)
	require.Error(t, err)
	assert.True(t, errors.Is(err, dto.ErrForbidden))
}

func TestApprove_DealNotFound(t *testing.T) {
	s, dealRepo, _, _, _, _ := newTestService(t)
	ctx := ctxWithUser(userID, 123456)

	dealRepo.EXPECT().GetByID(ctx, dealID).Return(nil, fmt.Errorf("get: %w", dto.ErrNotFound))

	err := s.Approve(ctx, dealID)
	require.Error(t, err)
	assert.True(t, errors.Is(err, dto.ErrNotFound))
}

func TestApprove_NoRole(t *testing.T) {
	s, dealRepo, channelRepo, _, _, _ := newTestService(t)
	ctx := ctxWithUser(userID, 123456)

	deal := &entity.Deal{ID: dealID, ChannelID: channelID, Status: entity.DealStatusPendingReview}
	dealRepo.EXPECT().GetByID(ctx, dealID).Return(deal, nil)
	channelRepo.EXPECT().GetRole(ctx, channelID, userID).Return(nil, fmt.Errorf("get role: %w", dto.ErrNotFound))

	err := s.Approve(ctx, dealID)
	require.Error(t, err)
	assert.True(t, errors.Is(err, dto.ErrForbidden))
}

func TestApprove_WrongStatus(t *testing.T) {
	s, dealRepo, channelRepo, _, _, _ := newTestService(t)
	ctx := ctxWithUser(userID, 123456)

	deal := &entity.Deal{ID: dealID, ChannelID: channelID, Status: entity.DealStatusChangesRequested}
	dealRepo.EXPECT().GetByID(ctx, dealID).Return(deal, nil)
	channelRepo.EXPECT().GetRole(ctx, channelID, userID).Return(&entity.ChannelRole{Role: entity.ChannelRoleTypeOwner}, nil)

	err := s.Approve(ctx, dealID)
	require.Error(t, err)
	assert.True(t, errors.Is(err, dto.ErrInvalidTransition))
}

func TestApprove_Success(t *testing.T) {
	s, dealRepo, channelRepo, _, _, _ := newTestService(t)
	ctx := ctxWithUser(userID, 123456)

	deal := &entity.Deal{ID: dealID, ChannelID: channelID, Status: entity.DealStatusPendingReview}
	dealRepo.EXPECT().GetByID(ctx, dealID).Return(deal, nil)
	channelRepo.EXPECT().GetRole(ctx, channelID, userID).Return(&entity.ChannelRole{Role: entity.ChannelRoleTypeOwner}, nil)
	dealRepo.EXPECT().UpdateStatus(ctx, dealID, entity.DealStatusApproved, (*string)(nil)).Return(nil)

	err := s.Approve(ctx, dealID)
	require.NoError(t, err)
}

// --- Reject ---

func TestReject_WrongStatus(t *testing.T) {
	s, dealRepo, channelRepo, _, _, _ := newTestService(t)
	ctx := ctxWithUser(userID, 123456)

	deal := &entity.Deal{ID: dealID, ChannelID: channelID, Status: entity.DealStatusApproved}
	dealRepo.EXPECT().GetByID(ctx, dealID).Return(deal, nil)
	channelRepo.EXPECT().GetRole(ctx, channelID, userID).Return(&entity.ChannelRole{Role: entity.ChannelRoleTypeOwner}, nil)

	err := s.Reject(ctx, dealID, nil)
	require.Error(t, err)
	assert.True(t, errors.Is(err, dto.ErrInvalidTransition))
}

func TestReject_Success(t *testing.T) {
	s, dealRepo, channelRepo, _, _, _ := newTestService(t)
	ctx := ctxWithUser(userID, 123456)
	reason := "bad quality"

	deal := &entity.Deal{ID: dealID, ChannelID: channelID, Status: entity.DealStatusPendingReview}
	dealRepo.EXPECT().GetByID(ctx, dealID).Return(deal, nil)
	channelRepo.EXPECT().GetRole(ctx, channelID, userID).Return(&entity.ChannelRole{Role: entity.ChannelRoleTypeManager}, nil)
	dealRepo.EXPECT().UpdateStatus(ctx, dealID, entity.DealStatusRejected, &reason).Return(nil)

	err := s.Reject(ctx, dealID, &reason)
	require.NoError(t, err)
}

// --- RequestChanges ---

func TestRequestChanges_WrongStatus(t *testing.T) {
	s, dealRepo, channelRepo, _, _, _ := newTestService(t)
	ctx := ctxWithUser(userID, 123456)

	deal := &entity.Deal{ID: dealID, ChannelID: channelID, Status: entity.DealStatusPendingPayment}
	dealRepo.EXPECT().GetByID(ctx, dealID).Return(deal, nil)
	channelRepo.EXPECT().GetRole(ctx, channelID, userID).Return(&entity.ChannelRole{Role: entity.ChannelRoleTypeOwner}, nil)

	err := s.RequestChanges(ctx, dealID, "fix text")
	require.Error(t, err)
	assert.True(t, errors.Is(err, dto.ErrInvalidTransition))
}

func TestRequestChanges_Success(t *testing.T) {
	s, dealRepo, channelRepo, _, _, _ := newTestService(t)
	ctx := ctxWithUser(userID, 123456)
	note := "fix text"

	deal := &entity.Deal{ID: dealID, ChannelID: channelID, Status: entity.DealStatusPendingReview}
	dealRepo.EXPECT().GetByID(ctx, dealID).Return(deal, nil)
	channelRepo.EXPECT().GetRole(ctx, channelID, userID).Return(&entity.ChannelRole{Role: entity.ChannelRoleTypeOwner}, nil)
	dealRepo.EXPECT().UpdateStatus(ctx, dealID, entity.DealStatusChangesRequested, &note).Return(nil)

	err := s.RequestChanges(ctx, dealID, note)
	require.NoError(t, err)
}

// --- SubmitRevision ---

func TestSubmitRevision_NoContext(t *testing.T) {
	s, _, _, _, _, _ := newTestService(t)
	_, err := s.SubmitRevision(context.Background(), dealID, nil)
	require.Error(t, err)
	assert.True(t, errors.Is(err, dto.ErrForbidden))
}

func TestSubmitRevision_NotAdvertiser(t *testing.T) {
	s, dealRepo, _, _, _, _ := newTestService(t)
	otherUser := uuid.Must(uuid.NewV7())
	ctx := ctxWithUser(otherUser, 999)

	deal := &entity.Deal{ID: dealID, AdvertiserID: userID, Status: entity.DealStatusChangesRequested}
	dealRepo.EXPECT().GetByID(ctx, dealID).Return(deal, nil)

	_, err := s.SubmitRevision(ctx, dealID, nil)
	require.Error(t, err)
	assert.True(t, errors.Is(err, dto.ErrForbidden))
}

func TestSubmitRevision_WrongStatus(t *testing.T) {
	s, dealRepo, _, _, _, _ := newTestService(t)
	ctx := ctxWithUser(userID, 123456)

	deal := &entity.Deal{ID: dealID, AdvertiserID: userID, Status: entity.DealStatusPendingReview}
	dealRepo.EXPECT().GetByID(ctx, dealID).Return(deal, nil)

	_, err := s.SubmitRevision(ctx, dealID, nil)
	require.Error(t, err)
	assert.True(t, errors.Is(err, dto.ErrInvalidTransition))
}

func TestSubmitRevision_Success(t *testing.T) {
	s, dealRepo, _, postRepo, _, tx := newTestService(t)
	ctx := ctxWithUser(userID, 123456)

	deal := &entity.Deal{ID: dealID, AdvertiserID: userID, Status: entity.DealStatusChangesRequested}
	dealRepo.EXPECT().GetByID(ctx, dealID).Return(deal, nil)

	v1 := 1
	latestPosts := []entity.Post{{ID: uuid.Must(uuid.NewV7()), Version: &v1}}
	postRepo.EXPECT().GetLatestAd(ctx, dealID).Return(latestPosts, nil)

	newPosts := []entity.Post{{Text: strPtr("new text")}}
	createdPosts := []entity.Post{{ID: uuid.Must(uuid.NewV7()), Type: entity.PostTypeAd, ExternalID: dealID}}

	tx.EXPECT().WithTx(ctx, gomock.Any()).DoAndReturn(
		func(ctx context.Context, f func(context.Context) error) error {
			return f(ctx)
		},
	)
	postRepo.EXPECT().AddAdVersion(ctx, dealID, 2, newPosts).Return(createdPosts, nil)
	dealRepo.EXPECT().UpdateStatus(ctx, dealID, entity.DealStatusPendingReview, (*string)(nil)).Return(nil)

	posts, err := s.SubmitRevision(ctx, dealID, newPosts)
	require.NoError(t, err)
	assert.Len(t, posts, 1)
}

// --- Cancel ---

func TestCancel_NoContext(t *testing.T) {
	s, _, _, _, _, _ := newTestService(t)
	err := s.Cancel(context.Background(), dealID)
	require.Error(t, err)
	assert.True(t, errors.Is(err, dto.ErrForbidden))
}

func TestCancel_NotAdvertiser(t *testing.T) {
	s, dealRepo, _, _, _, _ := newTestService(t)
	otherUser := uuid.Must(uuid.NewV7())
	ctx := ctxWithUser(otherUser, 999)

	deal := &entity.Deal{
		ID: dealID, AdvertiserID: userID,
		Status: entity.DealStatusPendingPayment, ScheduledAt: time.Now().Add(24 * time.Hour),
	}
	dealRepo.EXPECT().GetByID(ctx, dealID).Return(deal, nil)

	err := s.Cancel(ctx, dealID)
	require.Error(t, err)
	assert.True(t, errors.Is(err, dto.ErrForbidden))
}

func TestCancel_StatusApproved(t *testing.T) {
	s, dealRepo, _, _, _, _ := newTestService(t)
	ctx := ctxWithUser(userID, 123456)

	deal := &entity.Deal{
		ID: dealID, AdvertiserID: userID,
		Status: entity.DealStatusApproved, ScheduledAt: time.Now().Add(24 * time.Hour),
	}
	dealRepo.EXPECT().GetByID(ctx, dealID).Return(deal, nil)

	err := s.Cancel(ctx, dealID)
	require.Error(t, err)
	assert.True(t, errors.Is(err, dto.ErrInvalidTransition))
}

func TestCancel_ScheduledTimePassed(t *testing.T) {
	s, dealRepo, _, _, _, _ := newTestService(t)
	ctx := ctxWithUser(userID, 123456)

	deal := &entity.Deal{
		ID: dealID, AdvertiserID: userID,
		Status: entity.DealStatusPendingReview, ScheduledAt: time.Now().Add(-time.Hour),
	}
	dealRepo.EXPECT().GetByID(ctx, dealID).Return(deal, nil)

	err := s.Cancel(ctx, dealID)
	require.Error(t, err)
	requireAPIError(t, err, "invalid_request")
}

func TestCancel_Success_PendingPayment(t *testing.T) {
	s, dealRepo, _, _, _, _ := newTestService(t)
	ctx := ctxWithUser(userID, 123456)

	deal := &entity.Deal{
		ID: dealID, AdvertiserID: userID,
		Status: entity.DealStatusPendingPayment, ScheduledAt: time.Now().Add(24 * time.Hour),
	}
	dealRepo.EXPECT().GetByID(ctx, dealID).Return(deal, nil)
	dealRepo.EXPECT().UpdateStatus(ctx, dealID, entity.DealStatusCancelled, (*string)(nil)).Return(nil)

	err := s.Cancel(ctx, dealID)
	require.NoError(t, err)
}

func TestCancel_Success_PendingReview(t *testing.T) {
	s, dealRepo, _, _, _, _ := newTestService(t)
	ctx := ctxWithUser(userID, 123456)

	deal := &entity.Deal{
		ID: dealID, AdvertiserID: userID,
		Status: entity.DealStatusPendingReview, ScheduledAt: time.Now().Add(24 * time.Hour),
	}
	dealRepo.EXPECT().GetByID(ctx, dealID).Return(deal, nil)
	dealRepo.EXPECT().UpdateStatus(ctx, dealID, entity.DealStatusCancelled, (*string)(nil)).Return(nil)

	err := s.Cancel(ctx, dealID)
	require.NoError(t, err)
}

func TestCancel_Success_ChangesRequested(t *testing.T) {
	s, dealRepo, _, _, _, _ := newTestService(t)
	ctx := ctxWithUser(userID, 123456)

	deal := &entity.Deal{
		ID: dealID, AdvertiserID: userID,
		Status: entity.DealStatusChangesRequested, ScheduledAt: time.Now().Add(24 * time.Hour),
	}
	dealRepo.EXPECT().GetByID(ctx, dealID).Return(deal, nil)
	dealRepo.EXPECT().UpdateStatus(ctx, dealID, entity.DealStatusCancelled, (*string)(nil)).Return(nil)

	err := s.Cancel(ctx, dealID)
	require.NoError(t, err)
}

// --- GetDeal ---

func TestGetDeal_NoContext(t *testing.T) {
	s, _, _, _, _, _ := newTestService(t)
	_, _, err := s.GetDeal(context.Background(), dealID)
	require.Error(t, err)
	assert.True(t, errors.Is(err, dto.ErrForbidden))
}

func TestGetDeal_AsAdvertiser(t *testing.T) {
	s, dealRepo, _, postRepo, _, _ := newTestService(t)
	ctx := ctxWithUser(userID, 123456)

	deal := &entity.Deal{ID: dealID, ChannelID: channelID, AdvertiserID: userID}
	posts := []entity.Post{{ID: uuid.Must(uuid.NewV7())}}

	dealRepo.EXPECT().GetByID(ctx, dealID).Return(deal, nil)
	postRepo.EXPECT().GetLatestAd(ctx, dealID).Return(posts, nil)

	d, p, err := s.GetDeal(ctx, dealID)
	require.NoError(t, err)
	assert.Equal(t, dealID, d.ID)
	assert.Len(t, p, 1)
}

func TestGetDeal_AsPublisher(t *testing.T) {
	s, dealRepo, channelRepo, postRepo, _, _ := newTestService(t)
	publisherID := uuid.Must(uuid.NewV7())
	ctx := ctxWithUser(publisherID, 999)

	deal := &entity.Deal{ID: dealID, ChannelID: channelID, AdvertiserID: userID}
	posts := []entity.Post{{ID: uuid.Must(uuid.NewV7())}}

	dealRepo.EXPECT().GetByID(ctx, dealID).Return(deal, nil)
	channelRepo.EXPECT().GetRole(ctx, channelID, publisherID).Return(&entity.ChannelRole{Role: entity.ChannelRoleTypeOwner}, nil)
	postRepo.EXPECT().GetLatestAd(ctx, dealID).Return(posts, nil)

	d, p, err := s.GetDeal(ctx, dealID)
	require.NoError(t, err)
	assert.Equal(t, dealID, d.ID)
	assert.Len(t, p, 1)
}

func TestGetDeal_Unauthorized(t *testing.T) {
	s, dealRepo, channelRepo, _, _, _ := newTestService(t)
	otherUser := uuid.Must(uuid.NewV7())
	ctx := ctxWithUser(otherUser, 999)

	deal := &entity.Deal{ID: dealID, ChannelID: channelID, AdvertiserID: userID}
	dealRepo.EXPECT().GetByID(ctx, dealID).Return(deal, nil)
	channelRepo.EXPECT().GetRole(ctx, channelID, otherUser).Return(nil, fmt.Errorf("get role: %w", dto.ErrNotFound))

	_, _, err := s.GetDeal(ctx, dealID)
	require.Error(t, err)
	assert.True(t, errors.Is(err, dto.ErrForbidden))
}

// --- ListPublisherDeals ---

func TestListPublisherDeals_NoRole(t *testing.T) {
	s, _, channelRepo, _, _, _ := newTestService(t)
	ctx := ctxWithUser(userID, 123456)

	channelRepo.EXPECT().GetByTgChannelID(ctx, int64(-1001234567890)).Return(defaultChannel(), nil)
	channelRepo.EXPECT().GetRole(ctx, channelID, userID).Return(nil, fmt.Errorf("get role: %w", dto.ErrNotFound))

	_, _, err := s.ListPublisherDeals(ctx, -1001234567890, 10, 0)
	require.Error(t, err)
	assert.True(t, errors.Is(err, dto.ErrForbidden))
}

func TestListPublisherDeals_Success(t *testing.T) {
	s, dealRepo, channelRepo, _, _, _ := newTestService(t)
	ctx := ctxWithUser(userID, 123456)

	deals := []entity.Deal{{ID: dealID}}
	channelRepo.EXPECT().GetByTgChannelID(ctx, int64(-1001234567890)).Return(defaultChannel(), nil)
	channelRepo.EXPECT().GetRole(ctx, channelID, userID).Return(&entity.ChannelRole{Role: entity.ChannelRoleTypeOwner}, nil)
	dealRepo.EXPECT().GetByChannelID(ctx, channelID, 10, 0).Return(deals, 1, nil)

	result, total, err := s.ListPublisherDeals(ctx, -1001234567890, 10, 0)
	require.NoError(t, err)
	assert.Len(t, result, 1)
	assert.Equal(t, 1, total)
}

// --- ListAdvertiserDeals ---

func TestListAdvertiserDeals_NoContext(t *testing.T) {
	s, _, _, _, _, _ := newTestService(t)
	_, _, err := s.ListAdvertiserDeals(context.Background(), 10, 0)
	require.Error(t, err)
	assert.True(t, errors.Is(err, dto.ErrForbidden))
}

func TestListAdvertiserDeals_Success(t *testing.T) {
	s, dealRepo, _, _, _, _ := newTestService(t)
	ctx := ctxWithUser(userID, 123456)

	deals := []entity.Deal{{ID: dealID}}
	dealRepo.EXPECT().GetByAdvertiserID(ctx, userID, 10, 0).Return(deals, 1, nil)

	result, total, err := s.ListAdvertiserDeals(ctx, 10, 0)
	require.NoError(t, err)
	assert.Len(t, result, 1)
	assert.Equal(t, 1, total)
}

func strPtr(s string) *string { return &s }

func requireAPIError(t *testing.T, err error, code string) {
	t.Helper()
	var apiErr *dto.APIError
	require.True(t, errors.As(err, &apiErr), "expected APIError, got: %v", err)
	assert.Equal(t, code, apiErr.Code())
}
