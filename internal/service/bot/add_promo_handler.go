package bot

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	petname "github.com/dustinkirkland/golang-petname"
	tele "gopkg.in/telebot.v4"

	"github.com/bpva/ad-marketplace/internal/dto"
	"github.com/bpva/ad-marketplace/internal/entity"
)

type pendingGroup struct {
	timer *time.Timer
	mu    sync.Mutex
}

func (b *svc) handleAddPromo(c tele.Context) error {
	name := strings.TrimSpace(strings.TrimPrefix(c.Message().Text, "/add_promo"))
	if name == "" {
		name = petname.Generate(2, " ")
	}
	b.awaitingPost.Store(c.Sender().ID, name)
	return c.Send(fmt.Sprintf("Send me a post and I'll save it under name \"%s\".", name))
}

func (b *svc) handleIncomingMessage(c tele.Context) error {
	msg := c.Message()
	if msg == nil {
		return nil
	}

	saved, confusing := b.processMessage(msg)
	if confusing {
		return c.Send("confusing...")
	}
	if !saved {
		return nil
	}

	if msg.AlbumID != "" {
		b.debounceGroupConfirmation(c, msg.AlbumID)
		return nil
	}

	b.awaitingPost.Delete(c.Sender().ID)
	return c.Send("Post saved!")
}

func (b *svc) processMessage(msg *tele.Message) (saved, confusing bool) {
	mediaType, mediaFileID := extractMedia(msg)
	isMedia := mediaType != nil

	val, ok := b.awaitingPost.Load(msg.Sender.ID)
	if !ok {
		return false, !isMedia
	}
	name := val.(string)

	var text *string
	var entities []byte

	if isMedia {
		if msg.Caption != "" {
			text = &msg.Caption
		}
		if len(msg.CaptionEntities) > 0 {
			data, err := json.Marshal(msg.CaptionEntities)
			if err != nil {
				b.log.Error("failed to marshal caption entities", "error", err)
				return false, false
			}
			entities = data
		}
	} else {
		if msg.Text != "" {
			text = &msg.Text
		}
		if len(msg.Entities) > 0 {
			data, err := json.Marshal(msg.Entities)
			if err != nil {
				b.log.Error("failed to marshal entities", "error", err)
				return false, false
			}
			entities = data
		}
	}

	ctx := context.Background()

	user, err := b.getOrCreateUser(ctx, msg.Sender)
	if err != nil {
		b.log.Error("failed to get or create user", "error", err)
		return false, false
	}

	var mediaGroupID *string
	if msg.AlbumID != "" {
		mediaGroupID = &msg.AlbumID
	}

	_, err = b.postRepo.Create(
		ctx,
		user.ID,
		&name,
		mediaGroupID,
		text,
		entities,
		mediaType,
		mediaFileID,
		msg.HasMediaSpoiler,
		msg.CaptionAbove,
	)
	if err != nil {
		b.log.Error("failed to create post", "error", err)
		return false, false
	}

	return true, false
}

// HandleUpdate processes a raw update for integration testing.
func (b *svc) HandleUpdate(upd tele.Update) {
	if upd.MyChatMember != nil {
		if err := b.HandleChatMemberUpdate(upd.MyChatMember); err != nil {
			b.log.Error("failed to handle chat member update", "error", err)
		}
		return
	}
	if upd.Message == nil || upd.Message.Sender == nil {
		return
	}
	msg := upd.Message

	if strings.HasPrefix(msg.Text, "/add_promo") {
		name := strings.TrimSpace(strings.TrimPrefix(msg.Text, "/add_promo"))
		if name == "" {
			name = petname.Generate(2, " ")
		}
		b.awaitingPost.Store(msg.Sender.ID, name)
		return
	}

	saved, _ := b.processMessage(msg)
	if saved && msg.AlbumID == "" {
		b.awaitingPost.Delete(msg.Sender.ID)
	}
	if saved && msg.AlbumID != "" {
		b.pendingGroups.Store(msg.AlbumID, &pendingGroup{})
	}
}

func (b *svc) debounceGroupConfirmation(c tele.Context, groupID string) {
	senderID := c.Sender().ID

	val, _ := b.pendingGroups.LoadOrStore(groupID, &pendingGroup{})
	pg := val.(*pendingGroup)

	pg.mu.Lock()
	defer pg.mu.Unlock()

	if pg.timer != nil {
		pg.timer.Stop()
	}

	pg.timer = time.AfterFunc(time.Second, func() {
		b.pendingGroups.Delete(groupID)
		b.awaitingPost.Delete(senderID)
		if err := c.Send("Post saved!"); err != nil {
			b.log.Error("failed to send album confirmation", "error", err)
		}
	})
}

func (b *svc) getOrCreateUser(ctx context.Context, sender *tele.User) (*entity.User, error) {
	user, err := b.userRepo.GetByTgID(ctx, sender.ID)
	if errors.Is(err, dto.ErrNotFound) {
		name := sender.FirstName
		if sender.LastName != "" {
			name += " " + sender.LastName
		}
		user, err = b.userRepo.Create(ctx, sender.ID, name)
		if err != nil {
			return nil, err
		}
		b.log.Info("user created from add_promo",
			"user_id", user.ID,
			"telegram_id", sender.ID)
	} else if err != nil {
		return nil, err
	}
	return user, nil
}

func extractMedia(msg *tele.Message) (*entity.MediaType, *string) {
	switch {
	case msg.Photo != nil:
		mt := entity.MediaTypePhoto
		return &mt, &msg.Photo.FileID
	case msg.Video != nil:
		mt := entity.MediaTypeVideo
		return &mt, &msg.Video.FileID
	case msg.Document != nil:
		mt := entity.MediaTypeDocument
		return &mt, &msg.Document.FileID
	case msg.Animation != nil:
		mt := entity.MediaTypeAnimation
		return &mt, &msg.Animation.FileID
	case msg.Audio != nil:
		mt := entity.MediaTypeAudio
		return &mt, &msg.Audio.FileID
	case msg.Voice != nil:
		mt := entity.MediaTypeVoice
		return &mt, &msg.Voice.FileID
	case msg.VideoNote != nil:
		mt := entity.MediaTypeVideoNote
		return &mt, &msg.VideoNote.FileID
	case msg.Sticker != nil:
		mt := entity.MediaTypeSticker
		return &mt, &msg.Sticker.FileID
	default:
		return nil, nil
	}
}
