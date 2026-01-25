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
	GetRole(ctx context.Context, channelID, userID uuid.UUID) (*entity.ChannelRole, error)
	GetRolesByChannelID(ctx context.Context, channelID uuid.UUID) ([]entity.ChannelRole, error)
	CreateRole(
		ctx context.Context, channelID, userID uuid.UUID, role entity.ChannelRoleType,
	) (*entity.ChannelRole, error)
	DeleteRole(ctx context.Context, channelID, userID uuid.UUID) error
}

type UserRepository interface {
	GetByTgID(ctx context.Context, tgID int64) (*entity.User, error)
	GetByID(ctx context.Context, id uuid.UUID) (*entity.User, error)
	Create(ctx context.Context, tgID int64, name string) (*entity.User, error)
}

type TelebotClient interface {
	AdminsOf(channelID int64) ([]dto.ChannelAdmin, error)
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

func (s *svc) GetUserChannels(ctx context.Context) ([]dto.ChannelWithRoleResponse, error) {
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
		result = append(result, dto.ChannelWithRoleResponse{
			Channel: channelToResponse(&channels[i]),
			Role:    string(role.Role),
		})
	}

	return result, nil
}

func (s *svc) GetChannel(
	ctx context.Context, tgChannelID int64,
) (*dto.ChannelResponse, error) {
	channel, err := s.getChannelEntity(ctx, tgChannelID)
	if err != nil {
		return nil, err
	}

	resp := channelToResponse(channel)
	return &resp, nil
}

func (s *svc) GetChannelAdmins(
	ctx context.Context, tgChannelID int64,
) ([]dto.ChannelAdmin, error) {
	_, err := s.getChannelEntity(ctx, tgChannelID)
	if err != nil {
		return nil, err
	}

	admins, err := s.bot.AdminsOf(tgChannelID)
	if err != nil {
		return nil, fmt.Errorf("get admins: %w", err)
	}

	return admins, nil
}

func (s *svc) GetChannelManagers(
	ctx context.Context, tgChannelID int64,
) ([]dto.ManagerResponse, error) {
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
			Role:      string(role.Role),
			CreatedAt: role.CreatedAt,
		})
	}

	return result, nil
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

func channelToResponse(ch *entity.Channel) dto.ChannelResponse {
	resp := dto.ChannelResponse{
		TgChannelID: ch.TgChannelID,
		Title:       ch.Title,
	}
	if ch.Username != nil {
		resp.Username = *ch.Username
	}
	return resp
}
