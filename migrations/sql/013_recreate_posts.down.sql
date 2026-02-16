DROP TABLE IF EXISTS posts;

CREATE TABLE posts (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id),
    name TEXT,
    media_group_id TEXT,
    text TEXT,
    entities JSONB,
    media_type TEXT,
    media_file_id TEXT,
    has_media_spoiler BOOLEAN NOT NULL DEFAULT FALSE,
    show_caption_above_media BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

CREATE INDEX idx_posts_user_id ON posts(user_id);
CREATE INDEX idx_posts_media_group_id ON posts(media_group_id) WHERE media_group_id IS NOT NULL;
