package main

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/dangerclosesec/supra/internal/auth/graph"
)

// Permission Path Visualization API

// PermissionPathRequest represents a request to find a permission path
type PermissionPathRequest struct {
	SubjectType string `json:"subject_type"`
	SubjectID   string `json:"subject_id"`
	Permission  string `json:"permission"`
	ObjectType  string `json:"object_type"`
	ObjectID    string `json:"object_id"`
}

// PermissionPathResponse represents the nodes and relationships that form a permission path
type PermissionPathResponse struct {
	Nodes      []Node   `json:"nodes"`
	Links      []Link   `json:"links"`
	Paths      [][]Link `json:"paths"`      // Multiple possible paths
	Expression string   `json:"expression"` // The original permission expression
	Allowed    bool     `json:"allowed"`    // Whether permission is granted
}

// Add permission path endpoint
func (s *AuthzService) addPermissionPathEndpoint(mux *http.ServeMux) {
	mux.HandleFunc("/api/permission-path", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Parse request
		var req PermissionPathRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request format", http.StatusBadRequest)
			return
		}

		// Basic validation
		if req.SubjectType == "" || req.SubjectID == "" ||
			req.Permission == "" || req.ObjectType == "" || req.ObjectID == "" {
			http.Error(w, "Missing required fields", http.StatusBadRequest)
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
		defer cancel()

		// Check if permission exists and get the expression
		var conditionExpr string
		err := s.graph.Pool.QueryRow(ctx, `
			SELECT condition_expression
			FROM permission_definitions
			WHERE entity_type = $1 AND permission_name = $2
		`, req.ObjectType, req.Permission).Scan(&conditionExpr)

		if err != nil {
			http.Error(w, "{\"error\":\"Permission definition not found\"}", http.StatusNotFound)
			return
		}

		// Check if permission is allowed (for informational purposes)
		allowed, _ := s.graph.CheckPermission(ctx, req.SubjectType, req.SubjectID,
			req.Permission, req.ObjectType, req.ObjectID)

		// Parse the expression to understand what to look for
		parser := graph.NewConditionParser(conditionExpr)
		expression, err := parser.Parse()
		if err != nil {
			b, _ := json.Marshal(map[string]string{"error": err.Error()})
			http.Error(w, string(b), http.StatusNotFound)
			return
		}

		// Find relevant nodes and paths based on the expression
		paths, nodes, links, err := s.findPermissionPaths(ctx, expression,
			req.SubjectType, req.SubjectID, req.ObjectType, req.ObjectID)
		if err != nil {
			b, _ := json.Marshal(map[string]string{"error": err.Error()})
			http.Error(w, string(b), http.StatusNotFound)
			return
		}

		// Return permission path data
		response := PermissionPathResponse{
			Nodes:      nodes,
			Links:      links,
			Paths:      paths,
			Expression: conditionExpr,
			Allowed:    allowed,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	})
}

// findPermissionPaths analyzes an expression and finds all possible paths that could grant permission
func (s *AuthzService) findPermissionPaths(
	ctx context.Context,
	expr graph.Expression,
	subjectType, subjectID, objectType, objectID string,
) ([][]Link, []Node, []Link, error) {
	// A map to track unique nodes
	nodeMap := make(map[string]Node)

	// Create initial subject and object nodes
	subjectNodeID := subjectType + ":" + subjectID
	objectNodeID := objectType + ":" + objectID

	// Get entity details for the subject
	subjectEntity, err := s.graph.GetEntity(ctx, subjectType, subjectID)
	if err != nil {
		return nil, nil, nil, err
	}

	// Get entity details for the object
	objectEntity, err := s.graph.GetEntity(ctx, objectType, objectID)
	if err != nil {
		return nil, nil, nil, err
	}

	// Add subject and object to node map
	var subjectLabel, objectLabel string

	// Try to extract name from properties
	var subjectProps, objectProps map[string]interface{}
	subjectProps = subjectEntity.Properties
	objectProps = objectEntity.Properties

	if name, ok := subjectProps["name"].(string); ok {
		subjectLabel = name
	} else {
		subjectLabel = subjectID
	}

	if name, ok := objectProps["name"].(string); ok {
		objectLabel = name
	} else {
		objectLabel = objectID
	}

	nodeMap[subjectNodeID] = Node{
		ID:    subjectNodeID,
		Type:  subjectType,
		Label: subjectLabel,
		Group: subjectType,
	}

	nodeMap[objectNodeID] = Node{
		ID:    objectNodeID,
		Type:  objectType,
		Label: objectLabel,
		Group: objectType,
	}

	// Find all possible paths by analyzing the expression
	return s.findPathsFromExpression(ctx, expr, subjectType, subjectID, objectType, objectID, nodeMap)
}

