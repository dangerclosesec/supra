//go:build ignore

package examples

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ConditionedAccess struct {
	db *pgxpool.Pool
}

// TimeWindowParams represents the parameters for a time window condition
type TimeWindowParams struct {
	StartTime string `json:"start_time"` // Format: "HH:MM:SS"
	EndTime   string `json:"end_time"`   // Format: "HH:MM:SS"
}

// GrantTemporalAccess adds a temporal constraint to an existing permission
func (ca *ConditionedAccess) GrantTemporalAccess(
	ctx context.Context,
	actorRoleResourceID uuid.UUID,
	validFrom time.Time,
	validUntil *time.Time,
) error {
	_, err := ca.db.Exec(ctx, `
        INSERT INTO authz_temporal_grant 
        (actor_role_resource_id, valid_from, valid_until)
        VALUES ($1, $2, $3)
    `, actorRoleResourceID, validFrom, validUntil)

	if err != nil {
		return fmt.Errorf("creating temporal grant: %w", err)
	}

	return nil
}

// AddTimeWindowCondition adds a time window condition to an existing permission
func (ca *ConditionedAccess) AddTimeWindowCondition(
	ctx context.Context,
	actorRoleResourceID uuid.UUID,
	startTime, endTime string,
) error {
	// Get the time window condition type ID
	var conditionTypeID uuid.UUID
	err := ca.db.QueryRow(ctx, `
        SELECT id FROM authz_condition_type 
        WHERE name = 'time_window'
    `).Scan(&conditionTypeID)
	if err != nil {
		return fmt.Errorf("finding condition type: %w", err)
	}

	// Create parameters
	params := TimeWindowParams{
		StartTime: startTime,
		EndTime:   endTime,
	}

	paramsJSON, err := json.Marshal(params)
	if err != nil {
		return fmt.Errorf("marshaling parameters: %w", err)
	}

	// Add the condition
	_, err = ca.db.Exec(ctx, `
        INSERT INTO authz_access_condition
        (actor_role_resource_id, condition_type_id, condition_parameters)
        VALUES ($1, $2, $3)
    `, actorRoleResourceID, conditionTypeID, paramsJSON)

	if err != nil {
		return fmt.Errorf("creating access condition: %w", err)
	}

	return nil
}

// CheckAccess verifies if an actor has access considering temporal and conditional constraints
func (ca *ConditionedAccess) CheckAccess(
	ctx context.Context,
	actorID uuid.UUID,
	actorType string,
	resourceID uuid.UUID,
	resourceType string,
	action string,
) (bool, error) {
	var hasAccess bool
	err := ca.db.QueryRow(ctx, `
        WITH action_roles AS (
            SELECT DISTINCT ar.role_id
            FROM authz_active_grants ag
            JOIN authz_role_action ra ON ag.role_id = ra.role_id
            JOIN authz_action a ON ra.action_id = a.id
            WHERE ag.actor_id = $1
            AND ag.actor_type = $2
            AND ag.resource_id = $3
            AND ag.resource_type = $4
            AND a.name = $5
            AND NOT ag.is_negative
            AND (
                ag.condition_type_id IS NULL 
                OR (
                    SELECT (evaluate_time_window_condition(ag.condition_parameters)).is_met
                )
            )
        )
        SELECT EXISTS (SELECT 1 FROM action_roles)
    `, actorID, actorType, resourceID, resourceType, action).Scan(&hasAccess)

	if err != nil {
		return false, fmt.Errorf("checking access: %w", err)
	}

	return hasAccess, nil
}

// Usage example:
func Example() {
	ctx := context.Background()
	access := NewConditionedAccess(db)

	// Grant access only during business hours
	err := access.AddTimeWindowCondition(ctx, permissionID, "09:00:00", "17:00:00")
	if err != nil {
		// handle error
	}

	// Grant temporary access for 24 hours
	validFrom := time.Now()
	validUntil := validFrom.Add(24 * time.Hour)
	err = access.GrantTemporalAccess(ctx, permissionID, validFrom, &validUntil)
	if err != nil {
		// handle error
	}

	// Check if user has access
	hasAccess, err := access.CheckAccess(ctx,
		userID, "user",
		resourceID, "organization",
		"read",
	)
	if err != nil {
		// handle error
	}
	fmt.Printf("Has access: %v\n", hasAccess)
}
