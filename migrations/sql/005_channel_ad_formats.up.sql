ALTER TABLE channels ADD COLUMN is_listed BOOLEAN NOT NULL DEFAULT TRUE;

CREATE TABLE channel_ad_formats (
    id UUID PRIMARY KEY,
    channel_id UUID NOT NULL REFERENCES channels(id) ON DELETE CASCADE,
    format_type TEXT NOT NULL,
    is_native BOOLEAN NOT NULL DEFAULT FALSE,
    feed_hours INTEGER NOT NULL,
    top_hours INTEGER NOT NULL,
    price_nano_ton BIGINT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (channel_id, format_type, is_native, feed_hours, top_hours)
);

CREATE INDEX idx_channel_ad_formats_channel_id ON channel_ad_formats(channel_id);