// findPathsFromExpression analyzes the expression tree to find permission paths
func (s *AuthzService) findPathsFromExpression(
	ctx context.Context,
	expr graph.Expression,
	subjectType, subjectID, objectType, objectID string,
	nodeMap map[string]Node,
) ([][]Link, []Node, []Link, error) {
	var paths [][]Link
	var allLinks []Link

	// Process based on expression type
	switch e := expr.(type) {
	case *graph.AndExpression:
		// For AND expressions, we need both sides to match
		leftPaths, _, leftLinks, err := s.findPathsFromExpression(ctx, e.Left, subjectType, subjectID, objectType, objectID, nodeMap)
		if err != nil {
			return nil, nil, nil, err
		}

		rightPaths, _, rightLinks, err := s.findPathsFromExpression(ctx, e.Right, subjectType, subjectID, objectType, objectID, nodeMap)
		if err != nil {
			return nil, nil, nil, err
		}

		// Combine paths (AND means we need both, so we add them all)
		paths = append(paths, leftPaths...)
		paths = append(paths, rightPaths...)
		allLinks = append(allLinks, leftLinks...)
		allLinks = append(allLinks, rightLinks...)

	case *graph.OrExpression:
		// For OR expressions, either side can match
		leftPaths, _, leftLinks, err := s.findPathsFromExpression(ctx, e.Left, subjectType, subjectID, objectType, objectID, nodeMap)
		if err != nil {
			return nil, nil, nil, err
		}

		rightPaths, _, rightLinks, err := s.findPathsFromExpression(ctx, e.Right, subjectType, subjectID, objectType, objectID, nodeMap)
		if err != nil {
			return nil, nil, nil, err
		}

		// Add all paths
		paths = append(paths, leftPaths...)
		paths = append(paths, rightPaths...)
		allLinks = append(allLinks, leftLinks...)
		allLinks = append(allLinks, rightLinks...)

	case *graph.RelationExpression:
		// This is a direct or indirect relation
		if e.RelationPath == "" {
			// Direct relation - first try subject->object
			directPaths, directLinks, err := s.findDirectRelationPath(ctx, subjectType, subjectID, e.RelationName, objectType, objectID, nodeMap)
			if err != nil {
				return nil, nil, nil, err
			}
			paths = append(paths, directPaths...)
			allLinks = append(allLinks, directLinks...)

			// Also try object->subject direction
			reversePaths, reverseLinks, err := s.findDirectRelationPath(ctx, objectType, objectID, e.RelationName, subjectType, subjectID, nodeMap)
			if err != nil {
				return nil, nil, nil, err
			}
			paths = append(paths, reversePaths...)
			allLinks = append(allLinks, reverseLinks...)
		} else {
			// Indirect relation (e.g., organization.owner)
			indirectPaths, indirectLinks, err := s.findIndirectRelationPath(ctx, subjectType, subjectID, e.RelationPath, e.RelationName, objectType, objectID, nodeMap)
			if err != nil {
				return nil, nil, nil, err
			}
			paths = append(paths, indirectPaths...)
			allLinks = append(allLinks, indirectLinks...)
		}
	}

	// Convert node map to array
	var nodes []Node
	for _, node := range nodeMap {
		nodes = append(nodes, node)
	}

	return paths, nodes, allLinks, nil
}

