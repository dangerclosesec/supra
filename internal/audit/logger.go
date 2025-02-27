package audit

import (
	"context"
	"net/http"

	"github.com/dangerclosesec/supra/internal/model"
)

// Logger defines the interface for auditing operations
type Logger interface {
	// LogPermissionCheck logs a permission check operation
	LogPermissionCheck(
		ctx context.Context,
		subject model.Subject,
		permission string,
		object model.Entity,
		result bool,
		contextData map[string]interface{},
		req *http.Request,
	) error

	// LogEntityCreate logs an entity creation operation
	LogEntityCreate(
		ctx context.Context,
		entityType string,
		entityID string,
		attributes map[string]interface{},
		req *http.Request,
	) error

	// LogEntityDelete logs an entity deletion operation
	LogEntityDelete(
		ctx context.Context,
		entityType string,
		entityID string,
		req *http.Request,
	) error

	// LogRelationCreate logs a relation creation operation
	LogRelationCreate(
		ctx context.Context,
		object model.Entity,
		relation string,
		subject model.Subject,
		req *http.Request,
	) error

	// LogRelationDelete logs a relation deletion operation
	LogRelationDelete(
		ctx context.Context,
		object model.Entity,
		relation string,
		subject model.Subject,
		req *http.Request,
	) error
}

// NoOpLogger is a logger that does nothing
type NoOpLogger struct{}

// LogPermissionCheck implements Logger.LogPermissionCheck
func (l *NoOpLogger) LogPermissionCheck(
	ctx context.Context,
	subject model.Subject,
	permission string,
	object model.Entity,
	result bool,
	contextData map[string]interface{},
	req *http.Request,
) error {
	return nil
}

// LogEntityCreate implements Logger.LogEntityCreate
func (l *NoOpLogger) LogEntityCreate(
	ctx context.Context,
	entityType string,
	entityID string,
	attributes map[string]interface{},
	req *http.Request,
) error {
	return nil
}

// LogEntityDelete implements Logger.LogEntityDelete
func (l *NoOpLogger) LogEntityDelete(
	ctx context.Context,
	entityType string,
	entityID string,
	req *http.Request,
) error {
	return nil
}

// LogRelationCreate implements Logger.LogRelationCreate
func (l *NoOpLogger) LogRelationCreate(
	ctx context.Context,
	object model.Entity,
	relation string,
	subject model.Subject,
	req *http.Request,
) error {
	return nil
}

// LogRelationDelete implements Logger.LogRelationDelete
func (l *NoOpLogger) LogRelationDelete(
	ctx context.Context,
	object model.Entity,
	relation string,
	subject model.Subject,
	req *http.Request,
) error {
	return nil
}