package channel

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	petname "github.com/dustinkirkland/golang-petname"
	"github.com/google/uuid"

	"github.com/bpva/ad-marketplace/internal/dto"
	"github.com/bpva/ad-marketplace/internal/entity"
	"github.com/bpva/ad-marketplace/internal/logx"
)

type ChannelRepository interface {
	GetByID(ctx context.Context, id uuid.UUID) (*entity.Channel, error)
	GetByTgChannelID(ctx context.Context, tgChannelID int64) (*entity.Channel, error)
	GetChannelsByUserID(ctx context.Context, userID uuid.UUID) ([]entity.Channel, error)
	GetChannels(
		ctx context.Context,
		filters []entity.Filter,
		sort entity.ChannelSort,
		limit, offset int,
	) ([]entity.MVChannel, int, error)
	GetRole(ctx context.Context, channelID, userID uuid.UUID) (*entity.ChannelRole, error)
	GetRolesByChannelID(ctx context.Context, channelID uuid.UUID) ([]entity.ChannelRole, error)
	CreateRole(
		ctx context.Context, channelID, userID uuid.UUID, role entity.ChannelRoleType,
	) (*entity.ChannelRole, error)
	DeleteRole(ctx context.Context, channelID, userID uuid.UUID) error
	UpdateListing(ctx context.Context, channelID uuid.UUID, isListed bool) error
	CreateAdFormat(
		ctx context.Context,
		channelID uuid.UUID,
		formatType entity.AdFormatType,
		isNative bool,
		feedHours, topHours int,
		priceNanoTON int64,
	) (*entity.ChannelAdFormat, error)
	GetAdFormatByID(ctx context.Context, formatID uuid.UUID) (*entity.ChannelAdFormat, error)
	GetAdFormatsByChannelID(
		ctx context.Context,
		channelID uuid.UUID,
	) ([]entity.ChannelAdFormat, error)
	DeleteAdFormat(ctx context.Context, formatID uuid.UUID) error
	SetCategories(ctx context.Context, channelID uuid.UUID, categorySlugs []string) error
	GetCategoriesByChannelID(ctx context.Context, channelID uuid.UUID) ([]entity.Category, error)
	GetInfo(ctx context.Context, channelID uuid.UUID) (*entity.ChannelInfo, error)
	HasRecentStats(ctx context.Context, channelID uuid.UUID) (bool, error)
	GetOwnerWalletAddress(ctx context.Context, channelID uuid.UUID) (*string, error)
	RefreshMV(ctx context.Context) error
}

type UserRepository interface {
	GetByTgID(ctx context.Context, tgID int64) (*entity.User, error)
	GetByID(ctx context.Context, id uuid.UUID) (*entity.User, error)
	Create(ctx context.Context, tgID int64, name string) (*entity.User, error)
}

type TelebotClient interface {
	AdminsOf(channelID int64) ([]dto.ChannelAdmin, error)
	DownloadFile(fileID string) ([]byte, error)
}

type Transactor interface {
	WithTx(ctx context.Context, f func(ctx context.Context) error) error
}

type svc struct {
	channelRepo ChannelRepository
	userRepo    UserRepository
	bot         TelebotClient
	tx          Transactor
	log         *slog.Logger
}

func New(
	channelRepo ChannelRepository,
	userRepo UserRepository,
	bot TelebotClient,
	tx Transactor,
	log *slog.Logger,
) *svc {
	log = log.With(logx.Service("ChannelService"))
	return &svc{
		channelRepo: channelRepo,
		userRepo:    userRepo,
		bot:         bot,
		tx:          tx,
		log:         log,
	}
}

