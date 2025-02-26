-- +goose Up
-- Creates enum for different policy types
CREATE TYPE policy_type AS ENUM (
    'terms_of_service',
    'privacy_policy',
    'acceptable_use',
    'cookie_policy'
);

-- Table for storing the policies themselves
CREATE TABLE policies (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    type policy_type NOT NULL,
    name TEXT NOT NULL,
    description TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (type)
);

-- Table for storing different versions of policies
CREATE TABLE policy_versions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    policy_id UUID NOT NULL,
    version TEXT NOT NULL,
    content TEXT NOT NULL,
    effective_date TIMESTAMP NOT NULL,
    sunset_date TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (policy_id) REFERENCES policies(id) ON DELETE CASCADE,
    UNIQUE (policy_id, version)
);

-- Table for tracking user acceptances of policy versions
CREATE TABLE policy_acceptances (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,
    policy_version_id UUID NOT NULL,
    accepted_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    accepted_from_ip TEXT,
    user_agent TEXT,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (policy_version_id) REFERENCES policy_versions(id),
    UNIQUE (user_id, policy_version_id)
);

-- Index for looking up latest policy versions
CREATE INDEX idx_policy_versions_effective_date 
    ON policy_versions (policy_id, effective_date DESC);

-- Index for querying user acceptances
CREATE INDEX idx_policy_acceptances_user 
    ON policy_acceptances (user_id, accepted_at DESC);

-- +goose Down
DROP TABLE IF EXISTS policy_acceptances;
DROP TABLE IF EXISTS policy_versions;
DROP TABLE IF EXISTS policies;
DROP TYPE IF EXISTS policy_type;