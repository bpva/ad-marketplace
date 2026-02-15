CREATE TABLE channel_info (
    channel_id UUID PRIMARY KEY REFERENCES channels(id) ON DELETE CASCADE,
    about TEXT NOT NULL DEFAULT '',
    subscribers INT NOT NULL DEFAULT 0,
    linked_chat_id BIGINT,
    languages JSONB,
    top_hours JSONB,
    reactions_by_emotion JSONB,
    story_reactions_by_emotion JSONB,
    recent_posts JSONB,
    fetched_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE channel_historical_stats (
    channel_id UUID NOT NULL REFERENCES channels(id) ON DELETE CASCADE,
    date DATE NOT NULL,
    data JSONB NOT NULL,
    PRIMARY KEY (channel_id, date)
);