func (s *svc) GetUserChannels(ctx context.Context) (*dto.ChannelsResponse, error) {
	user, ok := dto.UserFromContext(ctx)
	if !ok {
		return nil, fmt.Errorf("get user channels: %w", dto.ErrForbidden)
	}

	channels, err := s.channelRepo.GetChannelsByUserID(ctx, user.ID)
	if err != nil {
		return nil, fmt.Errorf("get channels: %w", err)
	}

	result := make([]dto.ChannelWithRoleResponse, 0, len(channels))
	for i := range channels {
		role, err := s.channelRepo.GetRole(ctx, channels[i].ID, user.ID)
		if err != nil {
			return nil, fmt.Errorf("get role: %w", err)
		}
		resp := s.channelToResponse(ctx, &channels[i])
		walletAddr, err := s.channelRepo.GetOwnerWalletAddress(ctx, channels[i].ID)
		if err != nil && !errors.Is(err, dto.ErrNotFound) {
			return nil, fmt.Errorf("get owner wallet address: %w", err)
		}
		resp.PayoutAddress = walletAddr
		result = append(result, dto.ChannelWithRoleResponse{
			Channel: resp,
			Role:    role.Role,
		})
	}

	return &dto.ChannelsResponse{Channels: result}, nil
}

func (s *svc) GetChannel(
	ctx context.Context, tgChannelID int64,
) (*dto.ChannelResponse, error) {
	channel, err := s.getChannelEntity(ctx, tgChannelID)
	if err != nil {
		return nil, err
	}

	resp := s.channelToResponse(ctx, channel)
	walletAddr, err := s.channelRepo.GetOwnerWalletAddress(ctx, channel.ID)
	if err != nil && !errors.Is(err, dto.ErrNotFound) {
		return nil, fmt.Errorf("get owner wallet address: %w", err)
	}
	resp.PayoutAddress = walletAddr
	return &resp, nil
}

func (s *svc) GetChannelAdmins(
	ctx context.Context, tgChannelID int64,
) (*dto.ChannelAdminsResponse, error) {
	user, ok := dto.UserFromContext(ctx)
	if !ok {
		return nil, fmt.Errorf("get channel admins: %w", dto.ErrForbidden)
	}

	channel, err := s.getChannelEntity(ctx, tgChannelID)
	if err != nil {
		return nil, err
	}

	admins, err := s.bot.AdminsOf(tgChannelID)
	if err != nil {
		return nil, fmt.Errorf("get admins: %w", err)
	}

	roles, err := s.channelRepo.GetRolesByChannelID(ctx, channel.ID)
	if err != nil {
		return nil, fmt.Errorf("get roles: %w", err)
	}

	exclude := make(map[int64]bool)
	exclude[user.TgID] = true
	for _, role := range roles {
		u, err := s.userRepo.GetByID(ctx, role.UserID)
		if err != nil {
			continue
		}
		exclude[u.TgID] = true
	}

	result := make([]dto.ChannelAdmin, 0, len(admins))
	for _, admin := range admins {
		if !exclude[admin.TgID] {
			result = append(result, admin)
		}
	}

	return &dto.ChannelAdminsResponse{Admins: result}, nil
}

func (s *svc) GetChannelManagers(
	ctx context.Context, tgChannelID int64,
) (*dto.ChannelManagersResponse, error) {
	channel, err := s.getChannelEntity(ctx, tgChannelID)
	if err != nil {
		return nil, err
	}

	roles, err := s.channelRepo.GetRolesByChannelID(ctx, channel.ID)
	if err != nil {
		return nil, fmt.Errorf("get roles: %w", err)
	}

	result := make([]dto.ManagerResponse, 0, len(roles))
	for _, role := range roles {
		u, err := s.userRepo.GetByID(ctx, role.UserID)
		if err != nil {
			s.log.Warn("failed to get user for role", "user_id", role.UserID, "error", err)
			continue
		}
		result = append(result, dto.ManagerResponse{
			TgID:      u.TgID,
			Name:      u.Name,
			Role:      role.Role,
			CreatedAt: role.CreatedAt,
		})
	}

	return &dto.ChannelManagersResponse{Managers: result}, nil
}

