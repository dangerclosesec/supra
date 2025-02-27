package client

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestNewClient(t *testing.T) {
	// Test with nil config
	client := NewClient(nil)
	if client.config.BaseURL != "http://localhost:4780" {
		t.Errorf("Expected default BaseURL, got %s", client.config.BaseURL)
	}
	if client.client != http.DefaultClient {
		t.Error("Expected default HTTP client")
	}

	// Test with custom config
	customConfig := &Config{
		BaseURL:    "http://example.com",
		Timeout:    5 * time.Second,
		HTTPClient: &http.Client{Timeout: 10 * time.Second},
	}
	client = NewClient(customConfig)
	if client.config.BaseURL != "http://example.com" {
		t.Errorf("Expected custom BaseURL, got %s", client.config.BaseURL)
	}
	if client.config.Timeout != 5*time.Second {
		t.Errorf("Expected custom timeout, got %v", client.config.Timeout)
	}
	if client.client != customConfig.HTTPClient {
		t.Error("Expected custom HTTP client")
	}
}

func TestCheckPermission(t *testing.T) {
	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check request method and path
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST request, got %s", r.Method)
		}
		if r.URL.Path != "/check" {
			t.Errorf("Expected /check path, got %s", r.URL.Path)
		}

		// Decode request
		var req CheckPermissionRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Errorf("Failed to decode request: %v", err)
			http.Error(w, "Bad request", http.StatusBadRequest)
			return
		}

		// Check required fields
		if req.SubjectType == "" || req.SubjectID == "" || req.Permission == "" ||
			req.ObjectType == "" || req.ObjectID == "" {
			http.Error(w, "Missing required fields", http.StatusBadRequest)
			return
		}

		// Return response based on input
		allowed := req.SubjectType == "user" && req.SubjectID == "123" &&
			req.Permission == "read" && req.ObjectType == "document" && req.ObjectID == "456"

		resp := CheckPermissionResponse{
			Allowed: allowed,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	// Create client
	client := NewClient(&Config{
		BaseURL: server.URL,
	})

	// Test valid request (allowed)
	req := &CheckPermissionRequest{
		SubjectType: "user",
		SubjectID:   "123",
		Permission:  "read",
		ObjectType:  "document",
		ObjectID:    "456",
		Context:     map[string]interface{}{"key": "value"},
	}
	resp, err := client.CheckPermission(context.Background(), req)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if !resp.Allowed {
		t.Error("Expected permission to be allowed")
	}

	// Test valid request (denied)
	req = &CheckPermissionRequest{
		SubjectType: "user",
		SubjectID:   "789",
		Permission:  "read",
		ObjectType:  "document",
		ObjectID:    "456",
	}
	resp, err = client.CheckPermission(context.Background(), req)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if resp.Allowed {
		t.Error("Expected permission to be denied")
	}

	// Test nil request
	if _, err := client.CheckPermission(context.Background(), nil); err == nil {
		t.Error("Expected error for nil request")
	}

	// Test missing required fields
	req = &CheckPermissionRequest{
		SubjectType: "user",
		SubjectID:   "123",
		// Missing Permission
		ObjectType: "document",
		ObjectID:   "456",
	}
	if _, err := client.CheckPermission(context.Background(), req); err == nil {
		t.Error("Expected error for missing required fields")
	}
	
	// Test server error
	errorServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}))
	defer errorServer.Close()
	
	errorClient := NewClient(&Config{
		BaseURL: errorServer.URL,
	})
	
	validReq := &CheckPermissionRequest{
		SubjectType: "user",
		SubjectID:   "123",
		Permission:  "read",
		ObjectType:  "document",
		ObjectID:    "456",
	}
	_, err = errorClient.CheckPermission(context.Background(), validReq)
	if err == nil {
		t.Error("Expected error for server error")
	}
}

