-- +goose Up
-- Rule definitions table stores reusable rules for permission conditions
CREATE TABLE rule_definitions (
    id SERIAL PRIMARY KEY,
    rule_name TEXT NOT NULL UNIQUE,
    parameters JSONB NOT NULL,
    expression TEXT NOT NULL,
    description TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- +goose Down
DROP TABLE rule_definitions;

