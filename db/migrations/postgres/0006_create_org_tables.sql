-- +goose Up
-- Creates the organization type enum
CREATE TYPE organization_type AS ENUM (
    'personal',
    'educational',
    'corporate'
);

-- Creates the organization status enum
CREATE TYPE organization_status AS ENUM (
    'pending',
    'active',
    'locked',
    'disabled',
    'deleted'
);

-- Creates the organizations table
CREATE TABLE organizations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL,
    org_type organization_type NOT NULL DEFAULT 'personal',
    status organization_status NOT NULL DEFAULT 'active',
    created_by_id UUID,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (created_by_id) 
        REFERENCES users(id) 
        ON DELETE SET NULL
);

-- Creates unique index for personal organizations per user
CREATE UNIQUE INDEX idx_unique_personal_org_per_user 
    ON organizations (created_by_id) 
    WHERE org_type = 'personal';

-- Creates the organization_domains table
CREATE TABLE organization_domains (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL,
    created_by_id UUID,
    domain TEXT NOT NULL,
    verified BOOLEAN NOT NULL DEFAULT false,
    verification_text TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    verified_at TIMESTAMP,
    FOREIGN KEY (organization_id) 
        REFERENCES organizations(id) 
        ON DELETE CASCADE,
    FOREIGN KEY (created_by_id) 
        REFERENCES users(id) 
        ON DELETE SET NULL
);

-- Creates the organization_users table
CREATE TABLE organization_users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL,
    user_id UUID NOT NULL,
    role TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (organization_id) 
        REFERENCES organizations(id) 
        ON DELETE CASCADE,
    FOREIGN KEY (user_id) 
        REFERENCES users(id) 
        ON DELETE CASCADE
);

-- +goose Down
DROP TABLE IF EXISTS organization_users;
DROP TABLE IF EXISTS organization_domains;
DROP TABLE IF EXISTS organizations;
DROP TYPE IF EXISTS organization_status;
DROP TYPE IF EXISTS organization_type;