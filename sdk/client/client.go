package client

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"
)

// Config represents the configuration for the permission client
type Config struct {
	// BaseURL is the base URL of the permission service
	BaseURL string
	// HTTPClient is an optional custom HTTP client
	HTTPClient *http.Client
	// Timeout is the default request timeout
	Timeout time.Duration
}

// DefaultConfig returns the default configuration
func DefaultConfig() *Config {
	return &Config{
		BaseURL:    "http://localhost:4780",
		HTTPClient: http.DefaultClient,
		Timeout:    10 * time.Second,
	}
}

// Client is the permission service client
type Client struct {
	config *Config
	client *http.Client
}

// NewClient creates a new permission client with the given configuration
func NewClient(config *Config) *Client {
	if config == nil {
		config = DefaultConfig()
	}

	client := config.HTTPClient
	if client == nil {
		client = http.DefaultClient
	}

	return &Client{
		config: config,
		client: client,
	}
}

// CheckPermissionRequest represents a permission check request
type CheckPermissionRequest struct {
	SubjectType string                 `json:"subject_type"`
	SubjectID   string                 `json:"subject_id"`
	Permission  string                 `json:"permission"`
	ObjectType  string                 `json:"object_type"`
	ObjectID    string                 `json:"object_id"`
	Context     map[string]interface{} `json:"context,omitempty"`
}

// CheckPermissionResponse represents a permission check response
type CheckPermissionResponse struct {
	Allowed bool   `json:"allowed"`
	Error   string `json:"error,omitempty"`
}

// CheckPermission checks if a subject has permission on an object
func (c *Client) CheckPermission(ctx context.Context, req *CheckPermissionRequest) (*CheckPermissionResponse, error) {
	if req == nil {
		return nil, errors.New("request cannot be nil")
	}

	// Validate required fields
	if req.SubjectType == "" || req.SubjectID == "" || req.Permission == "" ||
		req.ObjectType == "" || req.ObjectID == "" {
		return nil, errors.New("subject_type, subject_id, permission, object_type, and object_id are required")
	}

	endpoint := fmt.Sprintf("%s/check", c.config.BaseURL)
	return c.doRequest(ctx, endpoint, req)
}

// CreateEntityRequest represents an entity creation request
type CreateEntityRequest struct {
	Type       string                 `json:"type"`
	ExternalID string                 `json:"external_id"`
	Properties map[string]interface{} `json:"properties"`
}

// EntityResponse represents an entity response
type EntityResponse struct {
	ID         int64                  `json:"id"`
	Type       string                 `json:"type"`
	ExternalID string                 `json:"external_id"`
	Properties map[string]interface{} `json:"properties"`
	CreatedAt  time.Time              `json:"created_at"`
	UpdatedAt  time.Time              `json:"updated_at"`
	Error      string                 `json:"error,omitempty"`
}

// CreateEntity creates a new entity
func (c *Client) CreateEntity(ctx context.Context, req *CreateEntityRequest) (*EntityResponse, error) {
	if req == nil {
		return nil, errors.New("request cannot be nil")
	}

	// Validate required fields
	if req.Type == "" || req.ExternalID == "" {
		return nil, errors.New("type and external_id are required")
	}

	endpoint := fmt.Sprintf("%s/entity", c.config.BaseURL)
	var resp EntityResponse
	err := c.post(ctx, endpoint, req, &resp)

	// Handle API errors
	if err != nil {
		var apiErr *APIError
		if errors.As(err, &apiErr) {
			// Handle specific error codes
			switch apiErr.Code {
			case "entity_already_exists":
				return nil, fmt.Errorf("entity with type '%s' and ID '%s' already exists: %w", req.Type, req.ExternalID, err)
			case "missing_fields":
				return nil, fmt.Errorf("missing required fields: %w", err)
			default:
				return nil, fmt.Errorf("failed to create entity: %w", err)
			}
		}
		return nil, fmt.Errorf("failed to create entity: %w", err)
	}

	if resp.Error != "" {
		return &resp, errors.New(resp.Error)
	}

	return &resp, nil
}

// GetEntity retrieves an entity by type and external ID
func (c *Client) GetEntity(ctx context.Context, entityType, externalID string) (*EntityResponse, error) {
	if entityType == "" || externalID == "" {
		return nil, errors.New("entity_type and external_id are required")
	}

	endpoint := fmt.Sprintf("%s/entity?type=%s&id=%s", c.config.BaseURL, entityType, externalID)
	var resp EntityResponse
	err := c.get(ctx, endpoint, &resp)
	if err != nil {
		return nil, err
	}

	if resp.Error != "" {
		return &resp, errors.New(resp.Error)
	}

	return &resp, nil
}