func TestCreateEntity(t *testing.T) {
	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check request method and path
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST request, got %s", r.Method)
		}
		if r.URL.Path != "/entity" {
			t.Errorf("Expected /entity path, got %s", r.URL.Path)
		}

		// Decode request
		var req CreateEntityRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Errorf("Failed to decode request: %v", err)
			http.Error(w, "Bad request", http.StatusBadRequest)
			return
		}

		// Check required fields
		if req.Type == "" || req.ExternalID == "" {
			http.Error(w, "Missing required fields", http.StatusBadRequest)
			return
		}

		// Return success response
		resp := EntityResponse{
			ID:         1,
			Type:       req.Type,
			ExternalID: req.ExternalID,
			Properties: req.Properties,
			CreatedAt:  time.Now(),
			UpdatedAt:  time.Now(),
		}
		w.WriteHeader(http.StatusCreated)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	// Create client
	client := NewClient(&Config{
		BaseURL: server.URL,
	})

	// Test valid request
	req := &CreateEntityRequest{
		Type:       "user",
		ExternalID: "123",
		Properties: map[string]interface{}{
			"name":  "Test User",
			"email": "test@example.com",
		},
	}
	resp, err := client.CreateEntity(context.Background(), req)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if resp.Type != "user" || resp.ExternalID != "123" {
		t.Errorf("Expected user:123, got %s:%s", resp.Type, resp.ExternalID)
	}

	// Test nil request
	if _, err := client.CreateEntity(context.Background(), nil); err == nil {
		t.Error("Expected error for nil request")
	}

	// Test missing required fields
	req = &CreateEntityRequest{
		Type: "user",
		// Missing ExternalID
		Properties: map[string]interface{}{
			"name": "Test User",
		},
	}
	if _, err := client.CreateEntity(context.Background(), req); err == nil {
		t.Error("Expected error for missing required fields")
	}
}

func TestGetEntity(t *testing.T) {
	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check request method and path
		if r.Method != http.MethodGet {
			t.Errorf("Expected GET request, got %s", r.Method)
		}
		if r.URL.Path != "/entity" {
			t.Errorf("Expected /entity path, got %s", r.URL.Path)
		}

		// Check query parameters
		entityType := r.URL.Query().Get("type")
		externalID := r.URL.Query().Get("id")
		if entityType == "" || externalID == "" {
			http.Error(w, "Missing required query parameters", http.StatusBadRequest)
			return
		}

		// Return success response
		resp := EntityResponse{
			ID:         1,
			Type:       entityType,
			ExternalID: externalID,
			Properties: map[string]interface{}{
				"name":  "Test User",
				"email": "test@example.com",
			},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	// Create client
	client := NewClient(&Config{
		BaseURL: server.URL,
	})

	// Test valid request
	resp, err := client.GetEntity(context.Background(), "user", "123")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if resp.Type != "user" || resp.ExternalID != "123" {
		t.Errorf("Expected user:123, got %s:%s", resp.Type, resp.ExternalID)
	}

	// Test missing required fields
	if _, err := client.GetEntity(context.Background(), "", "123"); err == nil {
		t.Error("Expected error for missing type")
	}
	if _, err := client.GetEntity(context.Background(), "user", ""); err == nil {
		t.Error("Expected error for missing external ID")
	}
	
	// Test server error
	errorServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}))
	defer errorServer.Close()
	
	errorClient := NewClient(&Config{
		BaseURL: errorServer.URL,
	})
	
	_, err = errorClient.GetEntity(context.Background(), "user", "123")
	if err == nil {
		t.Error("Expected error for server error")
	}
}

