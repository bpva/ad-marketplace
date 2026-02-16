package deal

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"slices"
	"time"

	"github.com/google/uuid"

	"github.com/bpva/ad-marketplace/internal/dto"
	"github.com/bpva/ad-marketplace/internal/entity"
	"github.com/bpva/ad-marketplace/internal/logx"
)

//go:generate mockgen -destination=mocks.go -package=deal . DealRepository,ChannelRepository,PostRepository,UserRepository,Transactor

type DealRepository interface {
	Create(ctx context.Context, deal *entity.Deal) (*entity.Deal, error)
	GetByID(ctx context.Context, id uuid.UUID) (*entity.Deal, error)
	GetByChannelID(
		ctx context.Context,
		channelID uuid.UUID,
		limit, offset int,
	) ([]entity.Deal, int, error)
	GetByAdvertiserID(
		ctx context.Context,
		advertiserID uuid.UUID,
		limit, offset int,
	) ([]entity.Deal, int, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, status entity.DealStatus, note *string) error
}

type ChannelRepository interface {
	GetByTgChannelID(ctx context.Context, tgChannelID int64) (*entity.Channel, error)
	GetByID(ctx context.Context, id uuid.UUID) (*entity.Channel, error)
	GetRole(ctx context.Context, channelID, userID uuid.UUID) (*entity.ChannelRole, error)
	GetAdFormatsByChannelID(
		ctx context.Context,
		channelID uuid.UUID,
	) ([]entity.ChannelAdFormat, error)
	GetOwnerWalletAddress(ctx context.Context, channelID uuid.UUID) (*string, error)
}

type PostRepository interface {
	GetByID(ctx context.Context, id uuid.UUID) (*entity.Post, error)
	CopyAsAd(
		ctx context.Context,
		templatePostID, dealID uuid.UUID,
		version int,
	) ([]entity.Post, error)
	AddAdVersion(
		ctx context.Context,
		dealID uuid.UUID,
		version int,
		posts []entity.Post,
	) ([]entity.Post, error)
	GetLatestAd(ctx context.Context, dealID uuid.UUID) ([]entity.Post, error)
}

type UserRepository interface {
	GetByID(ctx context.Context, id uuid.UUID) (*entity.User, error)
}

type Transactor interface {
	WithTx(ctx context.Context, f func(ctx context.Context) error) error
}

var validTransitions = map[entity.DealStatus][]entity.DealStatus{
	entity.DealStatusPendingPayment: {
		entity.DealStatusPendingReview,
		entity.DealStatusHoldFailed,
		entity.DealStatusCancelled,
	},
	entity.DealStatusPendingReview: {
		entity.DealStatusApproved,
		entity.DealStatusRejected,
		entity.DealStatusChangesRequested,
		entity.DealStatusCancelled,
	},
	entity.DealStatusChangesRequested: {entity.DealStatusPendingReview, entity.DealStatusCancelled},
	entity.DealStatusApproved:         {entity.DealStatusPosted},
	entity.DealStatusPosted:           {entity.DealStatusCompleted, entity.DealStatusDispute},
}

func canTransition(from, to entity.DealStatus) bool {
	return slices.Contains(validTransitions[from], to)
}

type CreateDealParams struct {
	TgChannelID    int64
	FormatType     entity.AdFormatType
	IsNative       bool
	FeedHours      int
	TopHours       int
	PriceNanoTON   int64
	TemplatePostID uuid.UUID
	ScheduledAt    time.Time
}

type svc struct {
	dealRepo    DealRepository
	channelRepo ChannelRepository
	postRepo    PostRepository
	userRepo    UserRepository
	tx          Transactor
	log         *slog.Logger
}

func New(
	dealRepo DealRepository,
	channelRepo ChannelRepository,
	postRepo PostRepository,
	userRepo UserRepository,
	tx Transactor,
	log *slog.Logger,
) *svc {
	log = log.With(logx.Service("DealService"))
	return &svc{
		dealRepo:    dealRepo,
		channelRepo: channelRepo,
		postRepo:    postRepo,
		userRepo:    userRepo,
		tx:          tx,
		log:         log,
	}
}

