package graph

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync"
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

// RuleDefinition represents a rule definition with parameters and expression
type RuleDefinition struct {
	ID         int64             `json:"id"`
	Name       string            `json:"name"`
	Parameters []RuleParameter   `json:"parameters"`
	Expression string            `json:"expression"`
	Description string           `json:"description,omitempty"`
	CreatedAt  time.Time         `json:"created_at"`
}

// RuleParameter represents a parameter for a rule
type RuleParameter struct {
	Name     string `json:"name"`
	DataType string `json:"data_type"`
}

// IdentityGraph manages the identity graph operations
type IdentityGraph struct {
	Pool        *pgxpool.Pool
	ruleCache   map[string]*RuleDefinition
	ruleCacheMu sync.RWMutex
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

	graph := &IdentityGraph{
		Pool:      pool,
		ruleCache: make(map[string]*RuleDefinition),
	}

	// Pre-load rules from the database
	if err := graph.loadRules(ctx); err != nil {
		return nil, fmt.Errorf("failed to load rules: %w", err)
	}

	return graph, nil
}

// loadRules loads all rule definitions from the database into the cache
func (g *IdentityGraph) loadRules(ctx context.Context) error {
	rows, err := g.Pool.Query(ctx, `
		SELECT id, rule_name, parameters, expression, description, created_at
		FROM rule_definitions
	`)
	if err != nil {
		return fmt.Errorf("failed to query rule definitions: %w", err)
	}
	defer rows.Close()

	g.ruleCacheMu.Lock()
	defer g.ruleCacheMu.Unlock()

	for rows.Next() {
		var rule RuleDefinition
		var parametersJSON []byte

		err := rows.Scan(&rule.ID, &rule.Name, &parametersJSON, &rule.Expression, &rule.Description, &rule.CreatedAt)
		if err != nil {
			return fmt.Errorf("failed to scan rule definition: %w", err)
		}

		// Parse parameters JSON
		if err := json.Unmarshal(parametersJSON, &rule.Parameters); err != nil {
			return fmt.Errorf("failed to unmarshal rule parameters: %w", err)
		}

		// Add to cache
		g.ruleCache[rule.Name] = &rule
	}

	if err := rows.Err(); err != nil {
		return fmt.Errorf("error iterating rule definitions: %w", err)
	}

	return nil
}

// GetRule retrieves a rule definition by name
func (g *IdentityGraph) GetRule(ruleName string) (*RuleDefinition, error) {
	g.ruleCacheMu.RLock()
	defer g.ruleCacheMu.RUnlock()

	rule, exists := g.ruleCache[ruleName]
	if !exists {
		return nil, fmt.Errorf("rule not found: %s", ruleName)
	}
	return rule, nil
}

// AddRule adds a new rule definition to the database and cache
func (g *IdentityGraph) AddRule(ctx context.Context, rule *RuleDefinition) error {
	// Convert parameters to JSON
	parametersJSON, err := json.Marshal(rule.Parameters)
	if err != nil {
		return fmt.Errorf("failed to marshal parameters: %w", err)
	}

	// Insert rule into database
	err = g.Pool.QueryRow(ctx, `
		INSERT INTO rule_definitions (rule_name, parameters, expression, description)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at
	`, rule.Name, parametersJSON, rule.Expression, rule.Description).Scan(&rule.ID, &rule.CreatedAt)

	if err != nil {
		return fmt.Errorf("failed to insert rule definition: %w", err)
	}

	// Add to cache
	g.ruleCacheMu.Lock()
	defer g.ruleCacheMu.Unlock()
	g.ruleCache[rule.Name] = rule

	return nil
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
	permission, objectType, objectID string, contextData map[string]interface{}) (bool, error) {

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
	return g.EvaluateCondition(ctx, conditionExpr, subjectType, subjectID, objectType, objectID, contextData)
}

