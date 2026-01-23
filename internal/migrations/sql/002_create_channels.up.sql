CREATE TABLE channels (
    id UUID PRIMARY KEY,
    telegram_channel_id BIGINT NOT NULL UNIQUE,
    title TEXT NOT NULL,
    username TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

CREATE INDEX idx_channels_telegram_channel_id ON channels(telegram_channel_id);

CREATE TABLE channel_roles (
    channel_id UUID NOT NULL REFERENCES channels(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id),
    role TEXT NOT NULL CHECK (role IN ('owner', 'manager')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (channel_id, user_id)
);

CREATE INDEX idx_channel_roles_user_id ON channel_roles(user_id);
