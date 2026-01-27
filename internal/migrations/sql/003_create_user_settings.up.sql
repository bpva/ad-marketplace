CREATE TABLE user_settings (
    user_id UUID PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    language VARCHAR(2) NOT NULL DEFAULT 'en',
    receive_notifications BOOLEAN NOT NULL DEFAULT TRUE,
    preferred_mode TEXT NOT NULL DEFAULT 'publisher',
    onboarding_finished BOOLEAN NOT NULL DEFAULT FALSE
);
