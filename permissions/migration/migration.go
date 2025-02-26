// File: migration/migration.go
package migration

import (
	"database/sql"
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

// LoadCurrentModel loads the current permission model from the database
func (m *Migrator) LoadCurrentModel() (*model.PermissionModel, error) {
	permModel := model.NewPermissionModel()

	rows, err := m.DB.Query(`
		SELECT entity_type, permission_name, condition_expression
		FROM permission_definitions
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	entityMap := make(map[string]*model.Entity)

	for rows.Next() {
		var entityType, permName, expr string
		if err := rows.Scan(&entityType, &permName, &expr); err != nil {
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

	return permModel, nil
}

// applyModelInTransaction applies the model changes within a transaction
func (m *Migrator) applyModelInTransaction(tx *sql.Tx, model *model.PermissionModel) error {
	// Clear existing permissions
	_, err := tx.Exec(`DELETE FROM permission_definitions`)
	if err != nil {
		return fmt.Errorf("failed to clear permissions: %w", err)
	}

	// Insert entities and permissions
	for _, entity := range model.Entities {
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