// CreateRelationRequest represents a relation creation request
type CreateRelationRequest struct {
	SubjectType string `json:"subject_type"`
	SubjectID   string `json:"subject_id"`
	Relation    string `json:"relation"`
	ObjectType  string `json:"object_type"`
	ObjectID    string `json:"object_id"`
}

// RelationResponse represents a relation response
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

// CreateRelation creates a new relation between entities
func (c *Client) CreateRelation(ctx context.Context, req *CreateRelationRequest) (*RelationResponse, error) {
	if req == nil {
		return nil, errors.New("request cannot be nil")
	}

	// Validate required fields
	if req.SubjectType == "" || req.SubjectID == "" || req.Relation == "" ||
		req.ObjectType == "" || req.ObjectID == "" {
		return nil, errors.New("subject_type, subject_id, relation, object_type, and object_id are required")
	}

	endpoint := fmt.Sprintf("%s/relation", c.config.BaseURL)
	var resp RelationResponse
	err := c.post(ctx, endpoint, req, &resp)
	if err != nil {
		return nil, err
	}

	if resp.Error != "" {
		return &resp, errors.New(resp.Error)
	}

	return &resp, nil
}

// TestRelationRequest represents a relation test request
type TestRelationRequest struct {
	SubjectType string `json:"subject_type"`
	SubjectID   string `json:"subject_id"`
	Relation    string `json:"relation"`
	ObjectType  string `json:"object_type"`
	ObjectID    string `json:"object_id"`
	Direction   string `json:"direction,omitempty"` // "normal", "reverse", or "both"
}

// TestRelationResponse represents a relation test response
type TestRelationResponse struct {
	NormalDirection  bool   `json:"normal_direction"`
	ReverseDirection bool   `json:"reverse_direction"`
	HasRelation      bool   `json:"has_relation"`
	Error            string `json:"error,omitempty"`
}

// TestRelation tests if a relation exists between entities
func (c *Client) TestRelation(ctx context.Context, req *TestRelationRequest) (*TestRelationResponse, error) {
	if req == nil {
		return nil, errors.New("request cannot be nil")
	}

	// Validate required fields
	if req.SubjectType == "" || req.SubjectID == "" || req.Relation == "" ||
		req.ObjectType == "" || req.ObjectID == "" {
		return nil, errors.New("subject_type, subject_id, relation, object_type, and object_id are required")
	}

	endpoint := fmt.Sprintf("%s/test-relation", c.config.BaseURL)
	var resp TestRelationResponse
	err := c.post(ctx, endpoint, req, &resp)
	if err != nil {
		return nil, err
	}

	if resp.Error != "" {
		return &resp, errors.New(resp.Error)
	}

	return &resp, nil
}

// CreatePermissionRequest represents a permission creation request
type CreatePermissionRequest struct {
	EntityType          string `json:"entity_type"`
	PermissionName      string `json:"permission_name"`
	ConditionExpression string `json:"condition_expression"`
	Description         string `json:"description,omitempty"`
}

// PermissionResponse represents a permission response
type PermissionResponse struct {
	ID                  int64     `json:"id"`
	EntityType          string    `json:"entity_type"`
	PermissionName      string    `json:"permission_name"`
	ConditionExpression string    `json:"condition_expression"`
	Description         string    `json:"description,omitempty"`
	CreatedAt           time.Time `json:"created_at"`
	Error               string    `json:"error,omitempty"`
}

// CreatePermission creates a new permission definition
func (c *Client) CreatePermission(ctx context.Context, req *CreatePermissionRequest) (*PermissionResponse, error) {
	if req == nil {
		return nil, errors.New("request cannot be nil")
	}

	// Validate required fields
	if req.EntityType == "" || req.PermissionName == "" || req.ConditionExpression == "" {
		return nil, errors.New("entity_type, permission_name, and condition_expression are required")
	}

	endpoint := fmt.Sprintf("%s/permission", c.config.BaseURL)
	var resp PermissionResponse
	err := c.post(ctx, endpoint, req, &resp)
	if err != nil {
		return nil, err
	}

	if resp.Error != "" {
		return &resp, errors.New(resp.Error)
	}

	return &resp, nil
}

// TestRuleRequest represents a rule test request
type TestRuleRequest struct {
	RuleName   string                 `json:"rule_name"`
	Parameters map[string]interface{} `json:"parameters"`
}

