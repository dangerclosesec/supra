package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	"github.com/dangerclosesec/supra/internal/model"
	"github.com/jackc/pgx/v5/pgxpool"
)

// AuthzAuditLogger handles audit logging for the authorization service
type AuthzAuditLogger struct {
	pool *pgxpool.Pool
}

// NewAuthzAuditLogger creates a new authorization audit logger
func NewAuthzAuditLogger(pool *pgxpool.Pool) *AuthzAuditLogger {
	return &AuthzAuditLogger{
		pool: pool,
	}
}

// LogPermissionCheck logs a permission check operation
func (l *AuthzAuditLogger) LogPermissionCheck(
	ctx context.Context,
	subject model.Subject,
	permission string,
	object model.Entity,
	result bool,
	contextData *map[string]interface{},
	req *http.Request,
) error {
	// Convert context data to JSON
	contextJSON, err := json.Marshal(contextData)
	if err != nil {
		log.Printf("Failed to marshal context data: %v", err)
		contextJSON = []byte("{}")
	}

	// Get request information
	requestID := ""
	clientIP := ""
	userAgent := ""
	if req != nil {
		requestID = req.Header.Get("X-Request-ID")
		clientIP = req.RemoteAddr
		userAgent = req.UserAgent()
	}

	// Insert audit log
	_, err = l.pool.Exec(ctx, `
		INSERT INTO authz_audit_logs (
			action_type, result, entity_type, entity_id, subject_type, subject_id,
			permission, context, request_id, client_ip, user_agent
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11
		)
	`,
		"permission_check", result, object.Type, object.ID,
		subject.Type, subject.ID, permission, contextJSON,
		requestID, clientIP, userAgent)

	if err != nil {
		log.Printf("Failed to log permission check: %v", err)
		return err
	}

	return nil
}

// LogEntityCreate logs an entity creation operation
func (l *AuthzAuditLogger) LogEntityCreate(
	ctx context.Context,
	entityType string,
	entityID string,
	attributes map[string]interface{},
	req *http.Request,
) error {
	// Convert attributes to JSON
	attributesJSON, err := json.Marshal(attributes)
	if err != nil {
		log.Printf("Failed to marshal entity attributes: %v", err)
		attributesJSON = []byte("{}")
	}

	// Get request information
	requestID := ""
	clientIP := ""
	userAgent := ""
	if req != nil {
		requestID = req.Header.Get("X-Request-ID")
		clientIP = req.RemoteAddr
		userAgent = req.UserAgent()
	}

	// Insert audit log
	_, err = l.pool.Exec(ctx, `
		INSERT INTO authz_audit_logs (
			action_type, entity_type, entity_id, context, request_id, client_ip, user_agent
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7
		)
	`,
		"entity_create", entityType, entityID, attributesJSON,
		requestID, clientIP, userAgent)

	if err != nil {
		log.Printf("Failed to log entity creation: %v", err)
		return err
	}

	return nil
}

// LogEntityDelete logs an entity deletion operation
func (l *AuthzAuditLogger) LogEntityDelete(
	ctx context.Context,
	entityType string,
	entityID string,
	req *http.Request,
) error {
	// Get request information
	requestID := ""
	clientIP := ""
	userAgent := ""
	if req != nil {
		requestID = req.Header.Get("X-Request-ID")
		clientIP = req.RemoteAddr
		userAgent = req.UserAgent()
	}

	// Insert audit log
	_, err := l.pool.Exec(ctx, `
		INSERT INTO authz_audit_logs (
			action_type, entity_type, entity_id, request_id, client_ip, user_agent
		) VALUES (
			$1, $2, $3, $4, $5, $6
		)
	`,
		"entity_delete", entityType, entityID,
		requestID, clientIP, userAgent)

	if err != nil {
		log.Printf("Failed to log entity deletion: %v", err)
		return err
	}

	return nil
}

// LogRelationCreate logs a relation creation operation
func (l *AuthzAuditLogger) LogRelationCreate(
	ctx context.Context,
	object model.Entity,
	relation string,
	subject model.Subject,
	req *http.Request,
) error {
	// Get request information
	requestID := ""
	clientIP := ""
	userAgent := ""
	if req != nil {
		requestID = req.Header.Get("X-Request-ID")
		clientIP = req.RemoteAddr
		userAgent = req.UserAgent()
	}

	// Insert audit log
	_, err := l.pool.Exec(ctx, `
		INSERT INTO authz_audit_logs (
			action_type, entity_type, entity_id, subject_type, subject_id, 
			relation, request_id, client_ip, user_agent
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9
		)
	`,
		"relation_create", object.Type, object.ID,
		subject.Type, subject.ID, relation,
		requestID, clientIP, userAgent)

	if err != nil {
		log.Printf("Failed to log relation creation: %v", err)
		return err
	}

	return nil
}

// LogRelationDelete logs a relation deletion operation
func (l *AuthzAuditLogger) LogRelationDelete(
	ctx context.Context,
	object model.Entity,
	relation string,
	subject model.Subject,
	req *http.Request,
) error {
	// Get request information
	requestID := ""
	clientIP := ""
	userAgent := ""
	if req != nil {
		requestID = req.Header.Get("X-Request-ID")
		clientIP = req.RemoteAddr
		userAgent = req.UserAgent()
	}

	// Insert audit log
	_, err := l.pool.Exec(ctx, `
		INSERT INTO authz_audit_logs (
			action_type, entity_type, entity_id, subject_type, subject_id, 
			relation, request_id, client_ip, user_agent
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9
		)
	`,
		"relation_delete", object.Type, object.ID,
		subject.Type, subject.ID, relation,
		requestID, clientIP, userAgent)

	if err != nil {
		log.Printf("Failed to log relation deletion: %v", err)
		return err
	}

	return nil
}
