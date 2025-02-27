package service

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/dangerclosesec/supra/internal/audit"
	"github.com/dangerclosesec/supra/internal/model"
	"github.com/dangerclosesec/supra/internal/repository"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/google/uuid"
)

// Ensure AuthzAuditLogService implements the audit.Logger interface
var _ audit.Logger = (*AuthzAuditLogService)(nil)

// AuthzAuditLogService handles operations related to authorization audit logs
type AuthzAuditLogService struct {
	repo *repository.AuthzAuditLogRepository
}

// NewAuthzAuditLogService creates a new AuthzAuditLogService
func NewAuthzAuditLogService(repo *repository.AuthzAuditLogRepository) *AuthzAuditLogService {
	return &AuthzAuditLogService{
		repo: repo,
	}
}

// LogPermissionCheck logs a permission check operation
func (s *AuthzAuditLogService) LogPermissionCheck(
	ctx context.Context,
	subject model.Subject,
	permission string,
	object model.Entity,
	result bool,
	contextData map[string]interface{},
	req *http.Request,
) error {
	log := &model.AuthzAuditLog{
		ActionType:  model.ActionPermissionCheck,
		Result:      &result,
		EntityType:  object.Type,
		EntityID:    object.ID,
		SubjectType: subject.Type,
		SubjectID:   subject.ID,
		Permission:  permission,
		Context:     model.JSONMap(contextData),
		Timestamp:   time.Now().UTC(),
	}

	if req != nil {
		log.RequestID = middleware.GetReqID(ctx)
		log.ClientIP = req.RemoteAddr
		log.UserAgent = req.UserAgent()
	}

	return s.repo.Create(ctx, log)
}

// LogEntityCreate logs an entity creation operation
func (s *AuthzAuditLogService) LogEntityCreate(
	ctx context.Context,
	entityType string,
	entityID string, 
	attributes map[string]interface{},
	req *http.Request,
) error {
	log := &model.AuthzAuditLog{
		ActionType: model.ActionEntityCreate,
		EntityType: entityType,
		EntityID:   entityID,
		Context:    model.JSONMap(attributes),
		Timestamp:  time.Now().UTC(),
	}

	if req != nil {
		log.RequestID = middleware.GetReqID(ctx)
		log.ClientIP = req.RemoteAddr
		log.UserAgent = req.UserAgent()
	}

	return s.repo.Create(ctx, log)
}

// LogEntityDelete logs an entity deletion operation
func (s *AuthzAuditLogService) LogEntityDelete(
	ctx context.Context,
	entityType string,
	entityID string,
	req *http.Request,
) error {
	log := &model.AuthzAuditLog{
		ActionType: model.ActionEntityDelete,
		EntityType: entityType,
		EntityID:   entityID,
		Timestamp:  time.Now().UTC(),
	}

	if req != nil {
		log.RequestID = middleware.GetReqID(ctx)
		log.ClientIP = req.RemoteAddr
		log.UserAgent = req.UserAgent()
	}

	return s.repo.Create(ctx, log)
}

// LogRelationCreate logs a relation creation operation
func (s *AuthzAuditLogService) LogRelationCreate(
	ctx context.Context,
	object model.Entity,
	relation string,
	subject model.Subject,
	req *http.Request,
) error {
	log := &model.AuthzAuditLog{
		ActionType:  model.ActionRelationCreate,
		EntityType:  object.Type,
		EntityID:    object.ID,
		SubjectType: subject.Type,
		SubjectID:   subject.ID,
		Relation:    relation,
		Timestamp:   time.Now().UTC(),
	}

	if req != nil {
		log.RequestID = middleware.GetReqID(ctx)
		log.ClientIP = req.RemoteAddr
		log.UserAgent = req.UserAgent()
	}

	return s.repo.Create(ctx, log)
}

// LogRelationDelete logs a relation deletion operation
func (s *AuthzAuditLogService) LogRelationDelete(
	ctx context.Context,
	object model.Entity,
	relation string,
	subject model.Subject,
	req *http.Request,
) error {
	log := &model.AuthzAuditLog{
		ActionType:  model.ActionRelationDelete,
		EntityType:  object.Type,
		EntityID:    object.ID,
		SubjectType: subject.Type,
		SubjectID:   subject.ID,
		Relation:    relation,
		Timestamp:   time.Now().UTC(),
	}

	if req != nil {
		log.RequestID = middleware.GetReqID(ctx)
		log.ClientIP = req.RemoteAddr
		log.UserAgent = req.UserAgent()
	}

	return s.repo.Create(ctx, log)
}

// GetAuditLogs retrieves audit logs based on query parameters
func (s *AuthzAuditLogService) GetAuditLogs(
	ctx context.Context,
	params repository.QueryParams,
) ([]model.AuthzAuditLog, int64, error) {
	return s.repo.Query(ctx, params)
}

// GetAuditLogByID retrieves an audit log by ID
func (s *AuthzAuditLogService) GetAuditLogByID(
	ctx context.Context,
	id uuid.UUID,
) (*model.AuthzAuditLog, error) {
	log, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get audit log by ID: %w", err)
	}

	return log, nil
}