func (s *svc) CreateDeal(
	ctx context.Context,
	params CreateDealParams,
) (*entity.Deal, []entity.Post, error) {
	user, ok := dto.UserFromContext(ctx)
	if !ok {
		return nil, nil, fmt.Errorf("create deal: %w", dto.ErrForbidden)
	}

	channel, err := s.channelRepo.GetByTgChannelID(ctx, params.TgChannelID)
	if err != nil {
		return nil, nil, fmt.Errorf("get channel: %w", err)
	}

	if !channel.IsListed {
		return nil, nil, fmt.Errorf("create deal: %w", dto.ErrChannelNotListed)
	}

	formats, err := s.channelRepo.GetAdFormatsByChannelID(ctx, channel.ID)
	if err != nil {
		return nil, nil, fmt.Errorf("get ad formats: %w", err)
	}

	var matched *entity.ChannelAdFormat
	for i := range formats {
		f := &formats[i]
		if f.FormatType == params.FormatType &&
			f.IsNative == params.IsNative &&
			f.FeedHours == params.FeedHours &&
			f.TopHours == params.TopHours {
			matched = f
			break
		}
	}
	if matched == nil {
		return nil, nil, fmt.Errorf("find ad format: %w", dto.ErrNotFound)
	}

	if matched.PriceNanoTON != params.PriceNanoTON {
		return nil, nil, fmt.Errorf("create deal: %w", dto.ErrPriceMismatch)
	}

	if time.Now().After(params.ScheduledAt) {
		return nil, nil, fmt.Errorf("create deal: %w",
			dto.ErrValidation.WithDetails(map[string]any{"scheduled_at": "must be in the future"}))
	}

	tmpl, err := s.postRepo.GetByID(ctx, params.TemplatePostID)
	if err != nil {
		return nil, nil, fmt.Errorf("get template: %w", err)
	}

	if tmpl.Type != entity.PostTypeTemplate || tmpl.ExternalID != user.ID {
		return nil, nil, fmt.Errorf("create deal: %w", dto.ErrForbidden)
	}

	advertiser, err := s.userRepo.GetByID(ctx, user.ID)
	if err != nil {
		return nil, nil, fmt.Errorf("get advertiser: %w", err)
	}

	payoutWallet, err := s.channelRepo.GetOwnerWalletAddress(ctx, channel.ID)
	if err != nil {
		return nil, nil, fmt.Errorf("get payout wallet: %w", err)
	}

	deal := &entity.Deal{
		ChannelID:               channel.ID,
		AdvertiserID:            user.ID,
		Status:                  entity.DealStatusPendingPayment,
		ScheduledAt:             params.ScheduledAt,
		AdvertiserWalletAddress: advertiser.WalletAddress,
		PayoutWalletAddress:     payoutWallet,
		FormatType:              matched.FormatType,
		IsNative:                matched.IsNative,
		FeedHours:               matched.FeedHours,
		TopHours:                matched.TopHours,
		PriceNanoTON:            matched.PriceNanoTON,
	}

	var created *entity.Deal
	var posts []entity.Post
	if err := s.tx.WithTx(ctx, func(txCtx context.Context) error {
		var txErr error
		created, txErr = s.dealRepo.Create(txCtx, deal)
		if txErr != nil {
			return fmt.Errorf("create deal: %w", txErr)
		}
		posts, txErr = s.postRepo.CopyAsAd(txCtx, params.TemplatePostID, created.ID, 1)
		if txErr != nil {
			return fmt.Errorf("copy template: %w", txErr)
		}
		return nil
	}); err != nil {
		return nil, nil, err
	}

	s.log.Info("deal created",
		"deal_id", created.ID,
		"channel_id", channel.TgChannelID,
		"advertiser_id", user.TgID,
	)

	return created, posts, nil
}

func (s *svc) GetDeal(
	ctx context.Context,
	dealID uuid.UUID,
) (*entity.Deal, []entity.Post, int64, error) {
	user, ok := dto.UserFromContext(ctx)
	if !ok {
		return nil, nil, 0, fmt.Errorf("get deal: %w", dto.ErrForbidden)
	}

	deal, err := s.dealRepo.GetByID(ctx, dealID)
	if err != nil {
		return nil, nil, 0, fmt.Errorf("get deal: %w", err)
	}

	channel, err := s.channelRepo.GetByID(ctx, deal.ChannelID)
	if err != nil {
		return nil, nil, 0, fmt.Errorf("get channel: %w", err)
	}

	if deal.AdvertiserID != user.ID {
		_, err := s.channelRepo.GetRole(ctx, deal.ChannelID, user.ID)
		if errors.Is(err, dto.ErrNotFound) {
			return nil, nil, 0, fmt.Errorf("get deal: %w", dto.ErrForbidden)
		}
		if err != nil {
			return nil, nil, 0, fmt.Errorf("get role: %w", err)
		}
	}

	posts, err := s.postRepo.GetLatestAd(ctx, dealID)
	if err != nil {
		return nil, nil, 0, fmt.Errorf("get latest ad: %w", err)
	}

	return deal, posts, channel.TgChannelID, nil
}

