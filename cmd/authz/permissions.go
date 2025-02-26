// File to be added to the Go backend to support permission exploration UI

package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// ListPermissionsResponse represents the response for listing permission definitions
type ListPermissionsResponse struct {
	Permissions []PermissionDefinition `json:"permissions"`
}

// PermissionDefinition holds a permission definition with its entity type and expression
type PermissionDefinition struct {
	ID                  int64  `json:"id"`
	EntityType          string `json:"entity_type"`
	PermissionName      string `json:"permission_name"`
	ConditionExpression string `json:"condition_expression"`
	Description         string `json:"description,omitempty"`
	CreatedAt           string `json:"created_at,omitempty"`
}

// ListEntitiesResponse represents the response for listing entity types
type ListEntitiesResponse struct {
	Entities []EntityInfo `json:"entities"`
}

// EntityInfo represents information about an entity type
type EntityInfo struct {
	Type        string `json:"type"`
	DisplayName string `json:"display_name,omitempty"`
	Count       int    `json:"count,omitempty"`
}

// ListRelationsResponse represents the response for listing relations
type ListRelationsResponse struct {
	Relations []RelationInfo `json:"relations"`
}

// RelationInfo represents information about a relation between entities
type RelationInfo struct {
	ID          int64  `json:"id"`
	SubjectType string `json:"subject_type"`
	SubjectID   string `json:"subject_id"`
	Relation    string `json:"relation"`
	ObjectType  string `json:"object_type"`
	ObjectID    string `json:"object_id"`
	CreatedAt   string `json:"created_at,omitempty"`
}

