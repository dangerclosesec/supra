-- +goose Up
-- Creates the user_status enum type
CREATE TYPE user_status AS ENUM ('pending', 'active', 'locked', 'suspended', 'deleted');

-- Creates the users table
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    status user_status NOT NULL DEFAULT 'pending',
    email citext NOT NULL,
    first_name TEXT NOT NULL,
    last_name TEXT,
    notification_type TEXT NOT NULL DEFAULT 'email',
    experience TEXT[] NOT NULL DEFAULT ARRAY['other'],
    theme TEXT NOT NULL DEFAULT 'light',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Creates a unique index on email
CREATE UNIQUE INDEX users_email_idx ON users (email);

-- +goose Down
DROP TABLE IF EXISTS users;
DROP TYPE IF EXISTS user_status;
