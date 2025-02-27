// File: migration/migration.go
package migration

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/dangerclosesec/supra/permissions/model"
	_ "github.com/lib/pq"
)

// Migrator handles database migrations for permission models
type Migrator struct {
	DB *sql.DB
}

// NewMigrator creates a new migrator
func NewMigrator(db *sql.DB) *Migrator {
	return &Migrator{DB: db}
}

// InitializeSchema initializes the database schema
func (m *Migrator) InitializeSchema() error {
	// Create the schema if it doesn't exist
	_, err := m.DB.Exec(`
	-- Create tables if they don't exist
	CREATE TABLE IF NOT EXISTS entity_types (
		id SERIAL PRIMARY KEY,
		name TEXT NOT NULL UNIQUE,
		created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
	);

	CREATE TABLE IF NOT EXISTS permission_definitions (
		id SERIAL PRIMARY KEY,
		entity_type TEXT NOT NULL,
		permission_name TEXT NOT NULL,
		condition_expression TEXT NOT NULL,
		description TEXT,
		created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
		UNIQUE(entity_type, permission_name)
	);

	CREATE TABLE IF NOT EXISTS permission_versions (
		id SERIAL PRIMARY KEY,
		version INT NOT NULL,
		description TEXT,
		source_file TEXT,
		applied_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
	);

	CREATE TABLE IF NOT EXISTS migration_history (
		id SERIAL PRIMARY KEY,
		version INT NOT NULL,
		created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
		success BOOLEAN NOT NULL,
		errors TEXT,
		diff TEXT
	);

	CREATE INDEX IF NOT EXISTS idx_entity_type_name ON permission_definitions(entity_type);
	`)

	return err
}

// GetCurrentVersion gets the current permission model version
func (m *Migrator) GetCurrentVersion() (int, error) {
	var version int
	err := m.DB.QueryRow(`
		SELECT COALESCE(MAX(version), 0) FROM permission_versions
	`).Scan(&version)
	return version, err
}

