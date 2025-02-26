package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/dangerclosesec/supra/internal/auth/graph"
)

// AuthzService provides HTTP endpoints for authorization decisions
type AuthzService struct {
	graph *graph.IdentityGraph
	addr  string
}

// NewAuthzService creates a new authorization service
func NewAuthzService(connString, addr string) (*AuthzService, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Initializes the identity graph
	graph, err := graph.NewIdentityGraph(ctx, connString)
	if err != nil {
		return nil, fmt.Errorf("failed to create identity graph: %w", err)
	}

	return &AuthzService{
		graph: graph,
		addr:  addr,
	}, nil
}

// Add a new testing endpoint to visualize permission condition expressions
func (s *AuthzService) addPermissionVisualizer(mux *http.ServeMux) {
	mux.HandleFunc("/visualize-condition", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Parse the condition expression from the request
		var req struct {
			Condition string `json:"condition"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request format", http.StatusBadRequest)
			return
		}

		if req.Condition == "" {
			http.Error(w, "Condition is required", http.StatusBadRequest)
			return
		}

		// Parse the condition using our new parser
		parser := graph.NewConditionParser(req.Condition)
		expr, err := parser.Parse()
		if err != nil {
			jsonResponse(w, map[string]string{
				"error": fmt.Sprintf("Failed to parse condition: %v", err),
			}, http.StatusBadRequest)
			return
		}

		// Return the parsed expression as a string representation
		jsonResponse(w, map[string]string{
			"parsed_expression": expr.String(),
		}, http.StatusOK)
	})

	// Add an endpoint to test relations directly
	mux.HandleFunc("/test-relation", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req struct {
			SubjectType string `json:"subject_type"`
			SubjectID   string `json:"subject_id"`
			Relation    string `json:"relation"`
			ObjectType  string `json:"object_type"`
			ObjectID    string `json:"object_id"`
			Direction   string `json:"direction"` // "normal", "reverse", or "both"
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request format", http.StatusBadRequest)
			return
		}

		ctx := r.Context()
		var result struct {
			NormalDirection  bool `json:"normal_direction"`
			ReverseDirection bool `json:"reverse_direction"`
			HasRelation      bool `json:"has_relation"`
		}

		// Test normal direction (subject -> relation -> object)
		var exists bool
		if req.Direction == "normal" || req.Direction == "both" || req.Direction == "" {
			err := s.graph.Pool.QueryRow(ctx, `
				SELECT EXISTS (
					SELECT 1 FROM relations 
					WHERE subject_type = $1 AND subject_id = $2 
					AND relation = $3 
					AND object_type = $4 AND object_id = $5
				)
			`, req.SubjectType, req.SubjectID, req.Relation, req.ObjectType, req.ObjectID).Scan(&exists)

			if err != nil {
				jsonResponse(w, map[string]string{
					"error": fmt.Sprintf("Database error: %v", err),
				}, http.StatusInternalServerError)
				return
			}

			result.NormalDirection = exists
		}

		// Test reverse direction (object -> relation -> subject)
		if req.Direction == "reverse" || req.Direction == "both" || req.Direction == "" {
			err := s.graph.Pool.QueryRow(ctx, `
				SELECT EXISTS (
					SELECT 1 FROM relations 
					WHERE subject_type = $1 AND subject_id = $2 
					AND relation = $3 
					AND object_type = $4 AND object_id = $5
				)
			`, req.ObjectType, req.ObjectID, req.Relation, req.SubjectType, req.SubjectID).Scan(&exists)

			if err != nil {
				jsonResponse(w, map[string]string{
					"error": fmt.Sprintf("Database error: %v", err),
				}, http.StatusInternalServerError)
				return
			}

			result.ReverseDirection = exists
		}

		result.HasRelation = result.NormalDirection || result.ReverseDirection
		jsonResponse(w, result, http.StatusOK)
	})
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Allow requests from the Next.js frontend
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		// Handle preflight requests
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// Update the Start method to include the visualization endpoint
func (s *AuthzService) Start() error {
	// Define routes
	mux := http.NewServeMux()

	// Existing endpoints
	mux.HandleFunc("/check", s.checkPermissionHandler)
	mux.HandleFunc("/entity", s.entityHandler)
	mux.HandleFunc("/relation", s.relationHandler)
	mux.HandleFunc("/api/relation", s.relationHandler)
	mux.HandleFunc("/permission", s.permissionHandler)

	s.addSchemaExplorerEndpoints(mux)

	// Add testing endpoints
	s.addPermissionVisualizer(mux)

	// Add graph visualization endpoints
	s.addGraphVisualizationEndpoint(mux)

	// Add permission path
	s.addPermissionPathEndpoint(mux)

	s.addHealthCheckEndpoints(mux)

	// Wrap with logging middleware and CORS middleware
	handler := corsMiddleware(logMiddleware(mux))

	log.Printf("Starting authorization service on %s", s.addr)
	return http.ListenAndServe(s.addr, handler)
}

// CheckPermissionRequest represents an access check request
type CheckPermissionRequest struct {
	SubjectType string `json:"subject_type"`
	SubjectID   string `json:"subject_id"`
	Permission  string `json:"permission"`
	ObjectType  string `json:"object_type"`
	ObjectID    string `json:"object_id"`
}

// CheckPermissionResponse is the result of a permission check
type CheckPermissionResponse struct {
	Allowed bool   `json:"allowed"`
	Error   string `json:"error,omitempty"`
}

// Update the CheckPermission method to use the new parser
func (s *AuthzService) checkPermissionHandler(w http.ResponseWriter, r *http.Request) {
	// Only allows POST requests
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parses the JSON request
	var req CheckPermissionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonResponse(w, CheckPermissionResponse{
			Allowed: false,
			Error:   "Invalid request format",
		}, http.StatusBadRequest)
		return
	}

	// Validates the required fields
	if req.SubjectType == "" || req.SubjectID == "" || req.Permission == "" ||
		req.ObjectType == "" || req.ObjectID == "" {
		jsonResponse(w, CheckPermissionResponse{
			Allowed: false,
			Error:   "Missing required fields",
		}, http.StatusBadRequest)
		return
	}

	// Add debugging log
	log.Printf("Checking permission: %s has %s on %s:%s",
		fmt.Sprintf("%s:%s", req.SubjectType, req.SubjectID),
		req.Permission,
		req.ObjectType, req.ObjectID)

	// Performs the permission check
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	// Get the permission definition
	var conditionExpr string
	err := s.graph.Pool.QueryRow(ctx, `
		SELECT condition_expression
		FROM permission_definitions
		WHERE entity_type = $1 AND permission_name = $2
	`, req.ObjectType, req.Permission).Scan(&conditionExpr)

	if err != nil {
		log.Printf("Error retrieving permission definition: %v", err)
		jsonResponse(w, CheckPermissionResponse{
			Allowed: false,
			Error:   fmt.Sprintf("Permission definition not found: %s.%s", req.ObjectType, req.Permission),
		}, http.StatusNotFound)
		return
	}

	log.Printf("Permission condition: %s", conditionExpr)

	// Use the new condition parser and evaluator
	allowed, err := s.graph.EvaluateCondition(ctx, conditionExpr,
		req.SubjectType, req.SubjectID, req.ObjectType, req.ObjectID)

	if err != nil {
		log.Printf("Error evaluating permission: %v", err)
		jsonResponse(w, CheckPermissionResponse{
			Allowed: false,
			Error:   err.Error(),
		}, http.StatusInternalServerError)
		return
	}

	log.Printf("Permission check result: %v", allowed)

	// Returns the result
	jsonResponse(w, CheckPermissionResponse{
		Allowed: allowed,
	}, http.StatusOK)
}

// EntityRequest for creating entities
type EntityRequest struct {
	Type       string                 `json:"type"`
	ExternalID string                 `json:"external_id"`
	Properties map[string]interface{} `json:"properties"`
}

// EntityResponse after entity operations
type EntityResponse struct {
	ID         int64                  `json:"id"`
	Type       string                 `json:"type"`
	ExternalID string                 `json:"external_id"`
	Properties map[string]interface{} `json:"properties"`
	CreatedAt  time.Time              `json:"created_at"`
	UpdatedAt  time.Time              `json:"updated_at"`
	Error      string                 `json:"error,omitempty"`
}

// entityHandler manages entity creation and retrieval
func (s *AuthzService) entityHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		// Creates a new entity
		var req EntityRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			jsonResponse(w, EntityResponse{Error: "Invalid request format"}, http.StatusBadRequest)
			return
		}

		if req.Type == "" || req.ExternalID == "" {
			jsonResponse(w, EntityResponse{Error: "Type and external_id are required"}, http.StatusBadRequest)
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()

		entity, err := s.graph.CreateEntity(ctx, req.Type, req.ExternalID, req.Properties)
		if err != nil {
			jsonResponse(w, EntityResponse{Error: err.Error()}, http.StatusInternalServerError)
			return
		}

		jsonResponse(w, EntityResponse{
			ID:         entity.ID,
			Type:       entity.Type,
			ExternalID: entity.ExternalID,
			Properties: entity.Properties,
			CreatedAt:  entity.CreatedAt,
			UpdatedAt:  entity.UpdatedAt,
		}, http.StatusCreated)

	case http.MethodGet:
		// Retrieves an entity
		entityType := r.URL.Query().Get("type")
		externalID := r.URL.Query().Get("id")

		if entityType == "" || externalID == "" {
			jsonResponse(w, EntityResponse{Error: "Type and id query parameters are required"}, http.StatusBadRequest)
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()

		entity, err := s.graph.GetEntity(ctx, entityType, externalID)
		if err != nil {
			jsonResponse(w, EntityResponse{Error: err.Error()}, http.StatusNotFound)
			return
		}

		jsonResponse(w, EntityResponse{
			ID:         entity.ID,
			Type:       entity.Type,
			ExternalID: entity.ExternalID,
			Properties: entity.Properties,
			CreatedAt:  entity.CreatedAt,
			UpdatedAt:  entity.UpdatedAt,
		}, http.StatusOK)

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// RelationRequest for creating relations
type RelationRequest struct {
	SubjectType string `json:"subject_type"`
	SubjectID   string `json:"subject_id"`
	Relation    string `json:"relation"`
	ObjectType  string `json:"object_type"`
	ObjectID    string `json:"object_id"`
}

// RelationResponse after relation operations
type RelationResponse struct {
	ID          int64     `json:"id"`
	SubjectType string    `json:"subject_type"`
	SubjectID   string    `json:"subject_id"`
	Relation    string    `json:"relation"`
	ObjectType  string    `json:"object_type"`
	ObjectID    string    `json:"object_id"`
	CreatedAt   time.Time `json:"created_at"`
	Error       string    `json:"error,omitempty"`
}

// relationHandler manages relation creation
func (s *AuthzService) relationHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req RelationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonResponse(w, RelationResponse{Error: "Invalid request format"}, http.StatusBadRequest)
		return
	}

	if req.SubjectType == "" || req.SubjectID == "" || req.Relation == "" ||
		req.ObjectType == "" || req.ObjectID == "" {
		jsonResponse(w, RelationResponse{Error: "All fields are required"}, http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	// Check and create subject entity if missing
	subjectExists, _ := s.entityExists(ctx, req.SubjectType, req.SubjectID)
	if !subjectExists {
		// Create stub entity with minimal properties
		_, err := s.graph.CreateEntity(ctx, req.SubjectType, req.SubjectID, map[string]interface{}{
			"name":         req.SubjectID, // Use ID as default name
			"auto_created": true,          // Flag to indicate it was auto-created
		})
		if err != nil {
			log.Printf("Warning: Failed to auto-create subject entity: %v", err)
		}
	}

	// Same for object entity
	objectExists, _ := s.entityExists(ctx, req.ObjectType, req.ObjectID)
	if !objectExists {
		_, err := s.graph.CreateEntity(ctx, req.ObjectType, req.ObjectID, map[string]interface{}{
			"name":         req.ObjectID,
			"auto_created": true,
		})
		if err != nil {
			log.Printf("Warning: Failed to auto-create object entity: %v", err)
		}
	}

	relation, err := s.graph.CreateRelation(ctx, req.SubjectType, req.SubjectID,
		req.Relation, req.ObjectType, req.ObjectID)
	if err != nil {
		jsonResponse(w, map[string]string{"error": err.Error()}, http.StatusInternalServerError)
		return
	}

	jsonResponse(w, RelationResponse{
		ID:          relation.ID,
		SubjectType: relation.SubjectType,
		SubjectID:   relation.SubjectID,
		Relation:    relation.Relation,
		ObjectType:  relation.ObjectType,
		ObjectID:    relation.ObjectID,
		CreatedAt:   relation.CreatedAt,
	}, http.StatusCreated)
}

// PermissionRequest for creating permission definitions
type PermissionRequest struct {
	EntityType          string `json:"entity_type"`
	PermissionName      string `json:"permission_name"`
	ConditionExpression string `json:"condition_expression"`
	Description         string `json:"description,omitempty"`
}

// PermissionResponse after permission operations
type PermissionResponse struct {
	ID                  int64     `json:"id"`
	EntityType          string    `json:"entity_type"`
	PermissionName      string    `json:"permission_name"`
	ConditionExpression string    `json:"condition_expression"`
	Description         string    `json:"description,omitempty"`
	CreatedAt           time.Time `json:"created_at"`
	Error               string    `json:"error,omitempty"`
}

// permissionHandler manages permission definition creation
func (s *AuthzService) permissionHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req PermissionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonResponse(w, PermissionResponse{Error: "Invalid request format"}, http.StatusBadRequest)
		return
	}

	if req.EntityType == "" || req.PermissionName == "" || req.ConditionExpression == "" {
		jsonResponse(w, PermissionResponse{Error: "EntityType, PermissionName, and ConditionExpression are required"}, http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	perm, err := s.graph.AddPermissionDefinition(ctx, req.EntityType, req.PermissionName,
		req.ConditionExpression, req.Description)
	if err != nil {
		jsonResponse(w, map[string]string{"error": err.Error()}, http.StatusInternalServerError)
		return
	}

	jsonResponse(w, PermissionResponse{
		ID:                  perm.ID,
		EntityType:          perm.EntityType,
		PermissionName:      perm.PermissionName,
		ConditionExpression: perm.ConditionExpression,
		Description:         perm.Description,
		CreatedAt:           perm.CreatedAt,
	}, http.StatusCreated)
}

// jsonResponse sends a JSON response with the specified status code
func jsonResponse(w http.ResponseWriter, data interface{}, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Printf("Error encoding JSON response: %v", err)
	}
}

// logMiddleware logs HTTP requests
func logMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Records the response status by wrapping the ResponseWriter
		wrapper := &responseWriterWrapper{
			ResponseWriter: w,
			Status:         http.StatusOK,
		}

		next.ServeHTTP(wrapper, r)

		log.Printf("[%s] %s %s %d %s", r.Method, r.URL.Path, r.RemoteAddr, wrapper.Status, time.Since(start))
	})
}

// responseWriterWrapper captures the status code of the response
type responseWriterWrapper struct {
	http.ResponseWriter
	Status int
}

// WriteHeader captures the status code before writing it
func (w *responseWriterWrapper) WriteHeader(statusCode int) {
	w.Status = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}

func main() {
	// Reads configuration from environment variables
	connString := os.Getenv("DB_URL")
	if connString == "" {
		connString = "postgres://postgres:password@localhost:5432/identity_graph"
	}

	addr := os.Getenv("LISTEN_ADDR")
	if addr == "" {
		addr = ":4780"
	}

	// Creates and starts the service
	service, err := NewAuthzService(connString, addr)
	if err != nil {
		log.Fatalf("Failed to create authorization service: %v", err)
	}

	// Registers signal handlers for graceful shutdown (omitted for brevity)

	// Starts the HTTP server
	log.Fatal(service.Start())
}

// Helper function to check if entity exists
func (s *AuthzService) entityExists(ctx context.Context, entityType, externalID string) (bool, error) {
	var exists bool
	err := s.graph.Pool.QueryRow(ctx, `
        SELECT EXISTS (
            SELECT 1 FROM entities 
            WHERE type = $1 AND external_id = $2
        )
    `, entityType, externalID).Scan(&exists)
	return exists, err
}