func TestCreateRelation(t *testing.T) {
	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check request method and path
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST request, got %s", r.Method)
		}
		if r.URL.Path != "/relation" {
			t.Errorf("Expected /relation path, got %s", r.URL.Path)
		}

		// Decode request
		var req CreateRelationRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Errorf("Failed to decode request: %v", err)
			http.Error(w, "Bad request", http.StatusBadRequest)
			return
		}

		// Check required fields
		if req.SubjectType == "" || req.SubjectID == "" || req.Relation == "" ||
			req.ObjectType == "" || req.ObjectID == "" {
			http.Error(w, "Missing required fields", http.StatusBadRequest)
			return
		}

		// Return success response
		resp := RelationResponse{
			ID:          1,
			SubjectType: req.SubjectType,
			SubjectID:   req.SubjectID,
			Relation:    req.Relation,
			ObjectType:  req.ObjectType,
			ObjectID:    req.ObjectID,
			CreatedAt:   time.Now(),
		}
		w.WriteHeader(http.StatusCreated)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	// Create client
	client := NewClient(&Config{
		BaseURL: server.URL,
	})

	// Test valid request
	req := &CreateRelationRequest{
		SubjectType: "user",
		SubjectID:   "123",
		Relation:    "owner",
		ObjectType:  "document",
		ObjectID:    "456",
	}
	resp, err := client.CreateRelation(context.Background(), req)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if resp.SubjectType != "user" || resp.SubjectID != "123" ||
		resp.Relation != "owner" || resp.ObjectType != "document" || resp.ObjectID != "456" {
		t.Errorf("Unexpected response: %+v", resp)
	}

	// Test nil request
	if _, err := client.CreateRelation(context.Background(), nil); err == nil {
		t.Error("Expected error for nil request")
	}

	// Test missing required fields
	req = &CreateRelationRequest{
		SubjectType: "user",
		SubjectID:   "123",
		// Missing Relation
		ObjectType: "document",
		ObjectID:   "456",
	}
	if _, err := client.CreateRelation(context.Background(), req); err == nil {
		t.Error("Expected error for missing required fields")
	}
}

func TestTestRelation(t *testing.T) {
	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check request method and path
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST request, got %s", r.Method)
		}
		if r.URL.Path != "/test-relation" {
			t.Errorf("Expected /test-relation path, got %s", r.URL.Path)
		}

		// Decode request
		var req TestRelationRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Errorf("Failed to decode request: %v", err)
			http.Error(w, "Bad request", http.StatusBadRequest)
			return
		}

		// Check required fields
		if req.SubjectType == "" || req.SubjectID == "" || req.Relation == "" ||
			req.ObjectType == "" || req.ObjectID == "" {
			http.Error(w, "Missing required fields", http.StatusBadRequest)
			return
		}

		// Return response based on input
		hasRelation := (req.SubjectType == "user" && req.SubjectID == "123" &&
			req.Relation == "owner" && req.ObjectType == "document" && req.ObjectID == "456")

		resp := TestRelationResponse{
			NormalDirection:  hasRelation,
			ReverseDirection: false,
			HasRelation:      hasRelation,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	// Create client
	client := NewClient(&Config{
		BaseURL: server.URL,
	})

	// Test valid request (relation exists)
	req := &TestRelationRequest{
		SubjectType: "user",
		SubjectID:   "123",
		Relation:    "owner",
		ObjectType:  "document",
		ObjectID:    "456",
		Direction:   "both",
	}
	resp, err := client.TestRelation(context.Background(), req)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if !resp.HasRelation {
		t.Error("Expected relation to exist")
	}

	// Test valid request (relation does not exist)
	req = &TestRelationRequest{
		SubjectType: "user",
		SubjectID:   "789",
		Relation:    "owner",
		ObjectType:  "document",
		ObjectID:    "456",
	}
	resp, err = client.TestRelation(context.Background(), req)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if resp.HasRelation {
		t.Error("Expected relation not to exist")
	}

	// Test nil request
	if _, err := client.TestRelation(context.Background(), nil); err == nil {
		t.Error("Expected error for nil request")
	}

	// Test missing required fields
	req = &TestRelationRequest{
		SubjectType: "user",
		SubjectID:   "123",
		// Missing Relation
		ObjectType: "document",
		ObjectID:   "456",
	}
	if _, err := client.TestRelation(context.Background(), req); err == nil {
		t.Error("Expected error for missing required fields")
	}
}