func (s *svc) ListAdvertiserDeals(
	ctx context.Context,
	limit, offset int,
) ([]dto.DealListItem, int, error) {
	user, ok := dto.UserFromContext(ctx)
	if !ok {
		return nil, 0, fmt.Errorf("list advertiser deals: %w", dto.ErrForbidden)
	}

	deals, total, err := s.dealRepo.GetByAdvertiserID(ctx, user.ID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("list advertiser deals: %w", err)
	}

	items := make([]dto.DealListItem, len(deals))
	channelCache := make(map[uuid.UUID]int64)
	for i := range deals {
		tgChannelID, ok := channelCache[deals[i].ChannelID]
		if !ok {
			ch, err := s.channelRepo.GetByID(ctx, deals[i].ChannelID)
			if err != nil {
				return nil, 0, fmt.Errorf("get channel: %w", err)
			}
			tgChannelID = ch.TgChannelID
			channelCache[deals[i].ChannelID] = tgChannelID
		}
		items[i] = dto.DealListItem{
			Deal:        deals[i],
			TgChannelID: tgChannelID,
		}
	}

	return items, total, nil
}

func (s *svc) ListPublisherDeals(
	ctx context.Context,
	tgChannelID int64,
	limit, offset int,
) ([]dto.DealListItem, int, error) {
	user, ok := dto.UserFromContext(ctx)
	if !ok {
		return nil, 0, fmt.Errorf("list publisher deals: %w", dto.ErrForbidden)
	}

	channel, err := s.channelRepo.GetByTgChannelID(ctx, tgChannelID)
	if err != nil {
		return nil, 0, fmt.Errorf("get channel: %w", err)
	}

	_, err = s.channelRepo.GetRole(ctx, channel.ID, user.ID)
	if errors.Is(err, dto.ErrNotFound) {
		return nil, 0, fmt.Errorf("list publisher deals: %w", dto.ErrForbidden)
	}
	if err != nil {
		return nil, 0, fmt.Errorf("get role: %w", err)
	}

	deals, total, err := s.dealRepo.GetByChannelID(ctx, channel.ID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("list publisher deals: %w", err)
	}

	items := make([]dto.DealListItem, len(deals))
	for i := range deals {
		items[i] = dto.DealListItem{
			Deal:        deals[i],
			TgChannelID: tgChannelID,
		}
	}

	return items, total, nil
}

func (s *svc) Approve(ctx context.Context, dealID uuid.UUID) error {
	deal, err := s.requirePublisherRole(ctx, dealID)
	if err != nil {
		return err
	}

	if !canTransition(deal.Status, entity.DealStatusApproved) {
		return fmt.Errorf("approve deal: %w", dto.ErrInvalidTransition)
	}

	if err := s.dealRepo.UpdateStatus(ctx, dealID, entity.DealStatusApproved, nil); err != nil {
		return fmt.Errorf("approve deal: %w", err)
	}

	s.log.Info("deal approved", "deal_id", dealID)
	return nil
}

func (s *svc) Reject(ctx context.Context, dealID uuid.UUID, reason *string) error {
	deal, err := s.requirePublisherRole(ctx, dealID)
	if err != nil {
		return err
	}

	if !canTransition(deal.Status, entity.DealStatusRejected) {
		return fmt.Errorf("reject deal: %w", dto.ErrInvalidTransition)
	}

	if err := s.dealRepo.UpdateStatus(ctx, dealID, entity.DealStatusRejected, reason); err != nil {
		return fmt.Errorf("reject deal: %w", err)
	}

	s.log.Info("deal rejected", "deal_id", dealID)
	return nil
}

