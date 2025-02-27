package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/dangerclosesec/supra/internal/auth/graph"
	"github.com/dangerclosesec/supra/permissions/model"
	"github.com/dangerclosesec/supra/permissions/parser"
)

// RuleDefinitionResponse represents a rule definition for API responses
type RuleDefinitionResponse struct {
	ID          int64                 `json:"id"`
	Name        string                `json:"name"`
	Parameters  []ParameterDefinition `json:"parameters"`
	Expression  string                `json:"expression"`
	Description string                `json:"description,omitempty"`
	CreatedAt   string                `json:"created_at,omitempty"`
}

// ParameterDefinition represents a parameter in a rule
type ParameterDefinition struct {
	Name     string `json:"name"`
	DataType string `json:"data_type"`
}

// SyncPermissionModel loads a permission model from a file and syncs rules to the database
func SyncPermissionModel(ctx context.Context, pool interface{}, filePath string) error {
	// Read the permission model file
	content, err := readFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read permission model file: %w", err)
	}

	// Parse the permission model
	l := parser.NewLexer(string(content))
	p := parser.NewParser(l)
	permModel := p.ParsePermissionModel()

	if len(p.Errors()) > 0 {
		// Log parsing errors but continue with partial model
		for _, err := range p.Errors() {
			log.Printf("Permission model parsing error: %s", err)
		}
	}

	// Sync rules to database
	err = SyncRulesToDatabase(ctx, pool, permModel)
	if err != nil {
		return fmt.Errorf("failed to sync rules to database: %w", err)
	}

	return nil
}

