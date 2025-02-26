//go:build ignore

package examples

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type RelationshipManager struct {
	db *pgxpool.Pool
}

func NewRelationshipManager(db *pgxpool.Pool) *RelationshipManager {
	return &RelationshipManager{db: db}
}

// AddUserToOrganization creates a membership relationship between user and org
func (r *RelationshipManager) AddUserToOrganization(ctx context.Context, userID, orgID uuid.UUID, relationshipName string) error {
	// First get the relationship type ID
	var relationshipTypeID uuid.UUID
	err := r.db.QueryRow(ctx, `
        SELECT id FROM authz_relationship_type 
        WHERE name = $1 
        AND source_type = 'user' 
        AND target_type = 'organization'
    `, relationshipName).Scan(&relationshipTypeID)
	if err != nil {
		return fmt.Errorf("finding relationship type: %w", err)
	}

	// Create the relationship
	_, err = r.db.Exec(ctx, `
        INSERT INTO authz_resource_relationship
        (source_type, source_id, target_type, target_id, relationship_type_id)
        VALUES ($1, $2, $3, $4, $5)
    `, "user", userID, "organization", orgID, relationshipTypeID)

	if err != nil {
		return fmt.Errorf("creating relationship: %w", err)
	}

	return nil
}

// Example usage:
func Example() {
	// Add a user as a member of an organization
	err := relationshipMgr.AddUserToOrganization(ctx, userID, orgID, "member_of")
	if err != nil {
		// handle error
	}

	// Later, you can query relationships for authorization checks
	// For example, to check if a user is a member of an organization:
	var exists bool
	err = db.QueryRow(ctx, `
        SELECT EXISTS (
            SELECT 1 FROM authz_resource_relationship rr
            JOIN authz_relationship_type rt ON rr.relationship_type_id = rt.id
            WHERE rt.name = 'member_of'
            AND rr.source_type = 'user'
            AND rr.source_id = $1
            AND rr.target_type = 'organization'
            AND rr.target_id = $2
        )
    `, userID, orgID).Scan(&exists)
}