func (s *svc) AddManager(
	ctx context.Context, tgChannelID int64, tgID int64,
) error {
	user, ok := dto.UserFromContext(ctx)
	if !ok {
		return fmt.Errorf("add manager: %w", dto.ErrForbidden)
	}

	channel, err := s.channelRepo.GetByTgChannelID(ctx, tgChannelID)
	if err != nil {
		return fmt.Errorf("get channel: %w", err)
	}

	actorRole, err := s.channelRepo.GetRole(ctx, channel.ID, user.ID)
	if errors.Is(err, dto.ErrNotFound) {
		return fmt.Errorf("add manager: %w", dto.ErrForbidden)
	}
	if err != nil {
		return fmt.Errorf("get actor role: %w", err)
	}
	if actorRole.Role != entity.ChannelRoleTypeOwner {
		return fmt.Errorf("add manager: %w", dto.ErrForbidden)
	}

	return s.tx.WithTx(ctx, func(ctx context.Context) error {
		target, err := s.userRepo.GetByTgID(ctx, tgID)
		if errors.Is(err, dto.ErrNotFound) {
			name := petname.Generate(2, " ")
			target, err = s.userRepo.Create(ctx, tgID, name)
			if err != nil {
				return fmt.Errorf("create user: %w", err)
			}
			s.log.Info("user created for manager role", "telegram_id", tgID, "name", name)
		} else if err != nil {
			return fmt.Errorf("get user: %w", err)
		}

		_, err = s.channelRepo.CreateRole(ctx, channel.ID, target.ID, entity.ChannelRoleTypeManager)
		if err != nil {
			return fmt.Errorf("create role: %w", err)
		}

		s.log.Info("manager added",
			"channel_id", channel.ID, "user_id", target.ID, "added_by", user.ID)
		return nil
	})
}

func (s *svc) RemoveManager(ctx context.Context, tgChannelID int64, tgID int64) error {
	user, ok := dto.UserFromContext(ctx)
	if !ok {
		return fmt.Errorf("remove manager: %w", dto.ErrForbidden)
	}

	channel, err := s.channelRepo.GetByTgChannelID(ctx, tgChannelID)
	if err != nil {
		return fmt.Errorf("get channel: %w", err)
	}

	actorRole, err := s.channelRepo.GetRole(ctx, channel.ID, user.ID)
	if errors.Is(err, dto.ErrNotFound) {
		return fmt.Errorf("remove manager: %w", dto.ErrForbidden)
	}
	if err != nil {
		return fmt.Errorf("get actor role: %w", err)
	}
	if actorRole.Role != entity.ChannelRoleTypeOwner {
		return fmt.Errorf("remove manager: %w", dto.ErrForbidden)
	}

	target, err := s.userRepo.GetByTgID(ctx, tgID)
	if errors.Is(err, dto.ErrNotFound) {
		return fmt.Errorf("remove manager: %w", dto.ErrNotFound)
	}
	if err != nil {
		return fmt.Errorf("get user: %w", err)
	}

	managerRole, err := s.channelRepo.GetRole(ctx, channel.ID, target.ID)
	if err != nil {
		return fmt.Errorf("get manager role: %w", err)
	}
	if managerRole.Role == entity.ChannelRoleTypeOwner {
		return fmt.Errorf("remove manager: %w", dto.ErrCannotRemoveOwner)
	}

	if err := s.channelRepo.DeleteRole(ctx, channel.ID, target.ID); err != nil {
		return fmt.Errorf("delete role: %w", err)
	}

	s.log.Info("manager removed",
		"channel_id", channel.ID, "user_id", target.ID, "removed_by", user.ID)
	return nil
}

const marketplacePageSize = 10

