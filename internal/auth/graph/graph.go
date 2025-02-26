package graph

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Entity represents a node in the graph
type Entity struct {
	ID         int64                  `json:"id"`
	Type       string                 `json:"type"`
	ExternalID string                 `json:"external_id"`
	Properties map[string]interface{} `json:"properties"`
	CreatedAt  time.Time              `json:"created_at"`
	UpdatedAt  time.Time              `json:"updated_at"`
}

// Relation represents an edge in the graph
type Relation struct {
	ID          int64     `json:"id"`
	SubjectType string    `json:"subject_type"`
	SubjectID   string    `json:"subject_id"`
	Relation    string    `json:"relation"`
	ObjectType  string    `json:"object_type"`
	ObjectID    string    `json:"object_id"`
	CreatedAt   time.Time `json:"created_at"`
}

// PermissionDefinition defines a permission rule with a condition expression
type PermissionDefinition struct {
	ID                  int64     `json:"id"`
	EntityType          string    `json:"entity_type"`
	PermissionName      string    `json:"permission_name"`
	ConditionExpression string    `json:"condition_expression"`
	Description         string    `json:"description"`
	CreatedAt           time.Time `json:"created_at"`
}

// IdentityGraph manages the identity graph operations
type IdentityGraph struct {
	Pool *pgxpool.Pool
}

// NewIdentityGraph creates a new instance of IdentityGraph
func NewIdentityGraph(ctx context.Context, connString string) (*IdentityGraph, error) {
	pool, err := pgxpool.New(ctx, connString)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	// Validates the database connection
	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &IdentityGraph{Pool: pool}, nil
}

// Close releases the database connection pool
func (g *IdentityGraph) Close() {
	g.Pool.Close()
}

// CreateEntity adds a new entity to the graph
func (g *IdentityGraph) CreateEntity(ctx context.Context, entityType, externalID string,
	properties map[string]interface{}) (*Entity, error) {

	// Converts properties to JSON
	propertiesJSON, err := json.Marshal(properties)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal properties: %w", err)
	}

	var entity Entity
	err = g.Pool.QueryRow(ctx, `
		INSERT INTO entities (type, external_id, properties)
		VALUES ($1, $2, $3)
		RETURNING id, type, external_id, properties, created_at, updated_at
	`, entityType, externalID, propertiesJSON).Scan(
		&entity.ID, &entity.Type, &entity.ExternalID, &propertiesJSON, &entity.CreatedAt, &entity.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create entity: %w", err)
	}

	// Deserializes the JSON properties back to map
	if err := json.Unmarshal(propertiesJSON, &entity.Properties); err != nil {
		return nil, fmt.Errorf("failed to unmarshal properties: %w", err)
	}

	return &entity, nil
}

// GetEntity retrieves an entity by type and external ID
func (g *IdentityGraph) GetEntity(ctx context.Context, entityType, externalID string) (*Entity, error) {
	var entity Entity
	var propertiesJSON []byte

	err := g.Pool.QueryRow(ctx, `
		SELECT id, type, external_id, properties, created_at, updated_at
		FROM entities
		WHERE type = $1 AND external_id = $2
	`, entityType, externalID).Scan(
		&entity.ID, &entity.Type, &entity.ExternalID, &propertiesJSON, &entity.CreatedAt, &entity.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("entity not found: %s:%s", entityType, externalID)
		}
		return nil, fmt.Errorf("failed to get entity: %w", err)
	}

	// Deserializes the JSON properties
	if err := json.Unmarshal(propertiesJSON, &entity.Properties); err != nil {
		return nil, fmt.Errorf("failed to unmarshal properties: %w", err)
	}

	return &entity, nil
}

// CreateRelation adds a new relation between entities
func (g *IdentityGraph) CreateRelation(ctx context.Context, subjectType, subjectID,
	relation, objectType, objectID string) (*Relation, error) {

	var rel Relation
	err := g.Pool.QueryRow(ctx, `
		INSERT INTO relations (subject_type, subject_id, relation, object_type, object_id)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, subject_type, subject_id, relation, object_type, object_id, created_at
	`, subjectType, subjectID, relation, objectType, objectID).Scan(
		&rel.ID, &rel.SubjectType, &rel.SubjectID, &rel.Relation,
		&rel.ObjectType, &rel.ObjectID, &rel.CreatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create relation: %w", err)
	}

	return &rel, nil
}