// EvaluateCondition evaluates a permission condition expression
func (g *IdentityGraph) EvaluateCondition(ctx context.Context, conditionExpr,
	subjectType, subjectID, objectType, objectID string, contextData map[string]interface{}) (bool, error) {

	// Create a parser for the condition expression
	parser := NewConditionParser(conditionExpr)
	expr, err := parser.Parse()
	if err != nil {
		return false, fmt.Errorf("failed to parse condition: %w", err)
	}

	// Evaluate the parsed expression
	return g.evaluateExpression(ctx, expr, subjectType, subjectID, objectType, objectID, contextData)
}

// evaluateExpression evaluates a parsed condition expression
func (g *IdentityGraph) evaluateExpression(ctx context.Context, expr Expression,
	subjectType, subjectID, objectType, objectID string, contextData map[string]interface{}) (bool, error) {

	switch e := expr.(type) {
	case *AndExpression:
		// Evaluate left expression
		leftResult, err := g.evaluateExpression(ctx, e.Left, subjectType, subjectID, objectType, objectID, contextData)
		if err != nil {
			return false, err
		}

		// Short-circuit if left expression is false
		if !leftResult {
			return false, nil
		}

		// Evaluate right expression
		return g.evaluateExpression(ctx, e.Right, subjectType, subjectID, objectType, objectID, contextData)

	case *OrExpression:
		// Evaluate left expression
		leftResult, err := g.evaluateExpression(ctx, e.Left, subjectType, subjectID, objectType, objectID, contextData)
		if err != nil {
			return false, err
		}

		// Short-circuit if left expression is true
		if leftResult {
			return true, nil
		}

		// Evaluate right expression
		return g.evaluateExpression(ctx, e.Right, subjectType, subjectID, objectType, objectID, contextData)

	case *RelationExpression:
		// Handle direct relation check (e.g., "owner")
		if e.RelationPath == "" {
			return g.checkDirectRelation(ctx, subjectType, subjectID, e.RelationName, objectType, objectID)
		}

		// Handle indirect relation check (e.g., "organization.owner")
		return g.checkIndirectRelation(ctx, subjectType, subjectID, e.RelationPath, e.RelationName, objectType, objectID)

	case *ContextExpression:
		// Extract the value from the context data
		if contextData == nil {
			return false, fmt.Errorf("context data required but not provided")
		}
		
		// Navigate the path to get the value
		currentValue := contextData
		for i, pathPart := range e.Path {
			if i == len(e.Path)-1 {
				// Last path segment - should be the value we want
				// For comparison purposes, we just return true if the key exists
				_, exists := currentValue[pathPart]
				return exists, nil
			}
			
			// Get the next level of the path
			nextValue, exists := currentValue[pathPart]
			if !exists {
				return false, fmt.Errorf("context path not found: %s", strings.Join(e.Path, "."))
			}
			
			// Check if the next value is a map
			nextMap, ok := nextValue.(map[string]interface{})
			if !ok {
				return false, fmt.Errorf("context path is not a map: %s", strings.Join(e.Path[:i+1], "."))
			}
			
			currentValue = nextMap
		}
		
		return false, fmt.Errorf("invalid context path: %s", strings.Join(e.Path, "."))

	case *AttributeExpression:
		// Get the entity attribute value
		attributeValue, err := g.getEntityAttribute(ctx, e.EntityType, objectID, e.AttributeName)
		if err != nil {
			return false, fmt.Errorf("failed to get attribute: %w", err)
		}
		
		// For simple attribute existence check, just return true if we got here
		return attributeValue != nil, nil

	case *RuleExpression:
		// Evaluate rule expression with arguments
		return g.evaluateRule(ctx, e, subjectType, subjectID, objectType, objectID, contextData)

	default:
		return false, fmt.Errorf("unknown expression type: %T", expr)
	}
}

