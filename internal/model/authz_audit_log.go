package model

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
)

// AuthzAuditLog represents an authorization audit log entry
type AuthzAuditLog struct {
	ID         uuid.UUID  `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	Timestamp  time.Time  `json:"timestamp" gorm:"default:CURRENT_TIMESTAMP"`
	ActionType string     `json:"action_type"`
	Result     *bool      `json:"result"`
	EntityType string     `json:"entity_type"`
	EntityID   string     `json:"entity_id"`
	SubjectType string    `json:"subject_type"`
	SubjectID  string     `json:"subject_id"`
	Relation   string     `json:"relation"`
	Permission string     `json:"permission"`
	Context    JSONMap    `json:"context" gorm:"type:jsonb"`
	RequestID  string     `json:"request_id"`
	ClientIP   string     `json:"client_ip"`
	UserAgent  string     `json:"user_agent"`
	CreatedAt  time.Time  `json:"created_at" gorm:"default:CURRENT_TIMESTAMP"`
	UpdatedAt  time.Time  `json:"updated_at" gorm:"default:CURRENT_TIMESTAMP"`
}

// TableName specifies the table name for AuthzAuditLog
func (AuthzAuditLog) TableName() string {
	return "authz_audit_logs"
}

// JSONMap represents a generic map stored as JSONB in the database
type JSONMap map[string]interface{}

// Value implements the driver.Valuer interface for JSONMap
func (m JSONMap) Value() (driver.Value, error) {
	if m == nil {
		return nil, nil
	}
	return json.Marshal(m)
}

// Scan implements the sql.Scanner interface for JSONMap
func (m *JSONMap) Scan(value interface{}) error {
	if value == nil {
		*m = make(JSONMap)
		return nil
	}
	
	var bytes []byte
	switch v := value.(type) {
	case []byte:
		bytes = v
	case string:
		bytes = []byte(v)
	default:
		return errors.New("type assertion failed: failed to decode JSONB")
	}
	
	return json.Unmarshal(bytes, m)
}

// Constants for AuthzAuditLog action types
const (
	ActionPermissionCheck = "permission_check"
	ActionEntityCreate    = "entity_create"
	ActionEntityDelete    = "entity_delete"
	ActionRelationCreate  = "relation_create"
	ActionRelationDelete  = "relation_delete"
)