func TestCreatePermission(t *testing.T) {
	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check request method and path
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST request, got %s", r.Method)
		}
		if r.URL.Path != "/permission" {
			t.Errorf("Expected /permission path, got %s", r.URL.Path)
		}

		// Decode request
		var req CreatePermissionRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Errorf("Failed to decode request: %v", err)
			http.Error(w, "Bad request", http.StatusBadRequest)
			return
		}

		// Check required fields
		if req.EntityType == "" || req.PermissionName == "" || req.ConditionExpression == "" {
			http.Error(w, "Missing required fields", http.StatusBadRequest)
			return
		}

		// Return success response
		resp := PermissionResponse{
			ID:                  1,
			EntityType:          req.EntityType,
			PermissionName:      req.PermissionName,
			ConditionExpression: req.ConditionExpression,
			Description:         req.Description,
			CreatedAt:           time.Now(),
		}
		w.WriteHeader(http.StatusCreated)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	// Create client
	client := NewClient(&Config{
		BaseURL: server.URL,
	})

	// Test valid request
	req := &CreatePermissionRequest{
		EntityType:          "document",
		PermissionName:      "read",
		ConditionExpression: "owner",
		Description:         "Can read the document",
	}
	resp, err := client.CreatePermission(context.Background(), req)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if resp.EntityType != "document" || resp.PermissionName != "read" ||
		resp.ConditionExpression != "owner" || resp.Description != "Can read the document" {
		t.Errorf("Unexpected response: %+v", resp)
	}

	// Test nil request
	if _, err := client.CreatePermission(context.Background(), nil); err == nil {
		t.Error("Expected error for nil request")
	}

	// Test missing required fields
	req = &CreatePermissionRequest{
		EntityType:     "document",
		PermissionName: "read",
		// Missing ConditionExpression
	}
	if _, err := client.CreatePermission(context.Background(), req); err == nil {
		t.Error("Expected error for missing required fields")
	}
}

func TestTestRule(t *testing.T) {
	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check request method and path
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST request, got %s", r.Method)
		}
		if r.URL.Path != "/api/test-rule" {
			t.Errorf("Expected /api/test-rule path, got %s", r.URL.Path)
		}

		// Decode request
		var req TestRuleRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Errorf("Failed to decode request: %v", err)
			http.Error(w, "Bad request", http.StatusBadRequest)
			return
		}

		// Check required fields
		if req.RuleName == "" {
			http.Error(w, "Missing required fields", http.StatusBadRequest)
			return
		}

		// Return response based on input
		result := false
		if req.RuleName == "isAdmin" {
			if userID, ok := req.Parameters["user_id"]; ok && userID == "admin" {
				result = true
			}
		}

		resp := TestRuleResponse{
			Result: result,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	// Create client
	client := NewClient(&Config{
		BaseURL: server.URL,
	})

	// Test valid request (rule passes)
	req := &TestRuleRequest{
		RuleName: "isAdmin",
		Parameters: map[string]interface{}{
			"user_id": "admin",
		},
	}
	resp, err := client.TestRule(context.Background(), req)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if !resp.Result {
		t.Error("Expected rule to pass")
	}

	// Test valid request (rule fails)
	req = &TestRuleRequest{
		RuleName: "isAdmin",
		Parameters: map[string]interface{}{
			"user_id": "user",
		},
	}
	resp, err = client.TestRule(context.Background(), req)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if resp.Result {
		t.Error("Expected rule to fail")
	}

	// Test nil request
	if _, err := client.TestRule(context.Background(), nil); err == nil {
		t.Error("Expected error for nil request")
	}

	// Test missing required fields
	req = &TestRuleRequest{
		// Missing RuleName
		Parameters: map[string]interface{}{
			"user_id": "admin",
		},
	}
	if _, err := client.TestRule(context.Background(), req); err == nil {
		t.Error("Expected error for missing required fields")
	}
}

