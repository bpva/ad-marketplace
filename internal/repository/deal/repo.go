package deal

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/bpva/ad-marketplace/internal/dto"
	"github.com/bpva/ad-marketplace/internal/entity"
)

type db interface {
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
}

type repo struct {
	db db
}

func New(db db) *repo {
	return &repo{db: db}
}

const dealColumns = `
	id, channel_id, advertiser_id, status, scheduled_at,
	publisher_note, escrow_wallet_address, advertiser_wallet_address,
	payout_wallet_address, format_type, is_native, feed_hours,
	top_hours, price_nano_ton, posted_message_ids,
	paid_at, payment_tx_hash, posted_at, release_tx_hash,
	refund_tx_hash, created_at, updated_at
`

func (r *repo) Create(ctx context.Context, deal *entity.Deal) (*entity.Deal, error) {
	id, err := uuid.NewV7()
	if err != nil {
		return nil, fmt.Errorf("creating deal: %w", err)
	}

	rows, err := r.db.Query(ctx, `
		INSERT INTO deals (
			id, channel_id, advertiser_id, status, scheduled_at,
			publisher_note, escrow_wallet_address, advertiser_wallet_address,
			payout_wallet_address, format_type, is_native, feed_hours,
			top_hours, price_nano_ton
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
		RETURNING `+dealColumns,
		id, deal.ChannelID, deal.AdvertiserID, deal.Status, deal.ScheduledAt,
		deal.PublisherNote, deal.EscrowWalletAddress, deal.AdvertiserWalletAddress,
		deal.PayoutWalletAddress, deal.FormatType, deal.IsNative, deal.FeedHours,
		deal.TopHours, deal.PriceNanoTON)
	if err != nil {
		return nil, fmt.Errorf("creating deal: %w", err)
	}

	d, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[entity.Deal])
	if err != nil {
		return nil, fmt.Errorf("creating deal: %w", err)
	}

	return &d, nil
}

func (r *repo) GetByID(ctx context.Context, id uuid.UUID) (*entity.Deal, error) {
	rows, err := r.db.Query(ctx, `SELECT `+dealColumns+` FROM deals WHERE id = $1`, id)
	if err != nil {
		return nil, fmt.Errorf("getting deal by id: %w", err)
	}

	d, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[entity.Deal])
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, fmt.Errorf("getting deal by id: %w", dto.ErrNotFound)
	}
	if err != nil {
		return nil, fmt.Errorf("getting deal by id: %w", err)
	}

	return &d, nil
}

func (r *repo) GetByChannelID(
	ctx context.Context, channelID uuid.UUID, limit, offset int,
) ([]entity.Deal, int, error) {
	countRows, err := r.db.Query(ctx,
		`SELECT COUNT(*) FROM deals WHERE channel_id = $1`, channelID)
	if err != nil {
		return nil, 0, fmt.Errorf("counting deals by channel id: %w", err)
	}

	total, err := pgx.CollectOneRow(countRows, pgx.RowTo[int])
	if err != nil {
		return nil, 0, fmt.Errorf("counting deals by channel id: %w", err)
	}

	rows, err := r.db.Query(ctx, `
		SELECT `+dealColumns+`
		FROM deals
		WHERE channel_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`, channelID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("getting deals by channel id: %w", err)
	}

	deals, err := pgx.CollectRows(rows, pgx.RowToStructByName[entity.Deal])
	if err != nil {
		return nil, 0, fmt.Errorf("getting deals by channel id: %w", err)
	}

	return deals, total, nil
}

func (r *repo) GetByAdvertiserID(
	ctx context.Context, advertiserID uuid.UUID, limit, offset int,
) ([]entity.Deal, int, error) {
	countRows, err := r.db.Query(ctx,
		`SELECT COUNT(*) FROM deals WHERE advertiser_id = $1`, advertiserID)
	if err != nil {
		return nil, 0, fmt.Errorf("counting deals by advertiser id: %w", err)
	}

	total, err := pgx.CollectOneRow(countRows, pgx.RowTo[int])
	if err != nil {
		return nil, 0, fmt.Errorf("counting deals by advertiser id: %w", err)
	}

	rows, err := r.db.Query(ctx, `
		SELECT `+dealColumns+`
		FROM deals
		WHERE advertiser_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`, advertiserID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("getting deals by advertiser id: %w", err)
	}

	deals, err := pgx.CollectRows(rows, pgx.RowToStructByName[entity.Deal])
	if err != nil {
		return nil, 0, fmt.Errorf("getting deals by advertiser id: %w", err)
	}

	return deals, total, nil
}

func (r *repo) UpdateStatus(
	ctx context.Context, id uuid.UUID, status entity.DealStatus, note *string,
) error {
	tag, err := r.db.Exec(ctx, `
		UPDATE deals
		SET status = $2, publisher_note = $3, updated_at = NOW()
		WHERE id = $1
	`, id, status, note)
	if err != nil {
		return fmt.Errorf("updating deal status: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("updating deal status: %w", dto.ErrNotFound)
	}
	return nil
}

func (r *repo) SetPostedMessageIDs(ctx context.Context, id uuid.UUID, messageIDs []int64) error {
	tag, err := r.db.Exec(ctx, `
		UPDATE deals
		SET posted_message_ids = $2, updated_at = NOW()
		WHERE id = $1
	`, id, messageIDs)
	if err != nil {
		return fmt.Errorf("setting posted message ids: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("setting posted message ids: %w", dto.ErrNotFound)
	}
	return nil
}
