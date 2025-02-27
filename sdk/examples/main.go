package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/dangerclosesec/supra/sdk/client"
)

const (
	// Change these values to match your environment
	serviceURL = "http://localhost:4780"
)

func main() {
	// Initialize the client
	config := &client.Config{
		BaseURL: serviceURL,
		Timeout: 10 * time.Second,
	}
	c := client.NewClient(config)

	// Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// Run the example
	if err := runExample(ctx, c); err != nil {
		log.Fatalf("Error running example: %v", err)
	}
}

func runExample(ctx context.Context, c *client.Client) error {
	fmt.Println("Running permission SDK example...")

	// Step 1: Create document permission definition
	fmt.Println("\n1. Creating permission definition...")
	permReq := &client.CreatePermissionRequest{
		EntityType:          "document",
		PermissionName:      "read",
		ConditionExpression: "owner OR viewer",
		Description:         "Can read the document when user is an owner or viewer",
	}
	perm, err := c.CreatePermission(ctx, permReq)
	if err != nil {
		return fmt.Errorf("failed to create permission: %w", err)
	}
	fmt.Printf("Permission created: %s.%s\n", perm.EntityType, perm.PermissionName)

	// Step 2: Create entities
	fmt.Println("\n2. Creating entities...")
	
	// Create a user
	userReq := &client.CreateEntityRequest{
		Type:       "user",
		ExternalID: "alice",
		Properties: map[string]interface{}{
			"name":  "Alice",
			"email": "alice@example.com",
		},
	}
	user, err := c.CreateEntity(ctx, userReq)
	if err != nil {
		// Check if this is an APIError
		var apiErr *client.APIError
		if errors.As(err, &apiErr) {
			// If entity already exists, we can continue
			if apiErr.Code == "entity_already_exists" {
				fmt.Printf("User already exists: %s:%s (continuing with existing user)\n", userReq.Type, userReq.ExternalID)
				
				// Fetch the existing user
				user, fetchErr := c.GetEntity(ctx, userReq.Type, userReq.ExternalID)
				if fetchErr != nil {
					return fmt.Errorf("failed to fetch existing user: %w", fetchErr)
				}
			} else {
				// For other API errors, return detailed information
				return fmt.Errorf("API error creating user: [%s] %s - %s", 
					apiErr.Code, apiErr.Message, apiErr.Details)
			}
		} else {
			// For non-API errors
			return fmt.Errorf("failed to create user: %w", err)
		}
	} else {
		fmt.Printf("User created: %s:%s\n", user.Type, user.ExternalID)
	}

	// Create a document
	docReq := &client.CreateEntityRequest{
		Type:       "document",
		ExternalID: "doc1",
		Properties: map[string]interface{}{
			"title":     "Sample Document",
			"content":   "This is a sample document for testing permissions.",
			"createdAt": time.Now().Format(time.RFC3339),
		},
	}
	doc, err := c.CreateEntity(ctx, docReq)
	if err != nil {
		// Check if this is an APIError
		var apiErr *client.APIError
		if errors.As(err, &apiErr) {
			// If entity already exists, we can continue
			if apiErr.Code == "entity_already_exists" {
				fmt.Printf("Document already exists: %s:%s (continuing with existing document)\n", docReq.Type, docReq.ExternalID)
				
				// Fetch the existing document
				doc, fetchErr := c.GetEntity(ctx, docReq.Type, docReq.ExternalID)
				if fetchErr != nil {
					return fmt.Errorf("failed to fetch existing document: %w", fetchErr)
				}
			} else {
				// For other API errors, return detailed information
				return fmt.Errorf("API error creating document: [%s] %s - %s", 
					apiErr.Code, apiErr.Message, apiErr.Details)
			}
		} else {
			// For non-API errors
			return fmt.Errorf("failed to create document: %w", err)
		}
	} else {
		fmt.Printf("Document created: %s:%s\n", doc.Type, doc.ExternalID)
	}

	// Step 3: Create relations
	fmt.Println("\n3. Creating relations...")
	
	// Make user the owner of the document
	relReq := &client.CreateRelationRequest{
		SubjectType: "user",
		SubjectID:   "alice",
		Relation:    "owner",
		ObjectType:  "document",
		ObjectID:    "doc1",
	}
	rel, err := c.CreateRelation(ctx, relReq)
	if err != nil {
		return fmt.Errorf("failed to create relation: %w", err)
	}
	fmt.Printf("Relation created: %s:%s --%s--> %s:%s\n",
		rel.SubjectType, rel.SubjectID, rel.Relation, rel.ObjectType, rel.ObjectID)

	// Step 4: Test the relation exists
	fmt.Println("\n4. Testing relation...")
	testRelReq := &client.TestRelationRequest{
		SubjectType: "user",
		SubjectID:   "alice",
		Relation:    "owner",
		ObjectType:  "document",
		ObjectID:    "doc1",
		Direction:   "both", // Check both directions
	}
	testRel, err := c.TestRelation(ctx, testRelReq)
	if err != nil {
		return fmt.Errorf("failed to test relation: %w", err)
	}
	fmt.Printf("Relation test result: hasRelation=%v\n", testRel.HasRelation)

	// Step 5: Check permissions
	fmt.Println("\n5. Checking permissions...")
	
	// Check if Alice can read the document
	checkReq := &client.CheckPermissionRequest{
		SubjectType: "user",
		SubjectID:   "alice",
		Permission:  "read",
		ObjectType:  "document",
		ObjectID:    "doc1",
		Context:     map[string]interface{}{}, // No additional context needed
	}
	checkResp, err := c.CheckPermission(ctx, checkReq)
	if err != nil {
		return fmt.Errorf("failed to check permission: %w", err)
	}
	fmt.Printf("Permission check result: allowed=%v\n", checkResp.Allowed)

	// Step 6: Create and test a rule
	fmt.Println("\n6. Creating and testing a rule...")
	// First check if the rule exists
	ruleDefs, err := c.ListRuleDefinitions(ctx)
	if err != nil {
		return fmt.Errorf("failed to list rule definitions: %w", err)
	}
	
	ruleExists := false
	for _, rule := range ruleDefs {
		if rule.Name == "isDocumentOwner" {
			ruleExists = true
			break
		}
	}
	
	if !ruleExists {
		// Note: In a real application, you'd need a way to create rules or sync them from a schema file
		fmt.Println("Rule 'isDocumentOwner' doesn't exist. In a real application, you'd create it here.")
		fmt.Println("For the purpose of this example, we'll assume it exists and skip testing it.")
	} else {
		// Test the rule
		testRuleReq := &client.TestRuleRequest{
			RuleName: "isDocumentOwner",
			Parameters: map[string]interface{}{
				"user_id":    "alice",
				"document_id": "doc1",
			},
		}
		testRuleResp, err := c.TestRule(ctx, testRuleReq)
		if err != nil {
			return fmt.Errorf("failed to test rule: %w", err)
		}
		fmt.Printf("Rule test result: result=%v\n", testRuleResp.Result)
	}

	// Step 7: List all permissions
	fmt.Println("\n7. Listing all permission definitions...")
	permissions, err := c.ListPermissionDefinitions(ctx)
	if err != nil {
		return fmt.Errorf("failed to list permission definitions: %w", err)
	}
	
	fmt.Printf("Found %d permission definitions:\n", len(permissions))
	for _, p := range permissions {
		fmt.Printf("- %s.%s: %s\n", p.EntityType, p.PermissionName, p.ConditionExpression)
	}

	fmt.Println("\nExample completed successfully!")
	return nil
}