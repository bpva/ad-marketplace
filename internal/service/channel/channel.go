package channel

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"time"

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
	GetAdFormatsByChannelID(
		ctx context.Context,
		channelID uuid.UUID,
	) ([]entity.ChannelAdFormat, error)
	GetAdFormatByID(ctx context.Context, formatID uuid.UUID) (*entity.ChannelAdFormat, error)
	DeleteAdFormat(ctx context.Context, formatID uuid.UUID) error
	GetInfo(ctx context.Context, channelID uuid.UUID) (*entity.ChannelInfo, error)
	GetHistoricalStats(
		ctx context.Context, channelID uuid.UUID, from, to time.Time,
	) ([]entity.ChannelHistoricalStats, error)
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
		info, avgViews := s.fetchChannelStats(ctx, channels[i].ID)
		result = append(result, dto.ChannelWithRoleResponse{
			Channel: channelToResponse(&channels[i], info, avgViews),
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

	info, avgViews := s.fetchChannelStats(ctx, channel.ID)
	resp := channelToResponse(channel, info, avgViews)
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

	if err := s.channelRepo.UpdateListing(ctx, channel.ID, isListed); err != nil {
		return fmt.Errorf("update listing: %w", err)
	}

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

	s.log.Info("ad format removed", "channel_id", channel.ID, "format_id", formatID)
	return nil
}

func channelToResponse(
	ch *entity.Channel,
	info *entity.ChannelInfo,
	avgViews *int,
) dto.ChannelResponse {
	resp := dto.ChannelResponse{
		TgChannelID: ch.TgChannelID,
		Title:       ch.Title,
		IsListed:    ch.IsListed,
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
	if info != nil {
		resp.Subscribers = &info.Subscribers
	}
	resp.AvgViews = avgViews
	return resp
}

func (s *svc) fetchChannelStats(
	ctx context.Context, channelID uuid.UUID,
) (*entity.ChannelInfo, *int) {
	info, err := s.channelRepo.GetInfo(ctx, channelID)
	if err != nil {
		info = nil
	}

	to := time.Now()
	from := to.AddDate(0, 0, -7)
	stats, err := s.channelRepo.GetHistoricalStats(ctx, channelID, from, to)
	if err != nil {
		return info, nil
	}

	avgViews := computeAvgDailyViews(stats)
	return info, avgViews
}

func computeAvgDailyViews(stats []entity.ChannelHistoricalStats) *int {
	if len(stats) < 7 {
		return nil
	}

	var totalViews int
	for _, s := range stats {
		var data map[string]json.RawMessage
		if err := json.Unmarshal(s.Data, &data); err != nil {
			continue
		}
		viewsRaw, ok := data["views_by_source"]
		if !ok {
			continue
		}
		var viewsBySource map[string]int
		if err := json.Unmarshal(viewsRaw, &viewsBySource); err != nil {
			continue
		}
		for _, v := range viewsBySource {
			totalViews += v
		}
	}

	avg := totalViews / len(stats)
	return &avg
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