// TestRuleResponse represents a rule test response
type TestRuleResponse struct {
	Result bool   `json:"result"`
	Error  string `json:"error,omitempty"`
}

// TestRule tests a rule with parameters
func (c *Client) TestRule(ctx context.Context, req *TestRuleRequest) (*TestRuleResponse, error) {
	if req == nil {
		return nil, errors.New("request cannot be nil")
	}

	// Validate required fields
	if req.RuleName == "" {
		return nil, errors.New("rule_name is required")
	}

	endpoint := fmt.Sprintf("%s/api/test-rule", c.config.BaseURL)
	var resp TestRuleResponse
	err := c.post(ctx, endpoint, req, &resp)
	if err != nil {
		return nil, err
	}

	if resp.Error != "" {
		return &resp, errors.New(resp.Error)
	}

	return &resp, nil
}

// ListPermissionDefinitionsResponse represents a permission definitions list response
type ListPermissionDefinitionsResponse struct {
	Permissions []PermissionDefinition `json:"permissions"`
}

// PermissionDefinition represents a permission definition
type PermissionDefinition struct {
	ID                  int64  `json:"id"`
	EntityType          string `json:"entity_type"`
	PermissionName      string `json:"permission_name"`
	ConditionExpression string `json:"condition_expression"`
	Description         string `json:"description,omitempty"`
	CreatedAt           string `json:"created_at,omitempty"`
}

// ListPermissionDefinitions lists all permission definitions
func (c *Client) ListPermissionDefinitions(ctx context.Context) ([]PermissionDefinition, error) {
	endpoint := fmt.Sprintf("%s/api/permission-definitions", c.config.BaseURL)
	var resp []PermissionDefinition
	err := c.get(ctx, endpoint, &resp)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

// ListRuleDefinitionsResponse represents a rule definitions list response
type ListRuleDefinitionsResponse struct {
	Rules []RuleDefinition `json:"rules"`
}

// RuleDefinition represents a rule definition
type RuleDefinition struct {
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

// ListRuleDefinitions lists all rule definitions
func (c *Client) ListRuleDefinitions(ctx context.Context) ([]RuleDefinition, error) {
	endpoint := fmt.Sprintf("%s/api/rule-definitions", c.config.BaseURL)
	var resp []RuleDefinition
	err := c.get(ctx, endpoint, &resp)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

// DeleteEntityRequest represents an entity deletion request
type DeleteEntityRequest struct {
	Type       string `json:"type"`
	ExternalID string `json:"external_id"`
}

// DeleteEntity deletes an entity by type and external ID
func (c *Client) DeleteEntity(ctx context.Context, entityType, externalID string) error {
	if entityType == "" || externalID == "" {
		return errors.New("entity_type and external_id are required")
	}

	endpoint := fmt.Sprintf("%s/entity?type=%s&id=%s", c.config.BaseURL, entityType, externalID)
	return c.delete(ctx, endpoint)
}

// DeleteRelationRequest represents a relation deletion request
type DeleteRelationRequest struct {
	SubjectType string `json:"subject_type"`
	SubjectID   string `json:"subject_id"`
	Relation    string `json:"relation"`
	ObjectType  string `json:"object_type"`
	ObjectID    string `json:"object_id"`
}

// DeleteRelation deletes a relation between entities
func (c *Client) DeleteRelation(ctx context.Context, req *DeleteRelationRequest) error {
	if req == nil {
		return errors.New("request cannot be nil")
	}

	// Validate required fields
	if req.SubjectType == "" || req.SubjectID == "" || req.Relation == "" ||
		req.ObjectType == "" || req.ObjectID == "" {
		return errors.New("subject_type, subject_id, relation, object_type, and object_id are required")
	}

	// Construct query parameters
	endpoint := fmt.Sprintf("%s/relation?subject_type=%s&subject_id=%s&relation=%s&object_type=%s&object_id=%s",
		c.config.BaseURL, req.SubjectType, req.SubjectID, req.Relation, req.ObjectType, req.ObjectID)

	return c.delete(ctx, endpoint)
}

// DeletePermissionRequest represents a permission deletion request
type DeletePermissionRequest struct {
	EntityType     string `json:"entity_type"`
	PermissionName string `json:"permission_name"`
}

// DeletePermission deletes a permission definition
func (c *Client) DeletePermission(ctx context.Context, entityType, permissionName string) error {
	if entityType == "" || permissionName == "" {
		return errors.New("entity_type and permission_name are required")
	}

	endpoint := fmt.Sprintf("%s/permission?entity_type=%s&permission_name=%s", 
		c.config.BaseURL, entityType, permissionName)
	
	return c.delete(ctx, endpoint)
}

// doRequest performs a POST request to the specified endpoint with the given request and unmarshals the response into a CheckPermissionResponse
func (c *Client) doRequest(ctx context.Context, endpoint string, req interface{}) (*CheckPermissionResponse, error) {
	var resp CheckPermissionResponse
	err := c.post(ctx, endpoint, req, &resp)
	if err != nil {
		return nil, err
	}

	if resp.Error != "" {
		return &resp, errors.New(resp.Error)
	}

	return &resp, nil
}

// APIError defines a standardized error response from the API
type APIError struct {
	StatusCode int    `json:"-"`
	Code       string `json:"code,omitempty"`
	Message    string `json:"message"`
	Details    string `json:"details,omitempty"`
}

func (e *APIError) Error() string {
	if e.Code != "" {
		return fmt.Sprintf("[%s] %s (Status: %d)", e.Code, e.Message, e.StatusCode)
	}
	return fmt.Sprintf("%s (Status: %d)", e.Message, e.StatusCode)
}

// post performs a POST request to the specified endpoint with the given request and unmarshals the response into the specified response object
func (c *Client) post(ctx context.Context, endpoint string, req interface{}, resp interface{}) error {
	// Set up context with timeout
	if c.config.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, c.config.Timeout)
		defer cancel()
	}

	// Marshal request to JSON
	reqBody, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewBuffer(reqBody))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	// Send request
	httpResp, err := c.client.Do(httpReq)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer httpResp.Body.Close()

	// Check for non-success status code
	if httpResp.StatusCode < 200 || httpResp.StatusCode >= 300 {
		// Try to decode error response
		var apiErr APIError
		if err := json.NewDecoder(httpResp.Body).Decode(&apiErr); err != nil {
			// If we can't decode the error, create a generic one
			return &APIError{
				StatusCode: httpResp.StatusCode,
				Message:    fmt.Sprintf("request failed with status code %d", httpResp.StatusCode),
			}
		}

		apiErr.StatusCode = httpResp.StatusCode
		return &apiErr
	}

	// Decode response
	if err := json.NewDecoder(httpResp.Body).Decode(resp); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	return nil
}