// evaluateRule evaluates a rule expression by calling the appropriate rule function
func (g *IdentityGraph) evaluateRule(ctx context.Context, rule *RuleExpression,
	subjectType, subjectID, objectType, objectID string, contextData map[string]interface{}) (bool, error) {
	
	// Get the rule definition from the registry
	ruleDef, err := g.GetRule(rule.RuleName)
	if err != nil {
		return false, fmt.Errorf("failed to get rule definition: %w", err)
	}
	
	// Evaluate the arguments first
	argValues := make([]interface{}, len(rule.Arguments))
	for i, argExpr := range rule.Arguments {
		switch e := argExpr.(type) {
		case *ContextExpression:
			// Get value from context
			if contextData == nil {
				return false, fmt.Errorf("context data required for rule argument but not provided")
			}
			
			// Navigate path to get value
			currentValue := contextData
			for j, pathPart := range e.Path {
				if j == len(e.Path)-1 {
					// Last path segment - extract the value
					val, exists := currentValue[pathPart]
					if !exists {
						return false, fmt.Errorf("context path not found for rule argument: %s", strings.Join(e.Path, "."))
					}
					argValues[i] = val
					break
				}
				
				// Get the next level of the path
				nextValue, exists := currentValue[pathPart]
				if !exists {
					return false, fmt.Errorf("context path not found for rule argument: %s", strings.Join(e.Path, "."))
				}
				
				// Check if the next value is a map
				nextMap, ok := nextValue.(map[string]interface{})
				if !ok {
					return false, fmt.Errorf("context path is not a map for rule argument: %s", strings.Join(e.Path[:j+1], "."))
				}
				
				currentValue = nextMap
			}
			
		case *AttributeExpression:
			// Get attribute value from the database
			attributeValue, err := g.getEntityAttribute(ctx, e.EntityType, objectID, e.AttributeName)
			if err != nil {
				return false, fmt.Errorf("failed to get attribute for rule argument: %w", err)
			}
			argValues[i] = attributeValue
			
		case *LiteralExpression:
			// Just use the literal value
			argValues[i] = e.Value
			
		default:
			// For other expression types, evaluate them first
			result, err := g.evaluateExpression(ctx, argExpr, subjectType, subjectID, objectType, objectID, contextData)
			if err != nil {
				return false, err
			}
			argValues[i] = result
		}
	}
	
	// Verify that we have the correct number of arguments
	if len(argValues) != len(ruleDef.Parameters) {
		return false, fmt.Errorf("rule %s requires %d arguments, got %d", 
			rule.RuleName, len(ruleDef.Parameters), len(argValues))
	}
	
	// Create a parser for the rule expression
	ruleCtx := make(map[string]interface{})
	
	// Add arguments to rule context
	for i, param := range ruleDef.Parameters {
		ruleCtx[param.Name] = argValues[i]
	}
	
	// Parse the rule expression
	parser := NewConditionParser(ruleDef.Expression)
	ruleExpr, err := parser.Parse()
	if err != nil {
		return false, fmt.Errorf("failed to parse rule expression: %w", err)
	}
	
	// Evaluate the rule expression with the rule context
	return g.evaluateRuleExpression(ctx, ruleExpr, ruleCtx)
}