// GetRelations retrieves all relations for a subject
func (g *IdentityGraph) GetRelations(ctx context.Context, subjectType, subjectID string) ([]Relation, error) {
	rows, err := g.Pool.Query(ctx, `
		SELECT id, subject_type, subject_id, relation, object_type, object_id, created_at
		FROM relations
		WHERE subject_type = $1 AND subject_id = $2
	`, subjectType, subjectID)

	if err != nil {
		return nil, fmt.Errorf("failed to get relations: %w", err)
	}
	defer rows.Close()

	var relations []Relation
	for rows.Next() {
		var rel Relation
		if err := rows.Scan(
			&rel.ID, &rel.SubjectType, &rel.SubjectID, &rel.Relation,
			&rel.ObjectType, &rel.ObjectID, &rel.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan relation: %w", err)
		}
		relations = append(relations, rel)
	}

	return relations, nil
}

// CheckPermission determines if a subject has a permission on an object
func (g *IdentityGraph) CheckPermission(ctx context.Context, subjectType, subjectID,
	permission, objectType, objectID string) (bool, error) {

	// First, get the permission definition to find the condition expression
	var conditionExpr string
	err := g.Pool.QueryRow(ctx, `
		SELECT condition_expression
		FROM permission_definitions
		WHERE entity_type = $1 AND permission_name = $2
	`, objectType, permission).Scan(&conditionExpr)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, fmt.Errorf("permission definition not found: %s.%s", objectType, permission)
		}
		return false, fmt.Errorf("failed to get permission definition: %w", err)
	}

	// Evaluates the condition expression
	return g.EvaluateCondition(ctx, conditionExpr, subjectType, subjectID, objectType, objectID)
}

// EvaluateCondition evaluates a permission condition expression
func (g *IdentityGraph) EvaluateCondition(ctx context.Context, conditionExpr,
	subjectType, subjectID, objectType, objectID string) (bool, error) {

	// Create a parser for the condition expression
	parser := NewConditionParser(conditionExpr)
	expr, err := parser.Parse()
	if err != nil {
		return false, fmt.Errorf("failed to parse condition: %w", err)
	}

	// Evaluate the parsed expression
	return g.evaluateExpression(ctx, expr, subjectType, subjectID, objectType, objectID)
}

// evaluateExpression evaluates a parsed condition expression
func (g *IdentityGraph) evaluateExpression(ctx context.Context, expr Expression,
	subjectType, subjectID, objectType, objectID string) (bool, error) {

	switch e := expr.(type) {
	case *AndExpression:
		// Evaluate left expression
		leftResult, err := g.evaluateExpression(ctx, e.Left, subjectType, subjectID, objectType, objectID)
		if err != nil {
			return false, err
		}

		// Short-circuit if left expression is false
		if !leftResult {
			return false, nil
		}

		// Evaluate right expression
		return g.evaluateExpression(ctx, e.Right, subjectType, subjectID, objectType, objectID)

	case *OrExpression:
		// Evaluate left expression
		leftResult, err := g.evaluateExpression(ctx, e.Left, subjectType, subjectID, objectType, objectID)
		if err != nil {
			return false, err
		}

		// Short-circuit if left expression is true
		if leftResult {
			return true, nil
		}

		// Evaluate right expression
		return g.evaluateExpression(ctx, e.Right, subjectType, subjectID, objectType, objectID)

	case *RelationExpression:
		// Handle direct relation check (e.g., "owner")
		if e.RelationPath == "" {
			return g.checkDirectRelation(ctx, subjectType, subjectID, e.RelationName, objectType, objectID)
		}

		// Handle indirect relation check (e.g., "organization.owner")
		return g.checkIndirectRelation(ctx, subjectType, subjectID, e.RelationPath, e.RelationName, objectType, objectID)

	default:
		return false, fmt.Errorf("unknown expression type")
	}
}