// ApplyMigration applies a permission model to the database
func (m *Migrator) ApplyMigration(model *model.PermissionModel, description string) (string, error) {
	// Get current version
	currentVersion, err := m.GetCurrentVersion()
	if err != nil {
		return "", fmt.Errorf("failed to get current version: %w", err)
	}

	// Load current model from database for diffing
	currentModel, err := m.LoadCurrentModel()
	if err != nil {
		return "", fmt.Errorf("failed to load current model: %w", err)
	}

	// Generate diff
	diff := GenerateDiff(currentModel, model)
	diffText := diff.String()

	// Check if there are actual changes
	hasChanges := !diff.IsEmpty()

	// If no changes, skip the migration process
	if !hasChanges {
		return "No changes detected. Migration skipped.", nil
	}

	// Calculate new version only if changes are detected
	newVersion := currentVersion + 1

	// Start transaction
	tx, err := m.DB.Begin()
	if err != nil {
		return "", fmt.Errorf("failed to start transaction: %w", err)
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Apply changes
	err = m.applyModelInTransaction(tx, model)
	if err != nil {
		tx.Rollback()
		// Record failed migration
		m.recordMigrationHistory(newVersion, false, err.Error(), diffText)
		return "", fmt.Errorf("failed to apply model: %w", err)
	}

	// Record version
	_, err = tx.Exec(`
		INSERT INTO permission_versions (version, description, source_file)
		VALUES ($1, $2, $3)
	`, newVersion, description, model.Source)
	if err != nil {
		tx.Rollback()
		return "", fmt.Errorf("failed to record version: %w", err)
	}

	// Commit transaction
	err = tx.Commit()
	if err != nil {
		return "", fmt.Errorf("failed to commit transaction: %w", err)
	}

	// Record successful migration
	m.recordMigrationHistory(newVersion, true, "", diffText)

	return diffText, nil
}

// convertRuleParameters converts model.Rule.Parameters to a format suitable for JSON storage
func convertRuleParameters(rule *model.Rule) []map[string]string {
	params := make([]map[string]string, len(rule.Parameters))
	for i, param := range rule.Parameters {
		params[i] = map[string]string{
			"name":      param.Name,
			"data_type": string(param.DataType),
		}
	}
	return params
}

// LoadCurrentModel loads the current permission model from the database
func (m *Migrator) LoadCurrentModel() (*model.PermissionModel, error) {
	permModel := model.NewPermissionModel()

	// Load permissions
	permRows, err := m.DB.Query(`
		SELECT entity_type, permission_name, condition_expression
		FROM permission_definitions
	`)
	if err != nil {
		return nil, err
	}
	defer permRows.Close()

	entityMap := make(map[string]*model.Entity)

	for permRows.Next() {
		var entityType, permName, expr string
		if err := permRows.Scan(&entityType, &permName, &expr); err != nil {
			return nil, err
		}

		// Get or create entity
		if _, ok := entityMap[entityType]; !ok {
			entityMap[entityType] = &model.Entity{
				Name: entityType,
			}
			permModel.AddEntity(entityMap[entityType])
		}

		// Add permission
		entity := entityMap[entityType]
		entity.Permissions = append(entity.Permissions, model.Permission{
			Name:       permName,
			Expression: expr,
		})
	}
	
	// Check if rule_definitions table exists
	var ruleTableExists bool
	err = m.DB.QueryRow(`
		SELECT EXISTS (
			SELECT FROM information_schema.tables 
			WHERE table_schema = 'public'
			AND table_name = 'rule_definitions'
		)
	`).Scan(&ruleTableExists)
	
	if err != nil {
		// If we can't check, we'll assume no rules
		log.Printf("Warning: Could not check if rule_definitions table exists: %v", err)
	} else if ruleTableExists {
		// Load rules if the table exists
		ruleRows, err := m.DB.Query(`
			SELECT rule_name, parameters, expression
			FROM rule_definitions
		`)
		if err != nil {
			log.Printf("Warning: Could not load rules: %v", err)
		} else {
			defer ruleRows.Close()
			
			for ruleRows.Next() {
				var ruleName string
				var parametersJSON []byte
				var expression string
				
				if err := ruleRows.Scan(&ruleName, &parametersJSON, &expression); err != nil {
					log.Printf("Warning: Could not scan rule row: %v", err)
					continue
				}
				
				// Parse parameters
				var paramsData []map[string]string
				if err := json.Unmarshal(parametersJSON, &paramsData); err != nil {
					log.Printf("Warning: Could not parse rule parameters: %v", err)
					continue
				}
				
				// Convert parameters to model.RuleParameter
				params := make([]model.RuleParameter, len(paramsData))
				for i, p := range paramsData {
					params[i] = model.RuleParameter{
						Name:     p["name"],
						DataType: model.AttributeDataType(p["data_type"]),
					}
				}
				
				// Create rule
				rule := &model.Rule{
					Name:       ruleName,
					Parameters: params,
					Expression: expression,
				}
				
				// Add rule to model
				permModel.AddRule(rule)
			}
		}
	}

	return permModel, nil
}

// applyModelInTransaction applies the model changes within a transaction
func (m *Migrator) applyModelInTransaction(tx *sql.Tx, model *model.PermissionModel) error {
	// Clear existing permissions
	_, err := tx.Exec(`DELETE FROM permission_definitions`)
	if err != nil {
		return fmt.Errorf("failed to clear permissions: %w", err)
	}

	// Check if rule_definitions table exists
	var ruleTableExists bool
	err = tx.QueryRow(`
		SELECT EXISTS (
			SELECT FROM information_schema.tables 
			WHERE table_schema = 'public'
			AND table_name = 'rule_definitions'
		)
	`).Scan(&ruleTableExists)

	if err != nil {
		return fmt.Errorf("failed to check if rule_definitions table exists: %w", err)
	}

	// Create rule_definitions table if it doesn't exist
	if !ruleTableExists {
		_, err = tx.Exec(`
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
	} else {
		// Clear existing rules if table exists
		_, err = tx.Exec(`DELETE FROM rule_definitions`)
		if err != nil {
			return fmt.Errorf("failed to clear rules: %w", err)
		}
	}

	// Insert global rules
	for _, rule := range model.Rules {
		// Skip nil rules (shouldn't happen but just in case)
		if rule == nil {
			continue
		}

		// Convert rule parameters to JSON
		parametersJSON, err := json.Marshal(convertRuleParameters(rule))
		if err != nil {
			return fmt.Errorf("failed to marshal rule parameters: %w", err)
		}

		// Insert rule
		_, err = tx.Exec(`
			INSERT INTO rule_definitions (rule_name, parameters, expression, description)
			VALUES ($1, $2, $3, $4)
		`, rule.Name, parametersJSON, rule.Expression, "")
		if err != nil {
			return fmt.Errorf("failed to insert rule definition: %w", err)
		}
	}

	// Insert entities and permissions
	for _, entity := range model.Entities {
		// Insert entity-level rules
		for i := range entity.Rules {
			rule := &entity.Rules[i]
			
			// Convert rule parameters to JSON
			parametersJSON, err := json.Marshal(convertRuleParameters(rule))
			if err != nil {
				return fmt.Errorf("failed to marshal rule parameters: %w", err)
			}

			// Check if rule already exists in global rules (skip if it does)
			var exists bool
			err = tx.QueryRow(`
				SELECT EXISTS (
					SELECT 1 FROM rule_definitions WHERE rule_name = $1
				)
			`, rule.Name).Scan(&exists)

			if err != nil {
				return fmt.Errorf("failed to check if rule exists: %w", err)
			}

			if !exists {
				// Insert rule
				_, err = tx.Exec(`
					INSERT INTO rule_definitions (rule_name, parameters, expression, description)
					VALUES ($1, $2, $3, $4)
				`, rule.Name, parametersJSON, rule.Expression, "")
				if err != nil {
					return fmt.Errorf("failed to insert entity rule definition: %w", err)
				}
			}
		}

		// Insert permissions
		for _, perm := range entity.Permissions {
			_, err := tx.Exec(`
				INSERT INTO permission_definitions (entity_type, permission_name, condition_expression, description)
				VALUES ($1, $2, $3, $4)
			`, entity.Name, perm.Name, perm.Expression, strings.Join(perm.Comments, "\n"))
			if err != nil {
				return fmt.Errorf("failed to insert permission %s.%s: %w", entity.Name, perm.Name, err)
			}
		}
	}

	return nil
}

// recordMigrationHistory records migration history
func (m *Migrator) recordMigrationHistory(version int, success bool, errorMsg string, diff string) {
	_, err := m.DB.Exec(`
		INSERT INTO migration_history (version, success, errors, diff)
		VALUES ($1, $2, $3, $4)
	`, version, success, errorMsg, diff)
	if err != nil {
		log.Printf("Failed to record migration history: %v", err)
	}
}