// evaluateRuleExpression evaluates a rule expression with the given context
func (g *IdentityGraph) evaluateRuleExpression(ctx context.Context, expr Expression, ruleCtx map[string]interface{}) (bool, error) {
	switch e := expr.(type) {
	case *AndExpression:
		// Evaluate left expression
		leftResult, err := g.evaluateRuleExpression(ctx, e.Left, ruleCtx)
		if err != nil {
			return false, err
		}
		
		// Short-circuit if left expression is false
		if !leftResult {
			return false, nil
		}
		
		// Evaluate right expression
		return g.evaluateRuleExpression(ctx, e.Right, ruleCtx)
		
	case *OrExpression:
		// Evaluate left expression
		leftResult, err := g.evaluateRuleExpression(ctx, e.Left, ruleCtx)
		if err != nil {
			return false, err
		}
		
		// Short-circuit if left expression is true
		if leftResult {
			return true, nil
		}
		
		// Evaluate right expression
		return g.evaluateRuleExpression(ctx, e.Right, ruleCtx)
		
	case *ComparisonExpression:
		// Evaluate left and right expressions to get their values
		leftValue, err := g.evaluateRuleValue(ctx, e.Left, ruleCtx)
		if err != nil {
			return false, err
		}
		
		rightValue, err := g.evaluateRuleValue(ctx, e.Right, ruleCtx)
		if err != nil {
			return false, err
		}
		
		// Perform the comparison based on the operator
		return g.compareValues(leftValue, e.Operator, rightValue)
		
	case *RelationExpression:
		// In rule context, relation expressions are treated as variable references
		value, exists := ruleCtx[e.RelationName]
		if !exists {
			return false, fmt.Errorf("variable not found in rule context: %s", e.RelationName)
		}
		
		// Convert value to boolean
		boolValue, ok := value.(bool)
		if ok {
			return boolValue, nil
		}
		
		// Non-boolean values are treated as "exists" checks
		return value != nil, nil
		
	case *ContextExpression:
		// Extract variable from context
		value, exists := ruleCtx[e.Path[0]]
		if !exists {
			return false, fmt.Errorf("variable not found in rule context: %s", e.Path[0])
		}
		
		// Convert value to boolean if possible
		boolValue, ok := value.(bool)
		if ok {
			return boolValue, nil
		}
		
		// Non-boolean values are treated as "exists" checks
		return value != nil, nil
		
	case *LiteralExpression:
		// Convert literal value to boolean if possible
		boolValue, ok := e.Value.(bool)
		if ok {
			return boolValue, nil
		}
		
		// Non-boolean values are treated as "exists" checks
		return e.Value != nil, nil
		
	default:
		return false, fmt.Errorf("unsupported expression type in rule: %T", expr)
	}
}

// evaluateRuleValue evaluates an expression to extract its value (not boolean result)
func (g *IdentityGraph) evaluateRuleValue(ctx context.Context, expr Expression, ruleCtx map[string]interface{}) (interface{}, error) {
	switch e := expr.(type) {
	case *RelationExpression:
		// Treat as variable reference
		value, exists := ruleCtx[e.RelationName]
		if !exists {
			return nil, fmt.Errorf("variable not found in rule context: %s", e.RelationName)
		}
		return value, nil
		
	case *ContextExpression:
		// Extract variable from context
		value, exists := ruleCtx[e.Path[0]]
		if !exists {
			return nil, fmt.Errorf("variable not found in rule context: %s", e.Path[0])
		}
		return value, nil
		
	case *LiteralExpression:
		return e.Value, nil
		
	default:
		// For complex expressions, evaluate them to a boolean result
		result, err := g.evaluateRuleExpression(ctx, expr, ruleCtx)
		if err != nil {
			return nil, err
		}
		return result, nil
	}
}

// EvaluateRule directly evaluates a rule with provided parameter values
func (g *IdentityGraph) EvaluateRule(ctx context.Context, ruleName string, params map[string]interface{}) (bool, error) {
	// Check rule cache for the rule
	g.ruleCacheMu.RLock()
	rule, exists := g.ruleCache[ruleName]
	g.ruleCacheMu.RUnlock()
	
	if !exists {
		// Try to load rule from database
		if err := g.loadRules(ctx); err != nil {
			return false, fmt.Errorf("failed to load rules: %w", err)
		}
		
		// Check again after loading
		g.ruleCacheMu.RLock()
		rule, exists = g.ruleCache[ruleName]
		g.ruleCacheMu.RUnlock()
		
		if !exists {
			return false, fmt.Errorf("rule not found: %s", ruleName)
		}
	}
	
	// Parse rule expression
	parser := NewConditionParser(rule.Expression)
	expr, err := parser.Parse()
	if err != nil {
		return false, fmt.Errorf("failed to parse rule expression: %w", err)
	}
	
	// Evaluate rule expression with parameters
	return g.evaluateRuleExpression(ctx, expr, params)
}