// findDirectRelationPath finds a direct relation between subject and object
func (s *AuthzService) findDirectRelationPath(
	ctx context.Context,
	subjectType, subjectID, relation, objectType, objectID string,
	nodeMap map[string]Node,
) ([][]Link, []Link, error) {
	var paths [][]Link
	var links []Link

	// Check if the direct relation exists
	var exists bool
	err := s.graph.Pool.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1
			FROM relations
			WHERE subject_type = $1
			AND subject_id = $2
			AND relation = $3
			AND object_type = $4
			AND object_id = $5
		)
	`, subjectType, subjectID, relation, objectType, objectID).Scan(&exists)

	if err != nil {
		return nil, nil, err
	}

	if exists {
		// Create link for this relation
		sourceID := subjectType + ":" + subjectID
		targetID := objectType + ":" + objectID

		link := Link{
			Source: sourceID,
			Target: targetID,
			Type:   relation,
			Label:  relation,
		}

		links = append(links, link)
		paths = append(paths, []Link{link})
	}

	return paths, links, nil
}

// findIndirectRelationPath finds an indirect relation path
func (s *AuthzService) findIndirectRelationPath(
	ctx context.Context,
	subjectType, subjectID, relationPath, relationName, objectType, objectID string,
	nodeMap map[string]Node,
) ([][]Link, []Link, error) {
	var allPaths [][]Link
	var allLinks []Link

	// First try with intermediate entity being the object
	query := `
		WITH RECURSIVE path(subject_type, subject_id, relation, object_type, object_id, path, depth) AS (
			-- Direct relation to the object of type relationPath
			SELECT 
				r.subject_type, r.subject_id, r.relation, r.object_type, r.object_id,
				ARRAY[ROW(r.subject_type, r.subject_id, r.relation, r.object_type, r.object_id)::record],
				1 AS depth
			FROM relations r
			WHERE r.object_type = $1
			AND r.object_id = $2
			
			UNION ALL
			
			-- Follow path to find the subject
			SELECT 
				r.subject_type, r.subject_id, r.relation, r.object_type, r.object_id,
				path || ROW(r.subject_type, r.subject_id, r.relation, r.object_type, r.object_id)::record,
				p.depth + 1
			FROM relations r
			JOIN path p ON r.object_type = p.subject_type AND r.object_id = p.subject_id
			WHERE p.depth < 5 -- Limit recursion depth
		)
		SELECT path
		FROM path
		WHERE relation = $3
		AND subject_type = $4
		AND subject_id = $5
		AND depth <= 5;
	`

	rows, err := s.graph.Pool.Query(ctx, query, relationPath, objectID, relationName, subjectType, subjectID)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	// Process path results
	for rows.Next() {
		var pathRecords []struct {
			SubjectType string
			SubjectID   string
			Relation    string
			ObjectType  string
			ObjectID    string
		}

		if err := rows.Scan(&pathRecords); err != nil {
			return nil, nil, err
		}

		// Convert path records to links
		var pathLinks []Link
		for _, record := range pathRecords {
			sourceID := record.SubjectType + ":" + record.SubjectID
			targetID := record.ObjectType + ":" + record.ObjectID

			// Add nodes to map if not already present
			if _, exists := nodeMap[sourceID]; !exists {
				// Fetch entity details
				entity, err := s.graph.GetEntity(ctx, record.SubjectType, record.SubjectID)
				if err != nil {
					// If entity doesn't exist in database, use default values
					nodeMap[sourceID] = Node{
						ID:    sourceID,
						Type:  record.SubjectType,
						Label: record.SubjectID,
						Group: record.SubjectType,
					}
				} else {
					// Extract name from properties if available
					var props = entity.Properties

					label := record.SubjectID
					if name, ok := props["name"].(string); ok {
						label = name
					}

					nodeMap[sourceID] = Node{
						ID:    sourceID,
						Type:  record.SubjectType,
						Label: label,
						Group: record.SubjectType,
					}
				}
			}

			if _, exists := nodeMap[targetID]; !exists {
				// Fetch entity details
				entity, err := s.graph.GetEntity(ctx, record.ObjectType, record.ObjectID)
				if err != nil {
					// If entity doesn't exist in database, use default values
					nodeMap[targetID] = Node{
						ID:    targetID,
						Type:  record.ObjectType,
						Label: record.ObjectID,
						Group: record.ObjectType,
					}
				} else {
					// Extract name from properties if available
					var props map[string]interface{}
					props = entity.Properties

					label := record.ObjectID
					if name, ok := props["name"].(string); ok {
						label = name
					}

					nodeMap[targetID] = Node{
						ID:    targetID,
						Type:  record.ObjectType,
						Label: label,
						Group: record.ObjectType,
					}
				}
			}

			// Create link
			link := Link{
				Source: sourceID,
				Target: targetID,
				Type:   record.Relation,
				Label:  record.Relation,
			}

			pathLinks = append(pathLinks, link)
			allLinks = append(allLinks, link)
		}

		if len(pathLinks) > 0 {
			allPaths = append(allPaths, pathLinks)
		}
	}

	// If no paths found, try the reverse direction
	if len(allPaths) == 0 {
		query = `
			WITH RECURSIVE path(subject_type, subject_id, relation, object_type, object_id, path, depth) AS (
				-- Start with direct relations
				SELECT 
					r.subject_type, r.subject_id, r.relation, r.object_type, r.object_id,
					ARRAY[ROW(r.subject_type, r.subject_id, r.relation, r.object_type, r.object_id)::record],
					1 AS depth
				FROM relations r
				WHERE r.subject_type = $1
				AND r.subject_id = $2
				
				UNION ALL
				
				-- Follow the path
				SELECT 
					r.subject_type, r.subject_id, r.relation, r.object_type, r.object_id,
					path || ROW(r.subject_type, r.subject_id, r.relation, r.object_type, r.object_id)::record,
					p.depth + 1
				FROM relations r
				JOIN path p ON r.subject_type = p.object_type AND r.subject_id = p.object_id
				WHERE p.depth < 5 -- Limit recursion depth
			)
			SELECT path
			FROM path
			WHERE relation = $3
			AND object_type = $4
			AND object_id = $5
			AND depth <= 5;
		`

		rows, err := s.graph.Pool.Query(ctx, query, relationPath, objectID, relationName, subjectType, subjectID)
		if err != nil {
			return nil, nil, err
		}
		defer rows.Close()

		// Process path results (similar to above)
		for rows.Next() {
			var pathRecords []struct {
				SubjectType string
				SubjectID   string
				Relation    string
				ObjectType  string
				ObjectID    string
			}

			if err := rows.Scan(&pathRecords); err != nil {
				return nil, nil, err
			}

			// Process path records (similar to above)
			var pathLinks []Link
			for _, record := range pathRecords {
				sourceID := record.SubjectType + ":" + record.SubjectID
				targetID := record.ObjectType + ":" + record.ObjectID

				// Add nodes to map (similar to above)
				if _, exists := nodeMap[sourceID]; !exists {
					nodeMap[sourceID] = Node{
						ID:    sourceID,
						Type:  record.SubjectType,
						Label: record.SubjectID,
						Group: record.SubjectType,
					}
				}

				if _, exists := nodeMap[targetID]; !exists {
					nodeMap[targetID] = Node{
						ID:    targetID,
						Type:  record.ObjectType,
						Label: record.ObjectID,
						Group: record.ObjectType,
					}
				}

				// Create link
				link := Link{
					Source: sourceID,
					Target: targetID,
					Type:   record.Relation,
					Label:  record.Relation,
				}

				pathLinks = append(pathLinks, link)
				allLinks = append(allLinks, link)
			}

			if len(pathLinks) > 0 {
				allPaths = append(allPaths, pathLinks)
			}
		}
	}

	return allPaths, allLinks, nil
}
