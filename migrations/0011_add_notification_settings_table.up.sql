CREATE TABLE IF NOT EXISTS notification_settings (
    user_id INTEGER NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    provider_type VARCHAR(50) NOT NULL,
    config JSONB NOT NULL,
    enabled BOOLEAN NOT NULL DEFAULT FALSE,
    PRIMARY KEY (user_id, provider_type)
);