// compareValues compares two values based on the comparison operator
func (g *IdentityGraph) compareValues(left interface{}, op tokenType, right interface{}) (bool, error) {
	// Convert both values to a common type for comparison if needed
	leftFloat, leftIsFloat := toFloat64(left)
	rightFloat, rightIsFloat := toFloat64(right)
	
	// If both values can be converted to numbers, compare them as numbers
	if leftIsFloat && rightIsFloat {
		switch op {
		case tokenEQ:
			return leftFloat == rightFloat, nil
		case tokenNEQ:
			return leftFloat != rightFloat, nil
		case tokenGT:
			return leftFloat > rightFloat, nil
		case tokenGTE:
			return leftFloat >= rightFloat, nil
		case tokenLT:
			return leftFloat < rightFloat, nil
		case tokenLTE:
			return leftFloat <= rightFloat, nil
		default:
			return false, fmt.Errorf("unsupported comparison operator: %v", op)
		}
	}
	
	// Try string comparison if one of the values is a string
	leftStr, leftIsStr := left.(string)
	rightStr, rightIsStr := right.(string)
	
	if leftIsStr && rightIsStr {
		switch op {
		case tokenEQ:
			return leftStr == rightStr, nil
		case tokenNEQ:
			return leftStr != rightStr, nil
		case tokenGT:
			return leftStr > rightStr, nil
		case tokenGTE:
			return leftStr >= rightStr, nil
		case tokenLT:
			return leftStr < rightStr, nil
		case tokenLTE:
			return leftStr <= rightStr, nil
		default:
			return false, fmt.Errorf("unsupported comparison operator: %v", op)
		}
	}
	
	// For mixed types, == and != can still be applied
	if op == tokenEQ {
		// For equality, just compare if the values are equal
		return left == right, nil
	} else if op == tokenNEQ {
		// For inequality, just compare if the values are not equal
		return left != right, nil
	}
	
	// For other comparison operators with incompatible types, return an error
	return false, fmt.Errorf("cannot compare values of different types: %T and %T", left, right)
}

// getEntityAttribute gets an attribute value for an entity from the database
func (g *IdentityGraph) getEntityAttribute(ctx context.Context, entityType, entityID, attributeName string) (interface{}, error) {
	// Fetch the entity's attributes from the database using the JSONB properties field
	var propertiesJSON []byte
	err := g.Pool.QueryRow(ctx, `
		SELECT properties
		FROM entities
		WHERE type = $1 AND external_id = $2
	`, entityType, entityID).Scan(&propertiesJSON)
	
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("entity not found: %s:%s", entityType, entityID)
		}
		return nil, fmt.Errorf("failed to get entity properties: %w", err)
	}
	
	// Parse the properties JSON
	var properties map[string]interface{}
	if err := json.Unmarshal(propertiesJSON, &properties); err != nil {
		return nil, fmt.Errorf("failed to parse entity properties: %w", err)
	}
	
	// Extract the attribute value
	value, exists := properties[attributeName]
	if !exists {
		return nil, fmt.Errorf("attribute not found: %s", attributeName)
	}
	
	return value, nil
}

// Helper function to convert various numeric types to float64
func toFloat64(value interface{}) (float64, bool) {
	switch v := value.(type) {
	case float64:
		return v, true
	case float32:
		return float64(v), true
	case int:
		return float64(v), true
	case int64:
		return float64(v), true
	case int32:
		return float64(v), true
	case json.Number:
		f, err := v.Float64()
		return f, err == nil
	case string:
		f, err := strconv.ParseFloat(v, 64)
		return f, err == nil
	default:
		return 0, false
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
