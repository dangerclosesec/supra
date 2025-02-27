package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/dangerclosesec/supra/internal/model"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// AuthzAuditLogRepository handles database operations for authorization audit logs
type AuthzAuditLogRepository struct {
	db *gorm.DB
}

// NewAuthzAuditLogRepository creates a new AuthzAuditLogRepository
func NewAuthzAuditLogRepository(db *gorm.DB) *AuthzAuditLogRepository {
	return &AuthzAuditLogRepository{
		db: db,
	}
}

// Create inserts a new audit log entry
func (r *AuthzAuditLogRepository) Create(ctx context.Context, log *model.AuthzAuditLog) error {
	if log.ID == uuid.Nil {
		log.ID = uuid.New()
	}

	if log.Timestamp.IsZero() {
		log.Timestamp = time.Now().UTC()
	}

	result := r.db.WithContext(ctx).Create(log)
	if result.Error != nil {
		return fmt.Errorf("failed to create authorization audit log: %w", result.Error)
	}

	return nil
}

// FindByID retrieves an audit log entry by its ID
func (r *AuthzAuditLogRepository) FindByID(ctx context.Context, id uuid.UUID) (*model.AuthzAuditLog, error) {
	var log model.AuthzAuditLog
	result := r.db.WithContext(ctx).Where("id = ?", id).First(&log)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to find authorization audit log: %w", result.Error)
	}

	return &log, nil
}

// QueryParams holds parameters for querying audit logs
type QueryParams struct {
	ActionType  string
	EntityType  string
	EntityID    string
	SubjectType string
	SubjectID   string
	Result      *bool
	StartTime   time.Time
	EndTime     time.Time
	Limit       int
	Offset      int
}

// Query retrieves audit logs based on the provided query parameters
func (r *AuthzAuditLogRepository) Query(ctx context.Context, params QueryParams) ([]model.AuthzAuditLog, int64, error) {
	var logs []model.AuthzAuditLog
	var count int64
	
	query := r.db.WithContext(ctx).Model(&model.AuthzAuditLog{})

	// Apply filters
	if params.ActionType != "" {
		query = query.Where("action_type = ?", params.ActionType)
	}
	if params.EntityType != "" {
		query = query.Where("entity_type = ?", params.EntityType)
	}
	if params.EntityID != "" {
		query = query.Where("entity_id = ?", params.EntityID)
	}
	if params.SubjectType != "" {
		query = query.Where("subject_type = ?", params.SubjectType)
	}
	if params.SubjectID != "" {
		query = query.Where("subject_id = ?", params.SubjectID)
	}
	if params.Result != nil {
		query = query.Where("result = ?", *params.Result)
	}
	if !params.StartTime.IsZero() {
		query = query.Where("timestamp >= ?", params.StartTime)
	}
	if !params.EndTime.IsZero() {
		query = query.Where("timestamp <= ?", params.EndTime)
	}

	// Get total count for pagination
	if err := query.Count(&count).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count authorization audit logs: %w", err)
	}

	// Apply pagination
	if params.Limit > 0 {
		query = query.Limit(params.Limit)
	} else {
		query = query.Limit(100) // Default limit
	}
	
	if params.Offset > 0 {
		query = query.Offset(params.Offset)
	}

	// Execute query with pagination and ordering
	result := query.Order("timestamp DESC").Find(&logs)
	if result.Error != nil {
		return nil, 0, fmt.Errorf("failed to query authorization audit logs: %w", result.Error)
	}

	return logs, count, nil
}