func (s *svc) GetMarketplaceChannels(
	ctx context.Context, req dto.MarketplaceChannelsRequest,
) (*dto.MarketplaceChannelsResponse, error) {
	var filters []entity.Filter
	filters = append(filters, entity.Filter{Name: "has_ad_formats"})
	for _, f := range req.Filters {
		filters = append(filters, entity.Filter{Name: f.Name, Value: f.Value})
	}

	sortBy := req.SortBy
	if sortBy == "" {
		sortBy = entity.ChannelSortBySubscribers
	}
	sortOrder := req.SortOrder
	if sortOrder == "" {
		sortOrder = entity.SortOrderDesc
	}
	sort := entity.ChannelSort{By: sortBy, Order: sortOrder}

	page := req.Page
	if page < 1 {
		page = 1
	}
	offset := (page - 1) * marketplacePageSize

	channels, total, err := s.channelRepo.GetChannels(
		ctx,
		filters,
		sort,
		marketplacePageSize,
		offset,
	)
	if err != nil {
		return nil, fmt.Errorf("get marketplace channels: %w", err)
	}

	result := make([]dto.MarketplaceChannel, 0, len(channels))
	for _, ch := range channels {
		mc := dto.MarketplaceChannel{
			TgChannelID:             ch.TgChannelID,
			Title:                   ch.Title,
			About:                   ch.About,
			Subscribers:             ch.Subscribers,
			Languages:               ch.Languages,
			TopHours:                ch.TopHours,
			ReactionsByEmotion:      ch.ReactionsByEmotion,
			StoryReactionsByEmotion: ch.StoryReactionsByEmotion,
			AvgDailyViews1d:         ch.AvgDailyViews1d,
			AvgDailyViews7d:         ch.AvgDailyViews7d,
			AvgDailyViews30d:        ch.AvgDailyViews30d,
			TotalViews7d:            ch.TotalViews7d,
			TotalViews30d:           ch.TotalViews30d,
			SubGrowth7d:             ch.SubGrowth7d,
			SubGrowth30d:            ch.SubGrowth30d,
			AvgInteractions7d:       ch.AvgInteractions7d,
			AvgInteractions30d:      ch.AvgInteractions30d,
			EngagementRate7d:        ch.EngagementRate7d,
			EngagementRate30d:       ch.EngagementRate30d,
		}

		formats := make([]dto.AdFormat, 0, len(ch.AdFormats))
		for i := range ch.AdFormats {
			f := &ch.AdFormats[i]
			formats = append(formats, dto.AdFormat{
				ID:           f.ID.String(),
				FormatType:   f.FormatType,
				IsNative:     f.IsNative,
				FeedHours:    f.FeedHours,
				TopHours:     f.TopHours,
				PriceNanoTON: f.PriceNanoTON,
			})
		}
		mc.AdFormats = formats
		mc.Categories = categoriesToResponse(ch.Categories)

		if ch.Username != nil {
			mc.Username = *ch.Username
		}
		if ch.PhotoSmallFileID != nil {
			mc.PhotoSmallURL = fmt.Sprintf(
				"/api/v1/channels/%d/photo?size=small",
				ch.TgChannelID,
			)
		}
		result = append(result, mc)
	}

	return &dto.MarketplaceChannelsResponse{
		Channels: result,
		Total:    total,
	}, nil
}

func (s *svc) UpdateCategories(
	ctx context.Context, tgChannelID int64, categories []string,
) error {
	if len(categories) > 3 {
		return fmt.Errorf("update categories: %w", dto.ErrTooManyCategories)
	}

	seen := make(map[string]struct{}, len(categories))
	for _, slug := range categories {
		if _, ok := entity.AllCategories[entity.ChannelCategory(slug)]; !ok {
			return fmt.Errorf(
				"update categories: unknown category %q: %w",
				slug,
				dto.ErrInvalidCategory,
			)
		}
		if _, dup := seen[slug]; dup {
			return fmt.Errorf(
				"update categories: duplicate category %q: %w",
				slug,
				dto.ErrInvalidCategory,
			)
		}
		seen[slug] = struct{}{}
	}

	channel, err := s.getChannelEntityAsOwner(ctx, tgChannelID)
	if err != nil {
		return err
	}

	if err := s.channelRepo.SetCategories(ctx, channel.ID, categories); err != nil {
		return fmt.Errorf("update categories: %w", err)
	}

	go func() {
		if err := s.channelRepo.RefreshMV(context.Background()); err != nil {
			s.log.Warn("refresh marketplace mv", "error", err)
		}
	}()

	s.log.Info("channel categories updated", "channel_id", channel.ID, "categories", categories)
	return nil
}

