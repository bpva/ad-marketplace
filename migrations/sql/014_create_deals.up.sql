CREATE TABLE deals (
    id UUID PRIMARY KEY,
    channel_id UUID NOT NULL REFERENCES channels(id),
    advertiser_id UUID NOT NULL REFERENCES users(id),
    status TEXT NOT NULL DEFAULT 'pending_payment',
    scheduled_at TIMESTAMPTZ NOT NULL,
    publisher_note TEXT,
    escrow_wallet_address TEXT,
    advertiser_wallet_address TEXT,
    payout_wallet_address TEXT,
    format_type TEXT NOT NULL,
    is_native BOOLEAN NOT NULL DEFAULT FALSE,
    feed_hours INTEGER NOT NULL,
    top_hours INTEGER NOT NULL,
    price_nano_ton BIGINT NOT NULL,
    posted_message_ids BIGINT[],
    paid_at TIMESTAMPTZ,
    payment_tx_hash TEXT,
    posted_at TIMESTAMPTZ,
    release_tx_hash TEXT,
    refund_tx_hash TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_deals_channel_id ON deals(channel_id);
CREATE INDEX idx_deals_advertiser_id ON deals(advertiser_id);
CREATE INDEX idx_deals_status ON deals(status);