// SyncRulesToDatabase syncs rule definitions from the model to the database
func SyncRulesToDatabase(ctx context.Context, pool interface{}, permModel *model.PermissionModel) error {
	pgPool, ok := pool.(*graph.IdentityGraph)
	if !ok {
		return fmt.Errorf("pool is not a valid IdentityGraph")
	}

	// First, check if the rules table exists
	var tableExists bool
	err := pgPool.Pool.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT FROM information_schema.tables 
			WHERE table_schema = 'public'
			AND table_name = 'rule_definitions'
		)
	`).Scan(&tableExists)

	if err != nil {
		// If we can't even check if the table exists, we have a serious DB problem
		return fmt.Errorf("failed to check if rule_definitions table exists: %w", err)
	}

	if !tableExists {
		log.Println("Rule definitions table does not exist. Creating it now.")
		_, err := pgPool.Pool.Exec(ctx, `
			CREATE TABLE rule_definitions (
				id SERIAL PRIMARY KEY,
				rule_name TEXT NOT NULL UNIQUE,
				parameters JSONB NOT NULL,
				expression TEXT NOT NULL,
				description TEXT,
				created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
			)
		`)
		if err != nil {
			return fmt.Errorf("failed to create rule_definitions table: %w", err)
		}
		log.Println("Created rule_definitions table")
	}

	// If model has no rules, we're done
	if len(permModel.Rules) == 0 {
		log.Println("No rules found in permission model")
		return nil
	}

	log.Printf("Found %d rules to sync to database", len(permModel.Rules))

	// Process each rule in the model
	for _, rule := range permModel.Rules {
		log.Printf("Processing rule: %s", rule.Name)

		// Convert rule parameters to JSON
		parametersJSON, err := json.Marshal(convertRuleParameters(rule))
		if err != nil {
			return fmt.Errorf("failed to marshal rule parameters for rule %s: %w", rule.Name, err)
		}

		// Check if the rule already exists
		var ruleID int64
		err = pgPool.Pool.QueryRow(ctx, `
			SELECT id FROM rule_definitions WHERE rule_name = $1
		`, rule.Name).Scan(&ruleID)

		if err == nil {
			// Rule exists, update it
			log.Printf("Updating existing rule: %s", rule.Name)
			_, err = pgPool.Pool.Exec(ctx, `
				UPDATE rule_definitions 
				SET parameters = $1, expression = $2
				WHERE id = $3
			`, parametersJSON, rule.Expression, ruleID)
			if err != nil {
				return fmt.Errorf("failed to update rule definition %s: %w", rule.Name, err)
			}
		} else if err == sql.ErrNoRows {
			// Rule doesn't exist, insert it
			log.Printf("Creating new rule: %s", rule.Name)
			_, err = pgPool.Pool.Exec(ctx, `
				INSERT INTO rule_definitions (rule_name, parameters, expression, description)
				VALUES ($1, $2, $3, $4)
			`, rule.Name, parametersJSON, rule.Expression, "")
			if err != nil {
				return fmt.Errorf("failed to insert rule definition %s: %w", rule.Name, err)
			}
		} else {
			// Some other error
			return fmt.Errorf("failed to check if rule %s exists: %w", rule.Name, err)
		}
	}

	log.Println("Rule definitions synchronized successfully")
	return nil
}

// convertRuleParameters converts model.RuleParameter to ParameterDefinition
func convertRuleParameters(rule *model.Rule) []ParameterDefinition {
	params := make([]ParameterDefinition, len(rule.Parameters))
	for i, param := range rule.Parameters {
		params[i] = ParameterDefinition{
			Name:     param.Name,
			DataType: string(param.DataType),
		}
	}
	return params
}

// TestRuleRequest represents a request to test a rule with parameters
type TestRuleRequest struct {
	RuleName   string                 `json:"rule_name"`
	Parameters map[string]interface{} `json:"parameters"`
}

// TestRuleResponse represents the result of a rule test
type TestRuleResponse struct {
	Result bool   `json:"result"`
	Error  string `json:"error,omitempty"`
}

// addRuleEndpoints adds endpoints for rule management
func (s *AuthzService) addRuleEndpoints(mux *http.ServeMux) {
	// Endpoint to get all rule definitions
	mux.HandleFunc("/api/rule-definitions", func(w http.ResponseWriter, r *http.Request) {
		// Only allow GET requests
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Get context with timeout
		ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
		defer cancel()

		// Query the database for all rule definitions
		rows, err := s.graph.Pool.Query(ctx, `
			SELECT id, rule_name, parameters, expression, description, created_at
			FROM rule_definitions
			ORDER BY rule_name
		`)
		if err != nil {
			log.Printf("Error retrieving rule definitions: %v", err)
			http.Error(w, "Error retrieving rule definitions", http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		// Process query results
		var rules []RuleDefinitionResponse
		for rows.Next() {
			var rule RuleDefinitionResponse
			var parametersJSON []byte
			var createdAt interface{} // Handle potential NULL value
			err := rows.Scan(
				&rule.ID,
				&rule.Name,
				&parametersJSON,
				&rule.Expression,
				&rule.Description,
				&createdAt,
			)
			if err != nil {
				log.Printf("Error scanning rule definition: %v", err)
				continue
			}

			// Parse parameters JSON
			if err := json.Unmarshal(parametersJSON, &rule.Parameters); err != nil {
				log.Printf("Error parsing rule parameters: %v", err)
				rule.Parameters = []ParameterDefinition{}
			}

			// Convert createdAt to string if not nil
			if createdAt != nil {
				rule.CreatedAt = createdAt.(time.Time).Format(time.RFC3339)
			}

			rules = append(rules, rule)
		}

		// Return the results
		jsonResponse(w, rules, http.StatusOK)
	})

	// Endpoint to test a rule with parameters
	mux.HandleFunc("/api/test-rule", func(w http.ResponseWriter, r *http.Request) {
		// Only allow POST requests
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Parse request
		var req TestRuleRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request format", http.StatusBadRequest)
			return
		}

		// Validate request
		if req.RuleName == "" {
			http.Error(w, "Rule name is required", http.StatusBadRequest)
			return
		}

		// Get context with timeout
		ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
		defer cancel()

		// Get rule definition from database
		var expression string
		var parametersJSON []byte
		err := s.graph.Pool.QueryRow(ctx, `
			SELECT expression, parameters
			FROM rule_definitions
			WHERE rule_name = $1
		`, req.RuleName).Scan(&expression, &parametersJSON)

		if err != nil {
			if err == sql.ErrNoRows {
				http.Error(w, "Rule not found", http.StatusNotFound)
			} else {
				log.Printf("Error retrieving rule definition: %v", err)
				http.Error(w, "Error retrieving rule definition", http.StatusInternalServerError)
			}
			return
		}

		// Parse parameters
		var paramDefs []ParameterDefinition
		if err := json.Unmarshal(parametersJSON, &paramDefs); err != nil {
			log.Printf("Error parsing rule parameters: %v", err)
			http.Error(w, "Error parsing rule parameters", http.StatusInternalServerError)
			return
		}

		// Validate parameter values
		for _, param := range paramDefs {
			val, exists := req.Parameters[param.Name]
			if !exists {
				jsonResponse(w, TestRuleResponse{
					Result: false,
					Error:  fmt.Sprintf("Missing parameter: %s", param.Name),
				}, http.StatusBadRequest)
				return
			}

			// Convert values to the right type if needed
			switch param.DataType {
			case "integer":
				// Convert string or float to int if needed
				switch v := val.(type) {
				case string:
					var intVal int64
					if err := json.Unmarshal([]byte(v), &intVal); err != nil {
						jsonResponse(w, TestRuleResponse{
							Result: false,
							Error:  fmt.Sprintf("Invalid integer value for parameter %s: %v", param.Name, val),
						}, http.StatusBadRequest)
						return
					}
					req.Parameters[param.Name] = intVal
				case float64:
					req.Parameters[param.Name] = int64(v)
				}
			case "float", "double":
				// Convert string to float if needed
				if strVal, ok := val.(string); ok {
					var floatVal float64
					if err := json.Unmarshal([]byte(strVal), &floatVal); err != nil {
						jsonResponse(w, TestRuleResponse{
							Result: false,
							Error:  fmt.Sprintf("Invalid float value for parameter %s: %v", param.Name, val),
						}, http.StatusBadRequest)
						return
					}
					req.Parameters[param.Name] = floatVal
				}
			case "boolean":
				// Convert string to bool if needed
				if strVal, ok := val.(string); ok {
					boolVal := strings.ToLower(strVal) == "true"
					req.Parameters[param.Name] = boolVal
				}
			}
		}

		// Evaluate rule using internal auth graph
		result, err := s.graph.EvaluateRule(ctx, req.RuleName, req.Parameters)
		if err != nil {
			log.Printf("Error evaluating rule: %v", err)
			jsonResponse(w, TestRuleResponse{
				Result: false,
				Error:  fmt.Sprintf("Error evaluating rule: %v", err),
			}, http.StatusInternalServerError)
			return
		}

		// Return the result
		jsonResponse(w, TestRuleResponse{
			Result: result,
		}, http.StatusOK)
	})
}

// Helper function to read a file
func readFile(path string) ([]byte, error) {
	// Use os.ReadFile to read the file contents
	return os.ReadFile(path)
}
