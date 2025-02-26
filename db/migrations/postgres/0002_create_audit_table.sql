-- +goose Up
CREATE TABLE audit_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    resource TEXT,
    resource_id TEXT,
    operation_type TEXT,
    actor TEXT,
    details JSONB,
    before JSONB,
    after JSONB,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- +goose Down
DROP TABLE IF EXISTS audit_logs;