// addSchemaExplorerEndpoints adds endpoints for exploring the permission schema
func (s *AuthzService) addSchemaExplorerEndpoints(mux *http.ServeMux) {
	// Endpoint to get all permission definitions
	mux.HandleFunc("/api/permission-definitions", func(w http.ResponseWriter, r *http.Request) {
		// Only allow GET requests
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Get context with timeout
		ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
		defer cancel()

		// Query the database for all permission definitions
		rows, err := s.graph.Pool.Query(ctx, `
			SELECT 
				id, entity_type, permission_name, condition_expression, description, created_at
			FROM 
				permission_definitions
			ORDER BY
				entity_type, permission_name
		`)
		if err != nil {
			log.Printf("Error retrieving permission definitions: %v", err)
			http.Error(w, "Error retrieving permission definitions", http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		// Process query results
		var permissions []PermissionDefinition
		for rows.Next() {
			var perm PermissionDefinition
			var createdAt interface{} // Handle potential NULL value
			err := rows.Scan(
				&perm.ID,
				&perm.EntityType,
				&perm.PermissionName,
				&perm.ConditionExpression,
				&perm.Description,
				&createdAt,
			)
			if err != nil {
				log.Printf("Error scanning permission definition: %v", err)
				continue
			}

			// Convert createdAt to string if not nil
			if createdAt != nil {
				perm.CreatedAt = createdAt.(time.Time).Format(time.RFC3339)
			}

			permissions = append(permissions, perm)
		}

		// Return the results
		jsonResponse(w, permissions, http.StatusOK)
	})

	// Endpoint to get all entity types
	mux.HandleFunc("/api/entity-types", func(w http.ResponseWriter, r *http.Request) {
		// Only allow GET requests
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Get context with timeout
		ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
		defer cancel()

		// Query distinct entity types from permission definitions
		rows, err := s.graph.Pool.Query(ctx, `
			SELECT DISTINCT entity_type
			FROM permission_definitions
			ORDER BY entity_type
		`)
		if err != nil {
			log.Printf("Error retrieving entity types: %v", err)
			http.Error(w, "Error retrieving entity types", http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		// Process query results
		var entityTypes []string
		for rows.Next() {
			var entityType string
			if err := rows.Scan(&entityType); err != nil {
				log.Printf("Error scanning entity type: %v", err)
				continue
			}
			entityTypes = append(entityTypes, entityType)
		}

		// Also check for entity types in relations (subject types and object types)
		rows, err = s.graph.Pool.Query(ctx, `
			SELECT DISTINCT subject_type
			FROM relations
			UNION
			SELECT DISTINCT object_type
			FROM relations
			ORDER BY 1
		`)
		if err != nil {
			log.Printf("Error retrieving relation entity types: %v", err)
		} else {
			defer rows.Close()
			for rows.Next() {
				var entityType string
				if err := rows.Scan(&entityType); err != nil {
					log.Printf("Error scanning relation entity type: %v", err)
					continue
				}

				// Check if this type is already in our list
				found := false
				for _, t := range entityTypes {
					if t == entityType {
						found = true
						break
					}
				}

				if !found {
					entityTypes = append(entityTypes, entityType)
				}
			}
		}

		// Convert to EntityInfo objects with displayable names
		var entities []EntityInfo
		for _, entityType := range entityTypes {
			// Convert snake_case to Title Case for display name
			displayName := toTitleCase(entityType)

			// Count entities of this type
			var count int
			err := s.graph.Pool.QueryRow(ctx, `
				SELECT COUNT(*) FROM entities WHERE type = $1
			`, entityType).Scan(&count)

			if err != nil {
				log.Printf("Error counting entities of type %s: %v", entityType, err)
				count = 0
			}

			entities = append(entities, EntityInfo{
				Type:        entityType,
				DisplayName: displayName,
				Count:       count,
			})
		}

		// Return the results
		jsonResponse(w, entities, http.StatusOK)
	})

	// Endpoint to get all relations
	mux.HandleFunc("/api/relations", func(w http.ResponseWriter, r *http.Request) {
		// Only allow GET requests
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Get query parameters for filtering
		limitStr := r.URL.Query().Get("limit")
		limit := 100 // Default limit
		if limitStr != "" {
			if n, err := strconv.Atoi(limitStr); err == nil && n > 0 {
				limit = n
			}
		}

		entityType := r.URL.Query().Get("entity_type")
		relationName := r.URL.Query().Get("relation")

		// Get context with timeout
		ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
		defer cancel()

		// Construct base query
		query := `
			SELECT id, subject_type, subject_id, relation, object_type, object_id, created_at
			FROM relations
			WHERE 1=1
		`
		args := []interface{}{}
		argPos := 1

		// Add filters if provided
		if entityType != "" {
			query += fmt.Sprintf(" AND (subject_type = $%d OR object_type = $%d)", argPos, argPos)
			args = append(args, entityType)
			argPos++
		}

		if relationName != "" {
			query += fmt.Sprintf(" AND relation = $%d", argPos)
			args = append(args, relationName)
			argPos++
		}

		// Add limit and order
		query += fmt.Sprintf(" ORDER BY id DESC LIMIT $%d", argPos)
		args = append(args, limit)

		// Query the database for relations
		rows, err := s.graph.Pool.Query(ctx, query, args...)
		if err != nil {
			log.Printf("Error retrieving relations: %v", err)
			http.Error(w, "Error retrieving relations", http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		// Process query results
		var relations []RelationInfo
		for rows.Next() {
			var rel RelationInfo
			var createdAt interface{} // Handle potential NULL value
			err := rows.Scan(
				&rel.ID,
				&rel.SubjectType,
				&rel.SubjectID,
				&rel.Relation,
				&rel.ObjectType,
				&rel.ObjectID,
				&createdAt,
			)
			if err != nil {
				log.Printf("Error scanning relation: %v", err)
				continue
			}

			// Convert createdAt to string if not nil
			if createdAt != nil {
				rel.CreatedAt = createdAt.(time.Time).Format(time.RFC3339)
			}

			relations = append(relations, rel)
		}

		// Return the results
		jsonResponse(w, relations, http.StatusOK)
	})
}

// toTitleCase converts snake_case to Title Case
func toTitleCase(s string) string {
	words := strings.Split(s, "_")
	for i, word := range words {
		if len(word) > 0 {
			words[i] = strings.ToUpper(word[0:1]) + word[1:]
		}
	}
	return strings.Join(words, " ")
}
