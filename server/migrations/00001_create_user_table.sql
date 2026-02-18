-- +goose Up
-- +goose StatementBegin
CREATE TABLE users (
    -- Internal ID for foreign key relations in your DB
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    -- The ID provided by Clerk (usually starts with 'user_...')
    clerk_id   VARCHAR(255) UNIQUE NOT NULL,

    -- Synced profile data
    email      VARCHAR(255) UNIQUE NOT NULL,
    first_name VARCHAR(100),
    last_name  VARCHAR(100),
    avatar_url TEXT,

    -- Metadata and Timestamps
    last_login_at TIMESTAMP WITH TIME ZONE,
    created_at    TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Index for fast lookups when the Clerk Webhook hits or Middleware runs
CREATE INDEX idx_users_clerk_id ON users(clerk_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS users;
-- +goose StatementEnd
