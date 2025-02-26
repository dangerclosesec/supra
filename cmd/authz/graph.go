package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/lib/pq"
)

// GraphData represents the complete visualization data structure
type GraphData struct {
	Nodes []Node `json:"nodes"`
	Links []Link `json:"links"`
}

// Node represents an entity in the graph visualization
type Node struct {
	ID    string `json:"id"`
	Type  string `json:"type"`
	Label string `json:"label"`
	Group string `json:"group"` // For visual grouping/coloring
}

// Link represents a relationship in the graph visualization
type Link struct {
	Source string `json:"source"`
	Target string `json:"target"`
	Type   string `json:"type"`  // Relationship type
	Label  string `json:"label"` // Display label
}

// addGraphVisualizationEndpoint adds the endpoint for graph data visualization
func (s *AuthzService) addGraphVisualizationEndpoint(mux *http.ServeMux) {
	// Endpoint to get graph data for visualization
	mux.HandleFunc("/api/graph", func(w http.ResponseWriter, r *http.Request) {
		// Get optional filters from query parameters
		entityType := r.URL.Query().Get("entity_type")
		entityID := r.URL.Query().Get("entity_id")
		relationshipType := r.URL.Query().Get("relation_type")
		depth := r.URL.Query().Get("depth")

		// Default depth if not provided
		if depth == "" {
			depth = "2" // Default to 2 levels
		}

		ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
		defer cancel()

		// Generate graph data based on filters
		graphData, err := s.generateGraphData(ctx, entityType, entityID, relationshipType, depth)
		if err != nil {
			log.Printf("Error generating graph data: %v", err)
			http.Error(w, "Error generating graph data", http.StatusInternalServerError)
			return
		}

		// Return the graph data as JSON
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(graphData); err != nil {
			log.Printf("Error encoding graph data: %v", err)
		}
	})

	// Serve the visualization HTML page
	// mux.HandleFunc("/visualize", func(w http.ResponseWriter, r *http.Request) {
	// 	w.Header().Set("Content-Type", "text/html")
	// 	w.Write([]byte(graphVisualizationHTML))
	// })
}