// get performs a GET request to the specified endpoint and unmarshals the response into the specified response object
func (c *Client) get(ctx context.Context, endpoint string, resp interface{}) error {
	// Set up context with timeout
	if c.config.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, c.config.Timeout)
		defer cancel()
	}

	// Create HTTP request
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	httpReq.Header.Set("Accept", "application/json")

	// Send request
	httpResp, err := c.client.Do(httpReq)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer httpResp.Body.Close()

	// Check for non-success status code
	if httpResp.StatusCode < 200 || httpResp.StatusCode >= 300 {
		// Try to decode error response
		var apiErr APIError
		if err := json.NewDecoder(httpResp.Body).Decode(&apiErr); err != nil {
			// If we can't decode the error, create a generic one
			return &APIError{
				StatusCode: httpResp.StatusCode,
				Message:    fmt.Sprintf("request failed with status code %d", httpResp.StatusCode),
			}
		}

		apiErr.StatusCode = httpResp.StatusCode
		return &apiErr
	}

	// Decode response
	if err := json.NewDecoder(httpResp.Body).Decode(resp); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	return nil
}

// delete performs a DELETE request to the specified endpoint
func (c *Client) delete(ctx context.Context, endpoint string) error {
	// Set up context with timeout
	if c.config.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, c.config.Timeout)
		defer cancel()
	}

	// Create HTTP request
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodDelete, endpoint, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Send request
	httpResp, err := c.client.Do(httpReq)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer httpResp.Body.Close()

	// Check for non-success status code
	if httpResp.StatusCode < 200 || httpResp.StatusCode >= 300 {
		// Try to decode error response
		var apiErr APIError
		if err := json.NewDecoder(httpResp.Body).Decode(&apiErr); err != nil {
			// If we can't decode the error, create a generic one
			return &APIError{
				StatusCode: httpResp.StatusCode,
				Message:    fmt.Sprintf("request failed with status code %d", httpResp.StatusCode),
			}
		}

		apiErr.StatusCode = httpResp.StatusCode
		return &apiErr
	}

	return nil
}
