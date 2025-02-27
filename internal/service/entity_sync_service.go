// internal/service/entity_sync_service.go
package service

import (
	"context"
	"fmt"
	"strings"

	"github.com/dangerclosesec/supra/internal/auth"
	"github.com/dangerclosesec/supra/internal/model"
	"github.com/google/uuid"
)

// EntitySyncService handles synchronization between database entities and permission model
type EntitySyncService struct {
	supraService *auth.SupraService
	// Could add caching, error tracking, etc.
}

// NewEntitySyncService creates a new sync service
func NewEntitySyncService(supraService *auth.SupraService) *EntitySyncService {
	return &EntitySyncService{
		supraService: supraService,
	}
}

// SyncUserToPermissions creates or updates a user entity in the permission system
func (s *EntitySyncService) SyncUserToPermissions(ctx context.Context, user *model.User) error {
	// Extract relevant attributes from user model
	attributes := map[string]interface{}{
		"is_verified":  user.Status == model.StatusActive,
		"email_domain": extractDomainFromEmail(user.Email),
	}

	// Create entity in the permission system
	entityID := user.ID.String()

	// Write attributes to the permission system
	if err := s.supraService.WriteEntityAttributes(ctx, "user", entityID, attributes); err != nil {
		return fmt.Errorf("writing user attributes: %w", err)
	}

	return nil
}

// SyncOrganizationToPermissions creates or updates an organization entity in the permission system
func (s *EntitySyncService) SyncOrganizationToPermissions(ctx context.Context, org *model.Organization) error {
	// Extract relevant attributes based on your permission schema
	attributes := map[string]interface{}{
		// Add relevant attributes dynamically based on your schema
		// Keeping this generic rather than hardcoding specific attributes
		"name": org.Name,
		"type": string(org.OrgType),
	}

	entityID := org.ID.String()

	// Write attributes to the permission system
	if err := s.supraService.WriteEntityAttributes(ctx, "organization", entityID, attributes); err != nil {
		return fmt.Errorf("writing organization attributes: %w", err)
	}

	return nil
}

// EstablishUserOrganizationRelation creates a relationship between a user and organization
func (s *EntitySyncService) EstablishUserOrganizationRelation(ctx context.Context, orgID, userID uuid.UUID, role string) error {
	// Map role to relation type as defined in your permission schema
	relation := role // In this case assuming role names match relation names

	// Write the relationship
	return s.supraService.WriteRelationship(
		auth.Entity{Type: "organization", ID: orgID.String()},
		relation,
		auth.Subject{Type: "user", ID: userID.String()},
	)
}

// Helper to extract domain from email
func extractDomainFromEmail(email string) string {
	parts := strings.Split(email, "@")
	if len(parts) == 2 {
		return parts[1]
	}
	return ""
}

// DeleteEntityFromPermissions removes an entity from the permission system
func (s *EntitySyncService) DeleteEntityFromPermissions(ctx context.Context, entityType string, id uuid.UUID) error {
	// First delete all relationships for this entity
	err := s.supraService.DeleteRelationship(
		auth.Entity{Type: entityType, ID: id.String()},
		"*",                              // Wildcard to match any relation
		auth.Subject{Type: "*", ID: "*"}, // Wildcard to match any subject
	)
	if err != nil {
		return fmt.Errorf("deleting entity relationships: %w", err)
	}

	// Then delete the entity itself
	return s.supraService.DeleteEntity(ctx, entityType, id.String())
}
