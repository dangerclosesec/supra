// internal/auth/auth.go

package auth

import (
	"context"
	"fmt"

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
	client         *client.Client
	tenant         string
	schemaVersion  string
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
		client: client.NewClient(config),
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

	return s.client.DeleteRelation(ctx, req)
}

// CheckPermission checks if a subject has permission on an object
func (s *SupraService) CheckPermission(ctx context.Context, subject Subject, permission string, object Entity, context map[string]interface{}) (bool, error) {
	req := &client.CheckPermissionRequest{
		SubjectType: subject.Type,
		SubjectID:   subject.ID,
		Permission:  permission,
		ObjectType:  object.Type,
		ObjectID:    object.ID,
		Context:     context,
	}

	resp, err := s.client.CheckPermission(ctx, req)
	if err != nil {
		return false, err
	}

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
	return s.client.DeleteEntity(ctx, entityType, entityID)
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
