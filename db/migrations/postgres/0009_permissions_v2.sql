-- +goose Up
-- Entity types table stores the different kinds of entities (user, organization, etc.)
CREATE TABLE entity_types (
    id SERIAL PRIMARY KEY,
    name TEXT NOT NULL UNIQUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Entities table stores individual entity instances
CREATE TABLE entities (
    id SERIAL PRIMARY KEY,
    type TEXT NOT NULL,  -- References entity_types.name
    external_id TEXT NOT NULL,
    properties JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(type, external_id)
);

-- Relations table represents the graph edges between entities
CREATE TABLE relations (
    id BIGSERIAL PRIMARY KEY,
    subject_type TEXT NOT NULL,
    subject_id TEXT NOT NULL,
    relation TEXT NOT NULL,
    object_type TEXT NOT NULL,
    object_id TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    -- Composite index for efficient graph traversal
    UNIQUE(subject_type, subject_id, relation, object_type, object_id)
);

-- Create appropriate indices for fast graph traversal
CREATE INDEX idx_relations_subject ON relations(subject_type, subject_id);
CREATE INDEX idx_relations_object ON relations(object_type, object_id);
CREATE INDEX idx_relations_relation ON relations(relation);

-- Permission definitions table stores permission rules with conditions
CREATE TABLE permission_definitions (
    id SERIAL PRIMARY KEY,
    entity_type TEXT NOT NULL,
    permission_name TEXT NOT NULL,
    condition_expression TEXT NOT NULL,
    description TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(entity_type, permission_name)
);

-- Create GIN index for JSONB properties for fast attribute filtering
CREATE INDEX idx_entities_properties ON entities USING GIN (properties);

-- +goose Down
DROP TABLE permission_definitions;
DROP TABLE relations;
DROP TABLE entities;
DROP TABLE entity_types;
