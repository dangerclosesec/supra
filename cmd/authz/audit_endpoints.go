package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// Define request/response types for audit log queries
type AuditLogQueryParams struct {
	ActionType  string     `json:"action_type"`
	EntityType  string     `json:"entity_type"`
	EntityID    string     `json:"entity_id"`
	SubjectType string     `json:"subject_type"`
	SubjectID   string     `json:"subject_id"`
	Result      *bool      `json:"result"`
	StartTime   *time.Time `json:"start_time"`
	EndTime     *time.Time `json:"end_time"`
	Limit       int        `json:"limit"`
	Offset      int        `json:"offset"`
}

type AuditLogEntry struct {
	ID          uuid.UUID              `json:"id"`
	Timestamp   time.Time              `json:"timestamp"`
	ActionType  string                 `json:"action_type"`
	Result      *bool                  `json:"result"`
	EntityType  string                 `json:"entity_type"`
	EntityID    string                 `json:"entity_id"`
	SubjectType string                 `json:"subject_type"`
	SubjectID   string                 `json:"subject_id"`
	Relation    string                 `json:"relation,omitempty"`
	Permission  string                 `json:"permission,omitempty"`
	Context     map[string]interface{} `json:"context,omitempty"`
	RequestID   string                 `json:"request_id,omitempty"`
	ClientIP    string                 `json:"client_ip,omitempty"`
	UserAgent   string                 `json:"user_agent,omitempty"`
	CreatedAt   time.Time              `json:"created_at"`
}

type AuditLogListResponse struct {
	Logs  []AuditLogEntry `json:"logs"`
	Total int64           `json:"total"`
}

// addAuditLogEndpoints registers endpoints for retrieving audit logs
func (s *AuthzService) addAuditLogEndpoints(mux *http.ServeMux) {
	// Handler for retrieving audit logs with filtering
	mux.HandleFunc("/api/audit/logs", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			s.getAuditLogsHandler(w, r)
		case http.MethodOptions:
			// Handle CORS preflight
			w.WriteHeader(http.StatusOK)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// Handler for retrieving a specific audit log by ID
	mux.HandleFunc("/api/audit/logs/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Extract ID from path
		parts := strings.Split(r.URL.Path, "/")
		if len(parts) != 5 {
			http.Error(w, "Invalid log ID", http.StatusBadRequest)
			return
		}

		logID := parts[4]
		if logID == "" {
			http.Error(w, "Log ID is required", http.StatusBadRequest)
			return
		}

		// Parse UUID
		id, err := uuid.Parse(logID)
		if err != nil {
			http.Error(w, "Invalid log ID format", http.StatusBadRequest)
			return
		}

		s.getAuditLogByIDHandler(w, r, id)
	})
}

// getAuditLogsHandler handles retrieving audit logs with filtering
func (s *AuthzService) getAuditLogsHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	// Parse query parameters
	query := r.URL.Query()
	params := AuditLogQueryParams{
		ActionType:  query.Get("action_type"),
		EntityType:  query.Get("entity_type"),
		EntityID:    query.Get("entity_id"),
		SubjectType: query.Get("subject_type"),
		SubjectID:   query.Get("subject_id"),
	}

	// Parse boolean result
	if resultStr := query.Get("result"); resultStr != "" {
		result, err := strconv.ParseBool(resultStr)
		if err != nil {
			http.Error(w, "Invalid result parameter, must be true or false", http.StatusBadRequest)
			return
		}
		params.Result = &result
	}

	// Parse time parameters
	if startTimeStr := query.Get("start_time"); startTimeStr != "" {
		startTime, err := time.Parse(time.RFC3339, startTimeStr)
		if err != nil {
			http.Error(w, "Invalid start_time format, use RFC3339", http.StatusBadRequest)
			return
		}
		params.StartTime = &startTime
	}

	if endTimeStr := query.Get("end_time"); endTimeStr != "" {
		endTime, err := time.Parse(time.RFC3339, endTimeStr)
		if err != nil {
			http.Error(w, "Invalid end_time format, use RFC3339", http.StatusBadRequest)
			return
		}
		params.EndTime = &endTime
	}

	// Parse pagination parameters
	if limitStr := query.Get("limit"); limitStr != "" {
		limit, err := strconv.Atoi(limitStr)
		if err != nil || limit < 1 {
			http.Error(w, "Invalid limit parameter", http.StatusBadRequest)
			return
		}
		params.Limit = limit
	} else {
		params.Limit = 100 // Default limit
	}

	if offsetStr := query.Get("offset"); offsetStr != "" {
		offset, err := strconv.Atoi(offsetStr)
		if err != nil || offset < 0 {
			http.Error(w, "Invalid offset parameter", http.StatusBadRequest)
			return
		}
		params.Offset = offset
	} else {
		params.Offset = 0 // Default offset
	}

	// Query audit logs
	logs, total, err := s.queryAuditLogs(ctx, params)
	if err != nil {
		log.Printf("Error querying audit logs: %v", err)
		http.Error(w, "Failed to retrieve audit logs", http.StatusInternalServerError)
		return
	}

	// Return response
	jsonResponse(w, AuditLogListResponse{
		Logs:  logs,
		Total: total,
	}, http.StatusOK)
}

