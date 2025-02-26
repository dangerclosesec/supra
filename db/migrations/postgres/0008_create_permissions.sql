-- +goose Up
CREATE TABLE authz_resource_type (
    id UUID DEFAULT gen_random_uuid() NOT NULL UNIQUE,
    name TEXT NOT NULL UNIQUE,
    description TEXT,
    created_at TIMESTAMPTZ DEFAULT now(),
    updated_at TIMESTAMPTZ DEFAULT now()
);


INSERT INTO authz_resource_type (name, description)
VALUES ('organization', 'Organization resource type'),
       ('user', 'User resource type'),
       ('project', 'Project resource type');

CREATE TABLE IF NOT EXISTS authz_relationship_type (
    id UUID DEFAULT gen_random_uuid() NOT NULL UNIQUE,
    name TEXT NOT NULL UNIQUE,
    description TEXT,
    source_type TEXT NOT NULL, -- e.g., 'user'
    target_type TEXT NOT NULL, -- e.g., 'organization'
    created_at TIMESTAMPTZ DEFAULT now(),
    updated_at TIMESTAMPTZ DEFAULT now(),
    UNIQUE(name, source_type, target_type)
);

-- Example relationship types
INSERT INTO authz_relationship_type (name, description, source_type, target_type)
VALUES 
    ('member_of', 'User is a member of an organization', 'user', 'organization'),
    ('owner_of', 'User is an owner of an organization', 'user', 'organization'),
    ('belongs_to', 'Project belongs to an organization', 'project', 'organization'),
    ('contributor_to', 'User is a contributor to a project', 'user', 'project');

CREATE TABLE authz_action (
    id UUID DEFAULT gen_random_uuid() NOT NULL UNIQUE,
    name TEXT NOT NULL,
    description TEXT,
    resource_type_id UUID REFERENCES authz_resource_type(id),
    created_at TIMESTAMPTZ DEFAULT now(),
    updated_at TIMESTAMPTZ DEFAULT now(),
    UNIQUE(name, resource_type_id)
);

INSERT INTO authz_action (name, description, resource_type_id)
VALUES ('create', 'Create action', (SELECT id FROM authz_resource_type WHERE name = 'organization')),
       ('read', 'Read action', (SELECT id FROM authz_resource_type WHERE name = 'organization')),
       ('update', 'Update action', (SELECT id FROM authz_resource_type WHERE name = 'organization')),
       ('delete', 'Delete action', (SELECT id FROM authz_resource_type WHERE name = 'organization')),
       ('manage_permissions', 'Manage permissions', (SELECT id FROM authz_resource_type WHERE name = 'organization')),
       ('create', 'Create action', (SELECT id FROM authz_resource_type WHERE name = 'user')),
       ('read', 'Read action', (SELECT id FROM authz_resource_type WHERE name = 'user')),
       ('update', 'Update action', (SELECT id FROM authz_resource_type WHERE name = 'user')),
       ('delete', 'Delete action', (SELECT id FROM authz_resource_type WHERE name = 'user')),
       ('manage_permissions', 'Manage permissions', (SELECT id FROM authz_resource_type WHERE name = 'user')),
       ('create', 'Create action', (SELECT id FROM authz_resource_type WHERE name = 'project')),
       ('read', 'Read action', (SELECT id FROM authz_resource_type WHERE name = 'project')),
       ('update', 'Update action', (SELECT id FROM authz_resource_type WHERE name = 'project')),
       ('delete', 'Delete action', (SELECT id FROM authz_resource_type WHERE name = 'project')),
       ('manage_permissions', 'Manage permissions', (SELECT id FROM authz_resource_type WHERE name = 'project'));


CREATE TABLE authz_role (
    id UUID DEFAULT gen_random_uuid() NOT NULL UNIQUE,
    name TEXT NOT NULL UNIQUE,
    description TEXT,
    created_at TIMESTAMPTZ DEFAULT now(),
    updated_at TIMESTAMPTZ DEFAULT now()
);

INSERT INTO authz_role (name, description)
VALUES ('Org owner', 'Org owner role'),
       ('Org member', 'Org member role');

CREATE TABLE authz_role_action (
    id UUID DEFAULT gen_random_uuid() NOT NULL UNIQUE,
    role_id UUID REFERENCES authz_role(id) ON DELETE CASCADE,
    action_id UUID REFERENCES authz_action(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ DEFAULT now(),
    updated_at TIMESTAMPTZ DEFAULT now(),
    UNIQUE(role_id, action_id)
);

CREATE OR REPLACE VIEW authz_resources AS
SELECT 
    id as resource_id,
    (SELECT id FROM authz_resource_type WHERE name = 'user') as resource_type_id,
    (SELECT name FROM authz_resource_type WHERE name = 'user') as resource_type,
    email as resource_name,
    NULL as parent_id
FROM users
UNION ALL
SELECT 
    id as resource_id,
    (SELECT id FROM authz_resource_type WHERE name = 'organization') as resource_type_id,
    (SELECT name FROM authz_resource_type WHERE name = 'organization') as resource_type,
    name as resource_name,
    NULL as parent_id
FROM organizations;

-- Modify resource relationships to use composite keys
CREATE TABLE authz_resource_relationship (
    id UUID DEFAULT gen_random_uuid() NOT NULL UNIQUE,
    source_type TEXT NOT NULL,
    source_id UUID NOT NULL,
    target_type TEXT NOT NULL,
    target_id UUID NOT NULL,
    relationship_type_id uuid NOT NULL REFERENCES authz_relationship_type(id),
    created_at TIMESTAMPTZ DEFAULT now(),
    updated_at TIMESTAMPTZ DEFAULT now(),
    UNIQUE(source_type, source_id, target_type, target_id, relationship_type_id)
);

-- Modify actor_role_resource to use direct references
CREATE TABLE authz_actor_role_resource (
    id UUID DEFAULT gen_random_uuid() NOT NULL UNIQUE,
    actor_id UUID NOT NULL,
    actor_type TEXT NOT NULL,
    role_id UUID REFERENCES authz_role(id) ON DELETE CASCADE,
    resource_type TEXT NOT NULL,
    resource_id UUID NOT NULL,
    is_negative BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMPTZ DEFAULT now(),
    PRIMARY KEY (actor_id, actor_type, role_id, resource_type, resource_id)
);

-- +goose Down
DROP TABLE authz_actor_role_resource;
DROP TABLE authz_resource_relationship;
DROP TABLE authz_relationship_type;
DROP VIEW authz_resources;
DROP TABLE authz_role_action;
DROP TABLE authz_role;
DROP TABLE authz_action;
DROP TABLE authz_resource_type;