// checkDirectRelation checks if subject has a direct relation to the object
// IMPORTANT: We now check in BOTH directions - where subject->object and where object->subject
func (g *IdentityGraph) checkDirectRelation(ctx context.Context,
	subjectType, subjectID, relation, objectType, objectID string) (bool, error) {

	// First check subject -> object direction (as before)
	var exists bool
	err := g.Pool.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1
			FROM relations
			WHERE subject_type = $1
			AND subject_id = $2
			AND relation = $3
			AND object_type = $4
			AND object_id = $5
		)
	`, subjectType, subjectID, relation, objectType, objectID).Scan(&exists)

	if err != nil {
		return false, fmt.Errorf("failed to check direct relation (subject->object): %w", err)
	}

	if exists {
		return true, nil
	}

	// If not found, check object -> subject direction (this is the fix for your schema format)
	err = g.Pool.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1
			FROM relations
			WHERE subject_type = $1
			AND subject_id = $2
			AND relation = $3
			AND object_type = $4
			AND object_id = $5
		)
	`, objectType, objectID, relation, subjectType, subjectID).Scan(&exists)

	if err != nil {
		return false, fmt.Errorf("failed to check direct relation (object->subject): %w", err)
	}

	return exists, nil
}

// checkIndirectRelation checks for relations through intermediate entities
func (g *IdentityGraph) checkIndirectRelation(ctx context.Context,
	subjectType, subjectID, relationPath, relationName, objectType, objectID string) (bool, error) {

	// Handle path.relation format (e.g. organization.owner)
	var exists bool

	// First try with organization being the subject
	err := g.Pool.QueryRow(ctx, `
		WITH RECURSIVE path(subject_type, subject_id, relation, object_type, object_id, depth) AS (
			-- Start with direct relations from the object
			SELECT subject_type, subject_id, relation, object_type, object_id, 1
			FROM relations
			WHERE subject_type = $1
			AND subject_id = $2
			
			UNION ALL
			
			-- Follow the graph
			SELECT r.subject_type, r.subject_id, r.relation, r.object_type, r.object_id, p.depth + 1
			FROM relations r
			JOIN path p ON r.subject_type = p.object_type AND r.subject_id = p.object_id
			WHERE p.depth < 10  -- Prevents infinite recursion
		)
		SELECT EXISTS (
			SELECT 1
			FROM path
			WHERE relation = $3
			AND object_type = $4 
			AND object_id = $5
		)
	`, relationPath, objectID, relationName, subjectType, subjectID).Scan(&exists)

	if err != nil {
		return false, fmt.Errorf("failed to check indirect relation: %w", err)
	}

	if exists {
		return true, nil
	}

	// If not found, try with organization being the object (this is the fix for your schema format)
	err = g.Pool.QueryRow(ctx, `
		WITH RECURSIVE path(subject_type, subject_id, relation, object_type, object_id, depth) AS (
			-- Start with direct relations to the object
			SELECT subject_type, subject_id, relation, object_type, object_id, 1
			FROM relations
			WHERE object_type = $1
			AND object_id = $2
			
			UNION ALL
			
			-- Follow the graph
			SELECT r.subject_type, r.subject_id, r.relation, r.object_type, r.object_id, p.depth + 1
			FROM relations r
			JOIN path p ON r.object_type = p.subject_type AND r.object_id = p.subject_id
			WHERE p.depth < 10  -- Prevents infinite recursion
		)
		SELECT EXISTS (
			SELECT 1
			FROM path
			WHERE relation = $3
			AND subject_type = $4 
			AND subject_id = $5
		)
	`, relationPath, objectID, relationName, subjectType, subjectID).Scan(&exists)

	if err != nil {
		return false, fmt.Errorf("failed to check indirect relation (reverse): %w", err)
	}

	return exists, nil
}

// AddPermissionDefinition adds a new permission definition
func (g *IdentityGraph) AddPermissionDefinition(ctx context.Context, entityType, permissionName,
	conditionExpr, description string) (*PermissionDefinition, error) {

	var def PermissionDefinition
	err := g.Pool.QueryRow(ctx, `
		INSERT INTO permission_definitions (entity_type, permission_name, condition_expression, description)
		VALUES ($1, $2, $3, $4)
		RETURNING id, entity_type, permission_name, condition_expression, description, created_at
	`, entityType, permissionName, conditionExpr, description).Scan(
		&def.ID, &def.EntityType, &def.PermissionName, &def.ConditionExpression,
		&def.Description, &def.CreatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to add permission definition: %w", err)
	}

	return &def, nil
}
