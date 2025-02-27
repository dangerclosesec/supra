package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/dangerclosesec/supra/internal/repository"
	"github.com/dangerclosesec/supra/internal/service"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

// AuthzAuditLogHandler handles API requests related to authorization audit logs
type AuthzAuditLogHandler struct {
	auditLogService *service.AuthzAuditLogService
}

// NewAuthzAuditLogHandler creates a new audit log handler
func NewAuthzAuditLogHandler(auditLogService *service.AuthzAuditLogService) *AuthzAuditLogHandler {
	return &AuthzAuditLogHandler{
		auditLogService: auditLogService,
	}
}

// GetAuditLogs handles requests to retrieve audit logs with filtering
func (h *AuthzAuditLogHandler) GetAuditLogs(w http.ResponseWriter, r *http.Request) {
	params := repository.QueryParams{}

	// Apply filters from query parameters
	if actionType := r.URL.Query().Get("action_type"); actionType != "" {
		params.ActionType = actionType
	}

	if entityType := r.URL.Query().Get("entity_type"); entityType != "" {
		params.EntityType = entityType
	}

	if entityID := r.URL.Query().Get("entity_id"); entityID != "" {
		params.EntityID = entityID
	}

	if subjectType := r.URL.Query().Get("subject_type"); subjectType != "" {
		params.SubjectType = subjectType
	}

	if subjectID := r.URL.Query().Get("subject_id"); subjectID != "" {
		params.SubjectID = subjectID
	}

	if resultStr := r.URL.Query().Get("result"); resultStr != "" {
		result, err := strconv.ParseBool(resultStr)
		if err == nil {
			params.Result = &result
		}
	}

	if startTimeStr := r.URL.Query().Get("start_time"); startTimeStr != "" {
		startTime, err := time.Parse(time.RFC3339, startTimeStr)
		if err == nil {
			params.StartTime = startTime
		}
	}

	if endTimeStr := r.URL.Query().Get("end_time"); endTimeStr != "" {
		endTime, err := time.Parse(time.RFC3339, endTimeStr)
		if err == nil {
			params.EndTime = endTime
		}
	}

	// Pagination
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		limit, err := strconv.Atoi(limitStr)
		if err == nil && limit > 0 {
			params.Limit = limit
		}
	}

	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		offset, err := strconv.Atoi(offsetStr)
		if err == nil && offset >= 0 {
			params.Offset = offset
		}
	}

	// Query logs
	logs, total, err := h.auditLogService.GetAuditLogs(r.Context(), params)
	if err != nil {
		http.Error(w, "Failed to retrieve audit logs: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Prepare response
	response := struct {
		Logs  interface{} `json:"logs"`
		Total int64       `json:"total"`
	}{
		Logs:  logs,
		Total: total,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Failed to encode response: "+err.Error(), http.StatusInternalServerError)
		return
	}
}

// GetAuditLogByID handles requests to retrieve a specific audit log by ID
func (h *AuthzAuditLogHandler) GetAuditLogByID(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	if idStr == "" {
		http.Error(w, "Missing audit log ID", http.StatusBadRequest)
		return
	}

	id, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "Invalid audit log ID format", http.StatusBadRequest)
		return
	}

	log, err := h.auditLogService.GetAuditLogByID(r.Context(), id)
	if err != nil {
		http.Error(w, "Failed to retrieve audit log: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(log); err != nil {
		http.Error(w, "Failed to encode response: "+err.Error(), http.StatusInternalServerError)
		return
	}
}