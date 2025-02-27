// internal/auth/auth.go

package auth

import (
	"context"
	"fmt"
	"net/http"

	"github.com/dangerclosesec/supra/internal/audit"
	"github.com/dangerclosesec/supra/internal/model"
	"github.com/dangerclosesec/supra/sdk/client"
)

// Entity represents a permission entity
type Entity struct {
	Type string
	ID   string
}

// Subject represents a permission subject
type Subject struct {
	Type string
	ID   string
}

// SupraServiceOption defines function signature for service options
type SupraServiceOption func(*SupraService)

// SupraService handles communication with the permission service
type SupraService struct {
	client        *client.Client
	tenant        string
	schemaVersion string
	auditLogger   audit.Logger
	httpRequest   *http.Request // Current HTTP request, if available
}

// WitTenant sets the tenant for the service
func WitTenant(tenant string) SupraServiceOption {
	return func(s *SupraService) {
		s.tenant = tenant
	}
}

// WithSchemaVersion sets the schema version for the service
func WithSchemaVersion(version string) SupraServiceOption {
	return func(s *SupraService) {
		s.schemaVersion = version
	}
}

// WithAuditLogger sets the audit logger for tracking operations
func WithAuditLogger(logger audit.Logger) SupraServiceOption {
	return func(s *SupraService) {
		s.auditLogger = logger
	}
}

// WithHTTPRequest sets the current HTTP request for audit logging
func WithHTTPRequest(req *http.Request) SupraServiceOption {
	return func(s *SupraService) {
		s.httpRequest = req
	}
}

// NewSupraService creates a new SupraService instance
func NewSupraService(host string, opts ...SupraServiceOption) (*SupraService, error) {
	if host == "" {
		return nil, fmt.Errorf("host is required")
	}

	// Configure the client
	config := &client.Config{
		BaseURL: host,
	}

	svc := &SupraService{
		client:      client.NewClient(config),
		auditLogger: &audit.NoOpLogger{}, // Default to no-op logger
	}

	// Apply options
	for _, opt := range opts {
		opt(svc)
	}

	return svc, nil
}

// WriteEntityAttributes creates or updates entity attributes
func (s *SupraService) WriteEntityAttributes(ctx context.Context, entityType, entityID string, attributes map[string]interface{}) error {
	req := &client.CreateEntityRequest{
		Type:       entityType,
		ExternalID: entityID,
		Properties: attributes,
	}

	_, err := s.client.CreateEntity(ctx, req)

	// Log the operation
	if err == nil {
		_ = s.auditLogger.LogEntityCreate(ctx, entityType, entityID, attributes, s.httpRequest)
	}

	return err
}

// WriteRelationship creates a relationship between two entities
func (s *SupraService) WriteRelationship(object Entity, relation string, subject Subject) error {
	ctx := context.Background()
	req := &client.CreateRelationRequest{
		ObjectType:  object.Type,
		ObjectID:    object.ID,
		Relation:    relation,
		SubjectType: subject.Type,
		SubjectID:   subject.ID,
	}

	_, err := s.client.CreateRelation(ctx, req)

	// Log the operation
	if err == nil {
		modelObj := model.Entity{Type: object.Type, ID: object.ID}
		modelSubj := model.Subject{Type: subject.Type, ID: subject.ID}
		_ = s.auditLogger.LogRelationCreate(ctx, modelObj, relation, modelSubj, s.httpRequest)
	}

	return err
}

// DeleteRelationship deletes a relationship between entities
func (s *SupraService) DeleteRelationship(object Entity, relation string, subject Subject) error {
	ctx := context.Background()
	req := &client.DeleteRelationRequest{
		ObjectType:  object.Type,
		ObjectID:    object.ID,
		Relation:    relation,
		SubjectType: subject.Type,
		SubjectID:   subject.ID,
	}

	err := s.client.DeleteRelation(ctx, req)

	// Log the operation
	if err == nil {
		modelObj := model.Entity{Type: object.Type, ID: object.ID}
		modelSubj := model.Subject{Type: subject.Type, ID: subject.ID}
		_ = s.auditLogger.LogRelationDelete(ctx, modelObj, relation, modelSubj, s.httpRequest)
	}

	return err
}

// CheckPermission checks if a subject has permission on an object
func (s *SupraService) CheckPermission(ctx context.Context, subject Subject, permission string, object Entity, contextData map[string]interface{}) (bool, error) {
	req := &client.CheckPermissionRequest{
		SubjectType: subject.Type,
		SubjectID:   subject.ID,
		Permission:  permission,
		ObjectType:  object.Type,
		ObjectID:    object.ID,
		Context:     contextData,
	}

	resp, err := s.client.CheckPermission(ctx, req)
	if err != nil {
		return false, err
	}

	// Log the permission check
	modelObj := model.Entity{Type: object.Type, ID: object.ID}
	modelSubj := model.Subject{Type: subject.Type, ID: subject.ID}
	_ = s.auditLogger.LogPermissionCheck(ctx, modelSubj, permission, modelObj, resp.Allowed, contextData, s.httpRequest)

	return resp.Allowed, nil
}

// TestRelationship tests if a relationship exists between entities
func (s *SupraService) TestRelationship(ctx context.Context, subject Subject, relation string, object Entity) (bool, error) {
	req := &client.TestRelationRequest{
		SubjectType: subject.Type,
		SubjectID:   subject.ID,
		Relation:    relation,
		ObjectType:  object.Type,
		ObjectID:    object.ID,
	}

	resp, err := s.client.TestRelation(ctx, req)
	if err != nil {
		return false, err
	}

	return resp.HasRelation, nil
}

// DeleteEntity deletes an entity from the permission system
func (s *SupraService) DeleteEntity(ctx context.Context, entityType string, entityID string) error {
	err := s.client.DeleteEntity(ctx, entityType, entityID)

	// Log the operation
	if err == nil {
		_ = s.auditLogger.LogEntityDelete(ctx, entityType, entityID, s.httpRequest)
	}

	return err
}

// ListPermissionDefinitions lists all permission definitions
func (s *SupraService) ListPermissionDefinitions(ctx context.Context) ([]client.PermissionDefinition, error) {
	return s.client.ListPermissionDefinitions(ctx)
}

// CreatePermission creates a new permission definition
func (s *SupraService) CreatePermission(ctx context.Context, entityType, permissionName, conditionExpression, description string) error {
	req := &client.CreatePermissionRequest{
		EntityType:          entityType,
		PermissionName:      permissionName,
		ConditionExpression: conditionExpression,
		Description:         description,
	}

	_, err := s.client.CreatePermission(ctx, req)
	return err
}

// DeletePermission deletes a permission definition
func (s *SupraService) DeletePermission(ctx context.Context, entityType, permissionName string) error {
	return s.client.DeletePermission(ctx, entityType, permissionName)
}

// SetHTTPRequest sets the current HTTP request for audit logging
func (s *SupraService) SetHTTPRequest(ctx context.Context, req *http.Request) {
	s.httpRequest = req
}