// getAuditLogByIDHandler handles retrieving a specific audit log by ID
func (s *AuthzService) getAuditLogByIDHandler(w http.ResponseWriter, r *http.Request, id uuid.UUID) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	entry, err := s.getAuditLogByID(ctx, id)
	if err != nil {
		if err == pgx.ErrNoRows {
			http.Error(w, "Log not found", http.StatusNotFound)
		} else {
			log.Printf("Error retrieving audit log: %v", err)
			http.Error(w, "Failed to retrieve audit log", http.StatusInternalServerError)
		}
		return
	}

	jsonResponse(w, entry, http.StatusOK)
}

// queryAuditLogs retrieves audit logs with filtering
func (s *AuthzService) queryAuditLogs(ctx context.Context, params AuditLogQueryParams) ([]AuditLogEntry, int64, error) {
	var total int64

	// Build query conditions and parameters
	conditions := []string{}
	queryParams := []interface{}{}
	paramIndex := 1

	// Add conditions based on parameters
	if params.ActionType != "" {
		conditions = append(conditions, fmt.Sprintf("action_type = $%d", paramIndex))
		queryParams = append(queryParams, params.ActionType)
		paramIndex++
	}

	if params.EntityType != "" {
		conditions = append(conditions, fmt.Sprintf("entity_type = $%d", paramIndex))
		queryParams = append(queryParams, params.EntityType)
		paramIndex++
	}

	if params.EntityID != "" {
		conditions = append(conditions, fmt.Sprintf("entity_id = $%d", paramIndex))
		queryParams = append(queryParams, params.EntityID)
		paramIndex++
	}

	if params.SubjectType != "" {
		conditions = append(conditions, fmt.Sprintf("subject_type = $%d", paramIndex))
		queryParams = append(queryParams, params.SubjectType)
		paramIndex++
	}

	if params.SubjectID != "" {
		conditions = append(conditions, fmt.Sprintf("subject_id = $%d", paramIndex))
		queryParams = append(queryParams, params.SubjectID)
		paramIndex++
	}

	if params.Result != nil {
		conditions = append(conditions, fmt.Sprintf("result = $%d", paramIndex))
		queryParams = append(queryParams, *params.Result)
		paramIndex++
	}

	if params.StartTime != nil {
		conditions = append(conditions, fmt.Sprintf("timestamp >= $%d", paramIndex))
		queryParams = append(queryParams, *params.StartTime)
		paramIndex++
	}

	if params.EndTime != nil {
		conditions = append(conditions, fmt.Sprintf("timestamp <= $%d", paramIndex))
		queryParams = append(queryParams, *params.EndTime)
		paramIndex++
	}

	// Construct WHERE clause
	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}

	// Get total count
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM authz_audit_logs %s", whereClause)
	err := s.graph.Pool.QueryRow(ctx, countQuery, queryParams...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count audit logs: %w", err)
	}

	// Build full query with pagination
	query := fmt.Sprintf(`
		SELECT 
			id, timestamp, action_type, result, 
			entity_type, entity_id, subject_type, subject_id,
			relation, permission, context, request_id, 
			client_ip, user_agent, created_at
		FROM 
			authz_audit_logs 
		%s
		ORDER BY 
			timestamp DESC
		LIMIT $%d OFFSET $%d
	`, whereClause, paramIndex, paramIndex+1)

	// Add pagination parameters
	queryParams = append(queryParams, params.Limit)
	queryParams = append(queryParams, params.Offset)

	// Execute query
	rows, err := s.graph.Pool.Query(ctx, query, queryParams...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query audit logs: %w", err)
	}
	defer rows.Close()

	// Parse results
	var logs []AuditLogEntry
	for rows.Next() {
		var log AuditLogEntry
		var contextBytes []byte
		
		// Use sql.NullString for fields that may be NULL
		var subjectType, subjectID, relation, permission, requestID, clientIP, userAgent sql.NullString

		err := rows.Scan(
			&log.ID, &log.Timestamp, &log.ActionType, &log.Result,
			&log.EntityType, &log.EntityID, &subjectType, &subjectID,
			&relation, &permission, &contextBytes, &requestID,
			&clientIP, &userAgent, &log.CreatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan audit log: %w", err)
		}
		
		// Convert NullString fields to string (empty string if NULL)
		if subjectType.Valid {
			log.SubjectType = subjectType.String
		}
		if subjectID.Valid {
			log.SubjectID = subjectID.String
		}
		if relation.Valid {
			log.Relation = relation.String
		}
		if permission.Valid {
			log.Permission = permission.String
		}
		if requestID.Valid {
			log.RequestID = requestID.String
		}
		if clientIP.Valid {
			log.ClientIP = clientIP.String
		}
		if userAgent.Valid {
			log.UserAgent = userAgent.String
		}

		// Parse context JSON
		if len(contextBytes) > 0 {
			err = json.Unmarshal(contextBytes, &log.Context)
			if err != nil {
				log.Context = map[string]interface{}{
					"error":       "Failed to parse context data",
					"raw_context": string(contextBytes),
				}
			}
		} else {
			log.Context = map[string]interface{}{}
		}

		logs = append(logs, log)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("error iterating audit logs: %w", err)
	}

	return logs, total, nil
}