func (s *AuthzService) generateGraphData(
	ctx context.Context,
	entityType,
	entityID,
	relationshipType,
	depth string,
) (*GraphData, error) {
	log.Printf("Generating graph data with filters - type: '%s', id: '%s', relation: '%s', depth: '%s'",
		entityType, entityID, relationshipType, depth)

	// Step 1: Start with the seed entities (based on filters)
	var seedEntities []struct {
		Type       string
		ExternalID string
		Properties []byte
	}

	seedQuery := `
		SELECT 
			type, external_id, properties
		FROM 
			entities
		WHERE 
			($1 = '' OR type = $1)
			AND ($2 = '' OR external_id = $2)
	`

	seedRows, err := s.graph.Pool.Query(ctx, seedQuery, entityType, entityID)
	if err != nil {
		return nil, fmt.Errorf("failed to query seed entities: %w", err)
	}
	defer seedRows.Close()

	for seedRows.Next() {
		var entity struct {
			Type       string
			ExternalID string
			Properties []byte
		}

		if err := seedRows.Scan(&entity.Type, &entity.ExternalID, &entity.Properties); err != nil {
			return nil, fmt.Errorf("failed to scan seed entity: %w", err)
		}

		seedEntities = append(seedEntities, entity)
	}

	log.Printf("Found %d seed entities based on filters", len(seedEntities))
	for _, e := range seedEntities {
		log.Printf("  Seed entity: %s:%s", e.Type, e.ExternalID)
	}

	// If no seed entities found and specific filters were provided, return early
	if len(seedEntities) == 0 && (entityType != "" || entityID != "") {
		log.Printf("No seed entities found with the given filters, returning empty graph")
		return &GraphData{
			Nodes: []Node{},
			Links: []Link{},
		}, nil
	}

	// Step 2: Build entity IDs to use in relationship traversal
	var seedEntityIds []string
	var placeholders []string
	var params []interface{}

	for i, entity := range seedEntities {
		seedEntityIds = append(seedEntityIds, fmt.Sprintf("%s:%s", entity.Type, entity.ExternalID))

		// For SQL parameterized query
		placeholders = append(placeholders, fmt.Sprintf("($%d, $%d)", i*2+1, i*2+2))
		params = append(params, entity.Type, entity.ExternalID)
	}

	// If no specific filters, but we're not finding organizations,
	// let's explicitly add them to the seed entities
	if entityType == "" && entityID == "" {
		// Explicitly query for organizations
		orgQuery := `
			SELECT 
				type, external_id, properties
			FROM 
				entities
			WHERE 
				type = 'organization'
			LIMIT 100
		`

		orgRows, err := s.graph.Pool.Query(ctx, orgQuery)
		if err != nil {
			log.Printf("Warning: Failed to query organizations: %v", err)
		} else {
			defer orgRows.Close()

			orgCount := 0
			for orgRows.Next() {
				var org struct {
					Type       string
					ExternalID string
					Properties []byte
				}

				if err := orgRows.Scan(&org.Type, &org.ExternalID, &org.Properties); err != nil {
					log.Printf("Warning: Failed to scan organization: %v", err)
					continue
				}

				orgCount++

				// Add to seed entities if not already included
				orgID := fmt.Sprintf("%s:%s", org.Type, org.ExternalID)
				if !contains(seedEntityIds, orgID) {
					seedEntities = append(seedEntities, org)
					seedEntityIds = append(seedEntityIds, orgID)
				}
			}

			log.Printf("Found %d organization entities to add to graph", orgCount)
		}
	}

	// Step 3: Use a simpler approach - get all directly connected entities within depth
	depthInt, err := strconv.Atoi(depth)
	if err != nil {
		depthInt = 2 // Default depth if parsing fails
	}

	var allEntities = make(map[string]struct {
		Type       string
		ExternalID string
		Properties []byte
	})

	// Add seed entities to the map
	for _, entity := range seedEntities {
		entityKey := fmt.Sprintf("%s:%s", entity.Type, entity.ExternalID)
		allEntities[entityKey] = entity
	}

	// For each level of depth, find connected entities
	currentDepth := 0
	currentEntityIds := seedEntityIds

	for currentDepth < depthInt && len(currentEntityIds) > 0 {
		log.Printf("Processing depth %d with %d entities", currentDepth, len(currentEntityIds))

		// Build query to find all entities connected to current set
		// This approach gets entities in both directions of the relationship
		relatedQuery := `
			-- Get entities that are targets of relations where current entities are subjects
			SELECT DISTINCT
				e.type, e.external_id, e.properties
			FROM
				relations r
			JOIN
				entities e ON r.object_type = e.type AND r.object_id = e.external_id
			WHERE
				(r.subject_type || ':' || r.subject_id) = ANY($1)
				AND ($2 = '' OR r.relation = $2)
				
			UNION
			
			-- Get entities that are subjects of relations where current entities are targets
			SELECT DISTINCT
				e.type, e.external_id, e.properties
			FROM
				relations r
			JOIN
				entities e ON r.subject_type = e.type AND r.subject_id = e.external_id
			WHERE
				(r.object_type || ':' || r.object_id) = ANY($1)
				AND ($2 = '' OR r.relation = $2)
		`

		relatedRows, err := s.graph.Pool.Query(ctx, relatedQuery, pq.Array(currentEntityIds), relationshipType)
		if err != nil {
			return nil, fmt.Errorf("failed to query related entities at depth %d: %w", currentDepth, err)
		}

		// Process related entities
		var newEntityIds []string
		relatedCount := 0

		for relatedRows.Next() {
			relatedCount++
			var entity struct {
				Type       string
				ExternalID string
				Properties []byte
			}

			if err := relatedRows.Scan(&entity.Type, &entity.ExternalID, &entity.Properties); err != nil {
				relatedRows.Close()
				return nil, fmt.Errorf("failed to scan related entity: %w", err)
			}

			entityKey := fmt.Sprintf("%s:%s", entity.Type, entity.ExternalID)

			// Add to all entities if not already included
			if _, exists := allEntities[entityKey]; !exists {
				allEntities[entityKey] = entity
				newEntityIds = append(newEntityIds, entityKey)
			}
		}
		relatedRows.Close()

		log.Printf("Found %d related entities at depth %d, %d new entities",
			relatedCount, currentDepth, len(newEntityIds))

		// Move to next depth
		currentDepth++
		currentEntityIds = newEntityIds
	}

	// Step 4: Process all discovered entities into nodes
	nodeMap := make(map[string]Node)
	var nodeKeys []string

	typeCount := make(map[string]int)

	for entityKey, entity := range allEntities {
		// Extract display name from properties if available
		var props map[string]interface{}
		if err := json.Unmarshal(entity.Properties, &props); err != nil {
			log.Printf("Warning: Failed to parse properties for %s: %v", entityKey, err)
		}

		// Determine label (use name property if available, otherwise use external ID)
		label := entity.ExternalID
		if name, ok := props["name"].(string); ok && name != "" {
			label = name
		}

		// Add to node map
		nodeMap[entityKey] = Node{
			ID:    entityKey,
			Type:  entity.Type,
			Label: label,
			Group: entity.Type,
		}
		nodeKeys = append(nodeKeys, entityKey)

		// Count by type for debugging
		typeCount[entity.Type]++
	}

	// Debug output for node types
	log.Printf("Node type distribution:")
	for t, count := range typeCount {
		log.Printf("  %s: %d nodes", t, count)
	}

	// Step 5: Query for relationships between these entities
	relQuery := `
		SELECT 
			subject_type, subject_id, relation, object_type, object_id
		FROM 
			relations
		WHERE 
			(subject_type || ':' || subject_id) = ANY($1)
			AND (object_type || ':' || object_id) = ANY($1)
			AND ($2 = '' OR relation = $2);
	`

	relRows, err := s.graph.Pool.Query(ctx, relQuery, pq.Array(nodeKeys), relationshipType)
	if err != nil {
		return nil, fmt.Errorf("failed to query relationships: %w", err)
	}
	defer relRows.Close()

	// Process relationship results
	var links []Link
	relCount := make(map[string]int)

	for relRows.Next() {
		var subjectType, subjectID, relation, objectType, objectID string

		if err := relRows.Scan(&subjectType, &subjectID, &relation, &objectType, &objectID); err != nil {
			return nil, fmt.Errorf("failed to scan relationship: %w", err)
		}

		sourceID := subjectType + ":" + subjectID
		targetID := objectType + ":" + objectID

		links = append(links, Link{
			Source: sourceID,
			Target: targetID,
			Type:   relation,
			Label:  relation,
		})

		// Count by relation type for debugging
		relCount[relation]++
	}

	// Debug output for relation types
	log.Printf("Relation type distribution:")
	for r, count := range relCount {
		log.Printf("  %s: %d relations", r, count)
	}

	// Convert node map to array
	var nodes []Node
	for _, node := range nodeMap {
		nodes = append(nodes, node)
	}

	log.Printf("Final graph has %d nodes and %d links", len(nodes), len(links))

	return &GraphData{
		Nodes: nodes,
		Links: links,
	}, nil
}

// Helper function to check if a string is in a slice
func contains(slice []string, str string) bool {
	for _, item := range slice {
		if item == str {
			return true
		}
	}
	return false
}
