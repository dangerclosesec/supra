-- +goose Up
CREATE TABLE authz_audit_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    timestamp TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    
    -- Action type: permission_check, entity_create, entity_delete, relation_create, relation_delete, etc.
    action_type TEXT NOT NULL,
    
    -- Related to the action - e.g., decision result for permission checks
    result BOOLEAN,
    
    -- Entity information
    entity_type TEXT,
    entity_id TEXT,
    
    -- Subject information (for relations and permission checks)
    subject_type TEXT,
    subject_id TEXT,
    
    -- Relation information (for relation operations)
    relation TEXT,
    
    -- Permission information (for permission checks)
    permission TEXT,
    
    -- Additional context
    context JSONB,
    
    -- Request information
    request_id TEXT,
    client_ip TEXT,
    user_agent TEXT,
    
    -- Metadata tracking
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Add indexes for common query patterns
CREATE INDEX idx_authz_audit_logs_timestamp ON authz_audit_logs (timestamp);
CREATE INDEX idx_authz_audit_logs_action_type ON authz_audit_logs (action_type);
CREATE INDEX idx_authz_audit_logs_entity ON authz_audit_logs (entity_type, entity_id);
CREATE INDEX idx_authz_audit_logs_subject ON authz_audit_logs (subject_type, subject_id);
CREATE INDEX idx_authz_audit_logs_result ON authz_audit_logs (result);

-- +goose Down
DROP TABLE IF EXISTS authz_audit_logs;