func (s *svc) getChannelEntity(
	ctx context.Context, tgChannelID int64,
) (*entity.Channel, error) {
	user, ok := dto.UserFromContext(ctx)
	if !ok {
		return nil, fmt.Errorf("get channel: %w", dto.ErrForbidden)
	}

	channel, err := s.channelRepo.GetByTgChannelID(ctx, tgChannelID)
	if err != nil {
		return nil, fmt.Errorf("get channel: %w", err)
	}

	_, err = s.channelRepo.GetRole(ctx, channel.ID, user.ID)
	if errors.Is(err, dto.ErrNotFound) {
		return nil, fmt.Errorf("get channel: %w", dto.ErrForbidden)
	}
	if err != nil {
		return nil, fmt.Errorf("get role: %w", err)
	}

	return channel, nil
}

func (s *svc) getChannelEntityAsOwner(
	ctx context.Context, tgChannelID int64,
) (*entity.Channel, error) {
	user, ok := dto.UserFromContext(ctx)
	if !ok {
		return nil, fmt.Errorf("get channel: %w", dto.ErrForbidden)
	}

	channel, err := s.channelRepo.GetByTgChannelID(ctx, tgChannelID)
	if err != nil {
		return nil, fmt.Errorf("get channel: %w", err)
	}

	role, err := s.channelRepo.GetRole(ctx, channel.ID, user.ID)
	if errors.Is(err, dto.ErrNotFound) {
		return nil, fmt.Errorf("get channel: %w", dto.ErrForbidden)
	}
	if err != nil {
		return nil, fmt.Errorf("get role: %w", err)
	}
	if role.Role != entity.ChannelRoleTypeOwner {
		return nil, fmt.Errorf("get channel: %w", dto.ErrForbidden)
	}

	return channel, nil
}

func (s *svc) UpdateListing(ctx context.Context, tgChannelID int64, isListed bool) error {
	channel, err := s.getChannelEntityAsOwner(ctx, tgChannelID)
	if err != nil {
		return err
	}

	if isListed {
		walletAddr, err := s.channelRepo.GetOwnerWalletAddress(ctx, channel.ID)
		if err != nil {
			return fmt.Errorf("update listing: %w", err)
		}
		if walletAddr == nil {
			return fmt.Errorf("update listing: %w", dto.ErrNoPayoutMethod)
		}
	}

	if err := s.channelRepo.UpdateListing(ctx, channel.ID, isListed); err != nil {
		return fmt.Errorf("update listing: %w", err)
	}

	go func() {
		if err := s.channelRepo.RefreshMV(context.Background()); err != nil {
			s.log.Warn("refresh marketplace mv", "error", err)
		}
	}()

	s.log.Info("channel listing updated", "channel_id", channel.ID, "is_listed", isListed)
	return nil
}

func (s *svc) GetAdFormats(ctx context.Context, tgChannelID int64) (*dto.AdFormatsResponse, error) {
	channel, err := s.getChannelEntity(ctx, tgChannelID)
	if err != nil {
		return nil, err
	}

	formats, err := s.channelRepo.GetAdFormatsByChannelID(ctx, channel.ID)
	if err != nil {
		return nil, fmt.Errorf("get ad formats: %w", err)
	}

	result := make([]dto.AdFormatResponse, 0, len(formats))
	for i := range formats {
		result = append(result, adFormatToResponse(&formats[i]))
	}

	return &dto.AdFormatsResponse{AdFormats: result}, nil
}

func (s *svc) AddAdFormat(
	ctx context.Context,
	tgChannelID int64,
	req dto.AddAdFormatRequest,
) error {
	if req.FormatType != entity.AdFormatTypePost {
		return fmt.Errorf("add ad format: %w", dto.ErrFormatTypeNotAllowed)
	}

	channel, err := s.getChannelEntityAsOwner(ctx, tgChannelID)
	if err != nil {
		return err
	}

	_, err = s.channelRepo.CreateAdFormat(
		ctx,
		channel.ID,
		req.FormatType,
		req.IsNative,
		req.FeedHours,
		req.TopHours,
		req.PriceNanoTON,
	)
	if err != nil {
		return fmt.Errorf("add ad format: %w", err)
	}

	go func() {
		if err := s.channelRepo.RefreshMV(context.Background()); err != nil {
			s.log.Warn("refresh marketplace mv", "error", err)
		}
	}()

	s.log.Info("ad format added", "channel_id", channel.ID, "format_type", req.FormatType)
	return nil
}

