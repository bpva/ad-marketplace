package entity

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

type DealStatus string

const (
	// Awaiting advertiser's TON payment to escrow wallet
	DealStatusPendingPayment DealStatus = "pending_payment"
	// Payment never arrived or failed on-chain
	DealStatusHoldFailed DealStatus = "hold_failed"
	// Payment confirmed; publisher must review the ad creative
	DealStatusPendingReview DealStatus = "pending_review"
	// Publisher requested edits; advertiser must revise the ad
	DealStatusChangesRequested DealStatus = "changes_requested"
	// Publisher approved; ad is scheduled, can no longer be cancelled
	DealStatusApproved DealStatus = "approved"
	// Publisher rejected the ad; funds returned to advertiser
	DealStatusRejected DealStatus = "rejected"
	// Advertiser cancelled before approval; funds returned
	DealStatusCancelled DealStatus = "cancelled"
	// Ad posted to channel; verification window active
	DealStatusPosted DealStatus = "posted"
	// Verification passed; funds released to publisher
	DealStatusCompleted DealStatus = "completed"
	// Post deleted/modified during verification; requires resolution
	DealStatusDispute DealStatus = "dispute"
)

func (s *DealStatus) Scan(src any) error {
	switch v := src.(type) {
	case string:
		*s = DealStatus(v)
	case []byte:
		return s.Scan(string(v))
	case nil:
		return nil
	default:
		return fmt.Errorf("cannot scan %T into DealStatus", src)
	}
	return nil
}

type Deal struct {
	ID                      uuid.UUID    `db:"id"`
	ChannelID               uuid.UUID    `db:"channel_id"`
	AdvertiserID            uuid.UUID    `db:"advertiser_id"`
	Status                  DealStatus   `db:"status"`
	ScheduledAt             time.Time    `db:"scheduled_at"`
	PublisherNote           *string      `db:"publisher_note"`
	EscrowWalletAddress     *string      `db:"escrow_wallet_address"`
	AdvertiserWalletAddress *string      `db:"advertiser_wallet_address"`
	PayoutWalletAddress     *string      `db:"payout_wallet_address"`
	FormatType              AdFormatType `db:"format_type"`
	IsNative                bool         `db:"is_native"`
	FeedHours               int          `db:"feed_hours"`
	TopHours                int          `db:"top_hours"`
	PriceNanoTON            int64        `db:"price_nano_ton"`
	PostedMessageIDs        []int64      `db:"posted_message_ids"`
	PaidAt                  *time.Time   `db:"paid_at"`
	PaymentTxHash           *string      `db:"payment_tx_hash"`
	PostedAt                *time.Time   `db:"posted_at"`
	ReleaseTxHash           *string      `db:"release_tx_hash"`
	RefundTxHash            *string      `db:"refund_tx_hash"`
	CreatedAt               time.Time    `db:"created_at"`
	UpdatedAt               time.Time    `db:"updated_at"`
}
