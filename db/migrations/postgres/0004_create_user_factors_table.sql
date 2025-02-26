-- +goose Up
-- Creates the user_factor_type enum
CREATE TYPE user_factor_type AS ENUM (
    'webauthn',
    'passkey',
    'hashpass',
    'pubkey',
    'totp',
    'openid',
    'sms',
    'email',
    'u2f',
    'backup_code',
    'verification_code'
);

-- Creates the user_factors table
CREATE TABLE user_factors (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,
    factor_type user_factor_type NOT NULL,
    federated_auth_provider TEXT,
    federated_auth_external_id TEXT,
    webauthn_id BYTEA,
    material TEXT,
    backup_codes JSONB,
    client_id TEXT,
    device_id TEXT,
    is_active BOOLEAN NOT NULL DEFAULT true,
    last_used_at TIMESTAMP,
    verified_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id)
);

-- Adds check constraint requiring client_id for OpenID factors
ALTER TABLE user_factors 
    ADD CONSTRAINT client_id_not_null_on_oauth 
    CHECK (
        factor_type != 'openid' OR 
        (factor_type = 'openid' AND client_id IS NOT NULL)
    );

-- Adds unique constraint for user factors
ALTER TABLE user_factors 
    ADD CONSTRAINT unique_user_factors 
    UNIQUE (user_id, factor_type, federated_auth_provider, federated_auth_external_id);

-- +goose Down
DROP TABLE IF EXISTS user_factors;
DROP TYPE IF EXISTS user_factor_type;