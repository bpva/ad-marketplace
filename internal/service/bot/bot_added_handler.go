package bot

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	tele "gopkg.in/telebot.v4"

	"github.com/bpva/ad-marketplace/internal/dto"
	"github.com/bpva/ad-marketplace/internal/entity"
)

func (b *svc) handleMyChatMember(c tele.Context) error {
	update := c.ChatMember()
	if update == nil {
		return nil
	}
	return b.HandleChatMemberUpdate(update)
}

func (b *svc) HandleChatMemberUpdate(update *tele.ChatMemberUpdate) error {
	chat := update.Chat
	if chat.Type != tele.ChatChannel {
		return nil
	}

	ctx := context.Background()
	newStatus := update.NewChatMember.Role

	if newStatus == tele.Left || newStatus == tele.Kicked {
		return b.handleBotRemoved(ctx, chat)
	}

	if newStatus == tele.Administrator {
		return b.handleBotAdded(ctx, chat, update.NewChatMember)
	}

	return nil
}

func (b *svc) handleBotAdded(
	ctx context.Context,
	chat *tele.Chat,
	member *tele.ChatMember,
) error {
	if !member.CanPostMessages {
		b.log.Info("bot added without post permission, ignoring",
			"channel_id", chat.ID,
			"title", chat.Title)
		return nil
	}

	var (
		admins []dto.ChannelAdmin
		err    error
	)

	// there's a delay
	for range b.cfg.MaxRetries {
		time.Sleep(b.cfg.RetryDelay)
		admins, err = b.client.AdminsOf(chat.ID)
		if err == nil {
			break
		}
	}
	if err != nil {
		b.log.Error("failed to get channel admins",
			"channel_id", chat.ID,
			"error", err)
		return nil
	}

	var creator *dto.ChannelAdmin
	for i := range admins {
		if admins[i].Role == dto.RoleCreator {
			creator = &admins[i]
			break
		}
	}

	if creator == nil {
		b.log.Warn("no creator found for channel",
			"channel_id", chat.ID,
			"title", chat.Title)
		return nil
	}

	user, err := b.userRepo.GetByTgID(ctx, creator.TgID)
	if errors.Is(err, dto.ErrNotFound) {
		name := creator.FirstName
		if creator.LastName != "" {
			name += " " + creator.LastName
		}
		user, err = b.userRepo.Create(ctx, creator.TgID, name)
		if err != nil {
			b.log.Error("failed to create user",
				"telegram_id", creator.TgID,
				"error", err)
			return nil
		}
		b.log.Info("user created from channel admin",
			"user_id", user.ID,
			"telegram_id", creator.TgID)
	} else if err != nil {
		b.log.Error("failed to get user",
			"telegram_id", creator.TgID,
			"error", err)
		return nil
	}

	var username *string
	if chat.Username != "" {
		username = &chat.Username
	}

	var channelID string
	err = b.tx.WithTx(ctx, func(ctx context.Context) error {
		channel, err := b.channelRepo.Create(ctx, chat.ID, chat.Title, username)
		if err != nil {
			return err
		}
		channelID = channel.ID.String()

		_, err = b.channelRepo.CreateRole(ctx, channel.ID, user.ID, entity.ChannelRoleTypeOwner)
		return err
	})
	if err != nil {
		b.log.Error("failed to create channel",
			"telegram_channel_id", chat.ID,
			"error", err)
		return nil
	}

	b.log.Info("channel registered",
		"channel_id", channelID,
		"telegram_channel_id", chat.ID,
		"title", chat.Title,
		"owner_id", user.ID)

	chUUID, err := uuid.Parse(channelID)
	if err != nil {
		b.log.Error("failed to parse channel ID",
			"channel_id", channelID,
			"error", err)
		return nil
	}

	go func(ctx context.Context) {
		small, big, err := b.client.GetChatPhoto(chat.ID)
		if err != nil {
			b.log.Error("failed to get channel photo",
				"channel_id", channelID,
				"telegram_channel_id", chat.ID,
				"error", err)
		} else if small != "" || big != "" {
			var sp, bp *string
			if small != "" {
				sp = &small
			}
			if big != "" {
				bp = &big
			}
			if err := b.channelRepo.UpdatePhoto(ctx, chUUID, sp, bp); err != nil {
				b.log.Error("failed to update channel photo",
					"channel_id", channelID,
					"error", err)
			}
		}

		if err := b.stats.FetchAndStore(ctx, chUUID, chat.ID); err != nil {
			b.log.Error("failed to fetch channel stats",
				"channel_id", channelID,
				"telegram_channel_id", chat.ID,
				"error", err)
		}
	}(ctx)

	return nil
}

func (b *svc) handleBotRemoved(ctx context.Context, chat *tele.Chat) error {
	err := b.channelRepo.SoftDelete(ctx, chat.ID)
	if errors.Is(err, dto.ErrNotFound) {
		b.log.Info("channel not found for soft delete",
			"telegram_channel_id", chat.ID)
		return nil
	}
	if err != nil {
		b.log.Error("failed to soft delete channel",
			"telegram_channel_id", chat.ID,
			"error", err)
		return nil
	}

	b.log.Info("channel soft deleted",
		"telegram_channel_id", chat.ID,
		"title", chat.Title)

	return nil
}