func (s *svc) RemoveAdFormat(ctx context.Context, tgChannelID int64, formatID uuid.UUID) error {
	channel, err := s.getChannelEntityAsOwner(ctx, tgChannelID)
	if err != nil {
		return err
	}

	format, err := s.channelRepo.GetAdFormatByID(ctx, formatID)
	if err != nil {
		return fmt.Errorf("remove ad format: %w", err)
	}

	if format.ChannelID != channel.ID {
		return fmt.Errorf("remove ad format: %w", dto.ErrForbidden)
	}

	if err := s.channelRepo.DeleteAdFormat(ctx, formatID); err != nil {
		return fmt.Errorf("remove ad format: %w", err)
	}

	go func() {
		if err := s.channelRepo.RefreshMV(context.Background()); err != nil {
			s.log.Warn("refresh marketplace mv", "error", err)
		}
	}()

	s.log.Info("ad format removed", "channel_id", channel.ID, "format_id", formatID)
	return nil
}

func (s *svc) channelToResponse(ctx context.Context, ch *entity.Channel) dto.ChannelResponse {
	resp := dto.ChannelResponse{
		TgChannelID: ch.TgChannelID,
		Title:       ch.Title,
		IsListed:    ch.IsListed,
		AdFormats:   []dto.AdFormatResponse{},
		Categories:  []dto.CategoryResponse{},
	}
	if ch.Username != nil {
		resp.Username = *ch.Username
	}
	if ch.PhotoSmallFileID != nil {
		resp.PhotoSmallURL = fmt.Sprintf("/api/v1/channels/%d/photo?size=small", ch.TgChannelID)
	}
	if ch.PhotoBigFileID != nil {
		resp.PhotoBigURL = fmt.Sprintf("/api/v1/channels/%d/photo?size=big", ch.TgChannelID)
	}
	info, err := s.channelRepo.GetInfo(ctx, ch.ID)
	if err == nil {
		resp.Subscribers = &info.Subscribers
	}
	formats, err := s.channelRepo.GetAdFormatsByChannelID(ctx, ch.ID)
	if err == nil {
		for i := range formats {
			resp.AdFormats = append(resp.AdFormats, adFormatToResponse(&formats[i]))
		}
	}
	categories, err := s.channelRepo.GetCategoriesByChannelID(ctx, ch.ID)
	if err == nil {
		resp.Categories = categoriesToResponse(categories)
	}
	resp.HasStats, _ = s.channelRepo.HasRecentStats(ctx, ch.ID)
	return resp
}

func (s *svc) GetChannelPhoto(
	ctx context.Context, tgChannelID int64, size string,
) ([]byte, error) {
	channel, err := s.channelRepo.GetByTgChannelID(ctx, tgChannelID)
	if err != nil {
		return nil, fmt.Errorf("get channel photo: %w", err)
	}

	var fileID *string
	switch size {
	case "big":
		fileID = channel.PhotoBigFileID
	default:
		fileID = channel.PhotoSmallFileID
	}
	if fileID == nil {
		return nil, fmt.Errorf("get channel photo: %w", dto.ErrNotFound)
	}

	data, err := s.bot.DownloadFile(*fileID)
	if err != nil {
		return nil, fmt.Errorf("download channel photo: %w", err)
	}

	return data, nil
}

func adFormatToResponse(f *entity.ChannelAdFormat) dto.AdFormatResponse {
	return dto.AdFormatResponse{
		ID:           f.ID.String(),
		FormatType:   f.FormatType,
		IsNative:     f.IsNative,
		FeedHours:    f.FeedHours,
		TopHours:     f.TopHours,
		PriceNanoTON: f.PriceNanoTON,
	}
}

func categoriesToResponse(cats []entity.Category) []dto.CategoryResponse {
	result := make([]dto.CategoryResponse, 0, len(cats))
	for i := range cats {
		result = append(result, dto.CategoryResponse{
			Slug:        cats[i].Slug,
			DisplayName: cats[i].DisplayName,
		})
	}
	return result
}