func (s *svc) RequestChanges(ctx context.Context, dealID uuid.UUID, note string) error {
	deal, err := s.requirePublisherRole(ctx, dealID)
	if err != nil {
		return err
	}

	if !canTransition(deal.Status, entity.DealStatusChangesRequested) {
		return fmt.Errorf("request changes: %w", dto.ErrInvalidTransition)
	}

	if err := s.dealRepo.UpdateStatus(
		ctx,
		dealID,
		entity.DealStatusChangesRequested,
		&note,
	); err != nil {
		return fmt.Errorf("request changes: %w", err)
	}

	s.log.Info("changes requested", "deal_id", dealID)
	return nil
}

func (s *svc) SubmitRevision(
	ctx context.Context,
	dealID uuid.UUID,
	newPosts []entity.Post,
) ([]entity.Post, error) {
	user, ok := dto.UserFromContext(ctx)
	if !ok {
		return nil, fmt.Errorf("submit revision: %w", dto.ErrForbidden)
	}

	deal, err := s.dealRepo.GetByID(ctx, dealID)
	if err != nil {
		return nil, fmt.Errorf("get deal: %w", err)
	}

	if deal.AdvertiserID != user.ID {
		return nil, fmt.Errorf("submit revision: %w", dto.ErrForbidden)
	}

	if !canTransition(deal.Status, entity.DealStatusPendingReview) {
		return nil, fmt.Errorf("submit revision: %w", dto.ErrInvalidTransition)
	}

	latest, err := s.postRepo.GetLatestAd(ctx, dealID)
	if err != nil {
		return nil, fmt.Errorf("get latest ad: %w", err)
	}

	currentVersion := 1
	if len(latest) > 0 && latest[0].Version != nil {
		currentVersion = *latest[0].Version
	}
	nextVersion := currentVersion + 1

	var posts []entity.Post
	if err := s.tx.WithTx(ctx, func(txCtx context.Context) error {
		var txErr error
		posts, txErr = s.postRepo.AddAdVersion(txCtx, dealID, nextVersion, newPosts)
		if txErr != nil {
			return fmt.Errorf("add ad version: %w", txErr)
		}
		return s.dealRepo.UpdateStatus(txCtx, dealID, entity.DealStatusPendingReview, nil)
	}); err != nil {
		return nil, err
	}

	s.log.Info("revision submitted", "deal_id", dealID, "version", nextVersion)
	return posts, nil
}

func (s *svc) Cancel(ctx context.Context, dealID uuid.UUID) error {
	user, ok := dto.UserFromContext(ctx)
	if !ok {
		return fmt.Errorf("cancel deal: %w", dto.ErrForbidden)
	}

	deal, err := s.dealRepo.GetByID(ctx, dealID)
	if err != nil {
		return fmt.Errorf("get deal: %w", err)
	}

	if deal.AdvertiserID != user.ID {
		return fmt.Errorf("cancel deal: %w", dto.ErrForbidden)
	}

	if !canTransition(deal.Status, entity.DealStatusCancelled) {
		return fmt.Errorf("cancel deal: %w", dto.ErrInvalidTransition)
	}

	if time.Now().After(deal.ScheduledAt) {
		return fmt.Errorf(
			"cancel deal: %w",
			dto.ErrValidation.WithDetails(
				map[string]any{"scheduled_at": "posting time has already arrived"},
			),
		)
	}

	if err := s.dealRepo.UpdateStatus(ctx, dealID, entity.DealStatusCancelled, nil); err != nil {
		return fmt.Errorf("cancel deal: %w", err)
	}

	s.log.Info("deal cancelled", "deal_id", dealID)
	return nil
}

func (s *svc) requirePublisherRole(ctx context.Context, dealID uuid.UUID) (*entity.Deal, error) {
	user, ok := dto.UserFromContext(ctx)
	if !ok {
		return nil, fmt.Errorf("check publisher role: %w", dto.ErrForbidden)
	}

	deal, err := s.dealRepo.GetByID(ctx, dealID)
	if err != nil {
		return nil, fmt.Errorf("get deal: %w", err)
	}

	_, err = s.channelRepo.GetRole(ctx, deal.ChannelID, user.ID)
	if errors.Is(err, dto.ErrNotFound) {
		return nil, fmt.Errorf("check publisher role: %w", dto.ErrForbidden)
	}
	if err != nil {
		return nil, fmt.Errorf("get role: %w", err)
	}

	return deal, nil
}
