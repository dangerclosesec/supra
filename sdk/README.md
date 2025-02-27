# Supra Permission SDK

A Go client SDK for the Supra permission service. This SDK provides a simple interface to interact with the permission service API.

## Features

- Check permissions between subjects and objects
- Create, retrieve, and manage entities
- Create relations between entities
- Test relations between entities
- Create and manage permission definitions
- Test rules with parameters
- List permission and rule definitions

## Installation

```bash
go get github.com/dangerclosesec/supra/sdk/client
```

## Basic Usage

```go
import (
    "context"
    "fmt"
    "time"
    
    "github.com/dangerclosesec/supra/sdk/client"
)

func main() {
    // Initialize the client
    c := client.NewClient(&client.Config{
        BaseURL: "http://localhost:4780",
        Timeout: 10 * time.Second,
    })
    
    // Create a context with timeout
    ctx, cancel := context.WithTimeout(context.Background(), 5 * time.Second)
    defer cancel()
    
    // Check a permission
    resp, err := c.CheckPermission(ctx, &client.CheckPermissionRequest{
        SubjectType: "user",
        SubjectID:   "123",
        Permission:  "read",
        ObjectType:  "document",
        ObjectID:    "456",
    })
    
    if err != nil {
        fmt.Printf("Error checking permission: %v\n", err)
        return
    }
    
    fmt.Printf("Permission allowed: %v\n", resp.Allowed)
}
```

## Examples

See the [examples directory](./examples) for complete examples of how to use the SDK.

## API Reference

### Client Configuration

```go
// Create a client with default configuration
c := client.NewClient(nil)

// Create a client with custom configuration
c := client.NewClient(&client.Config{
    BaseURL:    "http://localhost:4780",
    HTTPClient: &http.Client{Timeout: 20 * time.Second},
    Timeout:    10 * time.Second,
})
```

### Permission Operations

```go
// Check if a subject has permission on an object
resp, err := c.CheckPermission(ctx, &client.CheckPermissionRequest{
    SubjectType: "user",
    SubjectID:   "123",
    Permission:  "read",
    ObjectType:  "document",
    ObjectID:    "456",
    Context:     map[string]interface{}{"key": "value"}, // Optional context
})

// Create a permission definition
perm, err := c.CreatePermission(ctx, &client.CreatePermissionRequest{
    EntityType:          "document",
    PermissionName:      "read",
    ConditionExpression: "owner OR viewer",
    Description:         "Can read the document", // Optional
})

// List all permission definitions
permissions, err := c.ListPermissionDefinitions(ctx)
```

### Entity Operations

```go
// Create an entity
entity, err := c.CreateEntity(ctx, &client.CreateEntityRequest{
    Type:       "user",
    ExternalID: "123",
    Properties: map[string]interface{}{
        "name":  "John Doe",
        "email": "john@example.com",
    },
})

// Get an entity
entity, err := c.GetEntity(ctx, "user", "123")
```

### Relation Operations

```go
// Create a relation between entities
relation, err := c.CreateRelation(ctx, &client.CreateRelationRequest{
    SubjectType: "user",
    SubjectID:   "123",
    Relation:    "owner",
    ObjectType:  "document",
    ObjectID:    "456",
})

// Test if a relation exists
resp, err := c.TestRelation(ctx, &client.TestRelationRequest{
    SubjectType: "user",
    SubjectID:   "123",
    Relation:    "owner",
    ObjectType:  "document",
    ObjectID:    "456",
    Direction:   "both", // Optional: "normal", "reverse", or "both"
})
```

### Rule Operations

```go
// Test a rule with parameters
resp, err := c.TestRule(ctx, &client.TestRuleRequest{
    RuleName: "isAdmin",
    Parameters: map[string]interface{}{
        "user_id": "admin",
    },
})

// List all rule definitions
rules, err := c.ListRuleDefinitions(ctx)
```

## Error Handling

The SDK provides enhanced error handling with structured error responses from the API. All errors are categorized with error codes, messages, and detailed information to help debug issues.

### Standard Error Handling

All API methods return errors using Go's standard error interface. Each error contains detailed information about what went wrong.

```go
resp, err := c.CheckPermission(ctx, req)
if err != nil {
    // Handle connection or request error
    fmt.Printf("Error: %v\n", err)
    return err
}

if resp.Error != "" {
    // Handle API-level error
    return errors.New(resp.Error)
}
```

### API Error Types

For more granular error handling, you can use type assertion to check for specific API error types:

```go
entity, err := c.CreateEntity(ctx, &client.CreateEntityRequest{
    Type:       "user",
    ExternalID: "alice",
    Properties: map[string]interface{}{
        "name": "Alice",
    },
})

if err != nil {
    // Check if this is an APIError
    var apiErr *client.APIError
    if errors.As(err, &apiErr) {
        // Handle specific error codes
        switch apiErr.Code {
        case "entity_already_exists":
            fmt.Printf("Entity already exists: %s\n", apiErr.Message)
            // Maybe fetch the existing entity instead
            existingEntity, fetchErr := c.GetEntity(ctx, "user", "alice")
            if fetchErr == nil {
                // Use the existing entity
                entity = existingEntity
            }
        case "missing_fields":
            fmt.Printf("Missing required fields: %s\n", apiErr.Message) 
        default:
            fmt.Printf("API error: [%s] %s - %s\n", 
                apiErr.Code, apiErr.Message, apiErr.Details)
        }
    } else {
        // Handle non-API errors (network, timeout, etc.)
        fmt.Printf("Error: %v\n", err)
        return err
    }
}
```

### Common Error Codes

The API returns standardized error codes that you can handle in your application:

| Code                  | Description                                       | HTTP Status |
|-----------------------|---------------------------------------------------|-------------|
| `invalid_request`     | Invalid request format or structure               | 400         |
| `missing_fields`      | Required fields are missing                       | 400         |
| `entity_not_found`    | The requested entity does not exist               | 404         |
| `entity_already_exists` | Entity with the same type and ID already exists | 409         |
| `permission_not_found`| The requested permission does not exist           | 404         |
| `internal_error`      | Server encountered an error processing the request| 500         |

## Development

### Running Tests

```bash
go test -v ./sdk/client
```

The tests are designed to work with a mock server, so no actual permission service is needed to run them.

## License

This SDK is released under the same license as the Supra project.