// getAuditLogByID retrieves a specific audit log by ID
func (s *AuthzService) getAuditLogByID(ctx context.Context, id uuid.UUID) (*AuditLogEntry, error) {
	query := `
		SELECT 
			id, timestamp, action_type, result, 
			entity_type, entity_id, subject_type, subject_id,
			relation, permission, context, request_id, 
			client_ip, user_agent, created_at
		FROM 
			authz_audit_logs 
		WHERE 
			id = $1
	`

	var log AuditLogEntry
	var contextBytes []byte
	
	// Use sql.NullString for fields that may be NULL
	var subjectType, subjectID, relation, permission, requestID, clientIP, userAgent sql.NullString

	err := s.graph.Pool.QueryRow(ctx, query, id).Scan(
		&log.ID, &log.Timestamp, &log.ActionType, &log.Result,
		&log.EntityType, &log.EntityID, &subjectType, &subjectID,
		&relation, &permission, &contextBytes, &requestID,
		&clientIP, &userAgent, &log.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	
	// Convert NullString fields to string (empty string if NULL)
	if subjectType.Valid {
		log.SubjectType = subjectType.String
	}
	if subjectID.Valid {
		log.SubjectID = subjectID.String
	}
	if relation.Valid {
		log.Relation = relation.String
	}
	if permission.Valid {
		log.Permission = permission.String
	}
	if requestID.Valid {
		log.RequestID = requestID.String
	}
	if clientIP.Valid {
		log.ClientIP = clientIP.String
	}
	if userAgent.Valid {
		log.UserAgent = userAgent.String
	}

	// Parse context JSON
	if len(contextBytes) > 0 {
		err = json.Unmarshal(contextBytes, &log.Context)
		if err != nil {
			log.Context = map[string]interface{}{
				"error":       "Failed to parse context data",
				"raw_context": string(contextBytes),
			}
		}
	} else {
		log.Context = map[string]interface{}{}
	}

	return &log, nil
}