func TestListPermissionDefinitions(t *testing.T) {
	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check request method and path
		if r.Method != http.MethodGet {
			t.Errorf("Expected GET request, got %s", r.Method)
		}
		if r.URL.Path != "/api/permission-definitions" {
			t.Errorf("Expected /api/permission-definitions path, got %s", r.URL.Path)
		}

		// Return success response
		permissions := []PermissionDefinition{
			{
				ID:                  1,
				EntityType:          "document",
				PermissionName:      "read",
				ConditionExpression: "owner",
				Description:         "Can read the document",
				CreatedAt:           time.Now().Format(time.RFC3339),
			},
			{
				ID:                  2,
				EntityType:          "document",
				PermissionName:      "write",
				ConditionExpression: "owner",
				Description:         "Can write to the document",
				CreatedAt:           time.Now().Format(time.RFC3339),
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(permissions)
	}))
	defer server.Close()

	// Create client
	client := NewClient(&Config{
		BaseURL: server.URL,
	})

	// Test valid request
	permissions, err := client.ListPermissionDefinitions(context.Background())
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if len(permissions) != 2 {
		t.Errorf("Expected 2 permissions, got %d", len(permissions))
	}
	if permissions[0].EntityType != "document" || permissions[0].PermissionName != "read" {
		t.Errorf("Unexpected permission: %+v", permissions[0])
	}
	if permissions[1].EntityType != "document" || permissions[1].PermissionName != "write" {
		t.Errorf("Unexpected permission: %+v", permissions[1])
	}
}

func TestErrorHandling(t *testing.T) {
	// Test server error with invalid JSON response
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte("{invalid json"))
	}))
	defer server.Close()
	
	client := NewClient(&Config{
		BaseURL: server.URL,
	})
	
	_, err := client.ListPermissionDefinitions(context.Background())
	if err == nil {
		t.Error("Expected error for invalid JSON response")
	}
	
	// Test with response containing error field
	serverWithErrorField := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.URL.Path == "/check" {
			resp := CheckPermissionResponse{
				Allowed: false,
				Error:   "Permission check failed",
			}
			json.NewEncoder(w).Encode(resp)
		}
	}))
	defer serverWithErrorField.Close()
	
	clientWithError := NewClient(&Config{
		BaseURL: serverWithErrorField.URL,
	})
	
	req := &CheckPermissionRequest{
		SubjectType: "user",
		SubjectID:   "123",
		Permission:  "read",
		ObjectType:  "document",
		ObjectID:    "456",
	}
	_, err = clientWithError.CheckPermission(context.Background(), req)
	if err == nil {
		t.Error("Expected error when response contains error field")
	}
}

func TestListRuleDefinitions(t *testing.T) {
	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check request method and path
		if r.Method != http.MethodGet {
			t.Errorf("Expected GET request, got %s", r.Method)
		}
		if r.URL.Path != "/api/rule-definitions" {
			t.Errorf("Expected /api/rule-definitions path, got %s", r.URL.Path)
		}

		// Return success response
		rules := []RuleDefinition{
			{
				ID:   1,
				Name: "isAdmin",
				Parameters: []ParameterDefinition{
					{
						Name:     "user_id",
						DataType: "string",
					},
				},
				Expression:  "user_id == \"admin\"",
				Description: "Checks if the user is an admin",
				CreatedAt:   time.Now().Format(time.RFC3339),
			},
			{
				ID:   2,
				Name: "hasRole",
				Parameters: []ParameterDefinition{
					{
						Name:     "user_id",
						DataType: "string",
					},
					{
						Name:     "role",
						DataType: "string",
					},
				},
				Expression:  "user_id has role",
				Description: "Checks if the user has the specified role",
				CreatedAt:   time.Now().Format(time.RFC3339),
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(rules)
	}))
	defer server.Close()

	// Create client
	client := NewClient(&Config{
		BaseURL: server.URL,
	})

	// Test valid request
	rules, err := client.ListRuleDefinitions(context.Background())
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if len(rules) != 2 {
		t.Errorf("Expected 2 rules, got %d", len(rules))
	}
	if rules[0].Name != "isAdmin" || len(rules[0].Parameters) != 1 {
		t.Errorf("Unexpected rule: %+v", rules[0])
	}
	if rules[1].Name != "hasRole" || len(rules[1].Parameters) != 2 {
		t.Errorf("Unexpected rule: %+v", rules[1])
	}
}