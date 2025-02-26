package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
)

// OrphanedRelationship represents a relationship with a missing entity
type OrphanedRelationship struct {
	ID          int64     `json:"id"`
	SubjectType string    `json:"subject_type"`
	SubjectID   string    `json:"subject_id"`
	Relation    string    `json:"relation"`
	ObjectType  string    `json:"object_type"`
	ObjectID    string    `json:"object_id"`
	CreatedAt   time.Time `json:"created_at"`
	MissingType string    `json:"missing_type"` // "subject", "object", or "both"
}

// addHealthCheckEndpoints adds endpoints for system health checks
func (s *AuthzService) addHealthCheckEndpoints(mux *http.ServeMux) {
	// Endpoint to find orphaned relationships
	mux.HandleFunc("/api/health/orphaned-relationships", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
		defer cancel()

		// Query for relationships with missing entities
		query := `
			SELECT 
                r.id, 
                r.subject_type, 
                r.subject_id, 
                r.relation, 
                r.object_type, 
                r.object_id, 
                r.created_at,
                CASE 
                    WHEN NOT sub_exists.exists AND NOT obj_exists.exists THEN 'both'
                    WHEN NOT sub_exists.exists THEN 'subject'
                    WHEN NOT obj_exists.exists THEN 'object'
                    ELSE 'none' -- This should never happen
                END AS missing_type
            FROM 
                relations r
            LEFT JOIN LATERAL (
                SELECT EXISTS (
                    SELECT 1 FROM entities 
                    WHERE type = r.subject_type AND external_id = r.subject_id
                ) AS exists
            ) sub_exists ON true
            LEFT JOIN LATERAL (
                SELECT EXISTS (
                    SELECT 1 FROM entities 
                    WHERE type = r.object_type AND external_id = r.object_id
                ) AS exists
            ) obj_exists ON true
            WHERE 
                (NOT sub_exists.exists OR NOT obj_exists.exists)
            ORDER BY 
                r.created_at DESC
            LIMIT 100;
		`

		rows, err := s.graph.Pool.Query(ctx, query)
		if err != nil {
			log.Printf("Error querying orphaned relationships: %v", err)
			http.Error(w, "Error querying orphaned relationships", http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		var orphanedRelationships []OrphanedRelationship
		for rows.Next() {
			var rel OrphanedRelationship
			if err := rows.Scan(
				&rel.ID,
				&rel.SubjectType,
				&rel.SubjectID,
				&rel.Relation,
				&rel.ObjectType,
				&rel.ObjectID,
				&rel.CreatedAt,
				&rel.MissingType,
			); err != nil {
				log.Printf("Error scanning orphaned relationship: %v", err)
				continue
			}
			orphanedRelationships = append(orphanedRelationships, rel)
		}

		// Get summary statistics
		summaryQuery := `
            SELECT 
                COUNT(*) as total_orphaned,
                SUM(CASE WHEN NOT sub_exists.exists AND NOT obj_exists.exists THEN 1 ELSE 0 END) as both_missing,
                SUM(CASE WHEN NOT sub_exists.exists AND obj_exists.exists THEN 1 ELSE 0 END) as subject_missing,
                SUM(CASE WHEN sub_exists.exists AND NOT obj_exists.exists THEN 1 ELSE 0 END) as object_missing
            FROM 
                relations r
            LEFT JOIN LATERAL (
                SELECT EXISTS (
                    SELECT 1 FROM entities 
                    WHERE type = r.subject_type AND external_id = r.subject_id
                ) AS exists
            ) sub_exists ON true
            LEFT JOIN LATERAL (
                SELECT EXISTS (
                    SELECT 1 FROM entities 
                    WHERE type = r.object_type AND external_id = r.object_id
                ) AS exists
            ) obj_exists ON true
            WHERE 
                (NOT sub_exists.exists OR NOT obj_exists.exists);
        `

		var summary struct {
			TotalOrphaned  int `json:"total_orphaned"`
			BothMissing    int `json:"both_missing"`
			SubjectMissing int `json:"subject_missing"`
			ObjectMissing  int `json:"object_missing"`
		}

		err = s.graph.Pool.QueryRow(ctx, summaryQuery).Scan(
			&summary.TotalOrphaned,
			&summary.BothMissing,
			&summary.SubjectMissing,
			&summary.ObjectMissing,
		)
		if err != nil {
			log.Printf("Error getting orphaned relationships summary: %v", err)
			// Continue with what we have
		}

		// Return the results
		jsonResponse(w, map[string]interface{}{
			"orphaned_relationships": orphanedRelationships,
			"summary":                summary,
		}, http.StatusOK)
	})

	// Endpoint to auto-fix orphaned relationships
	mux.HandleFunc("/api/health/fix-orphaned-relationship", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req struct {
			RelationID int64 `json:"relation_id"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request format", http.StatusBadRequest)
			return
		}

		if req.RelationID == 0 {
			http.Error(w, "Relation ID is required", http.StatusBadRequest)
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
		defer cancel()

		// Get the relationship details
		var rel struct {
			SubjectType string
			SubjectID   string
			ObjectType  string
			ObjectID    string
		}

		err := s.graph.Pool.QueryRow(ctx, `
			SELECT subject_type, subject_id, object_type, object_id
			FROM relations
			WHERE id = $1
		`, req.RelationID).Scan(&rel.SubjectType, &rel.SubjectID, &rel.ObjectType, &rel.ObjectID)

		if err != nil {
			log.Printf("Error retrieving relationship: %v", err)
			http.Error(w, "Failed to retrieve relationship", http.StatusInternalServerError)
			return
		}

		// Check which entities are missing
		var subjectExists, objectExists bool
		err = s.graph.Pool.QueryRow(ctx, `
			SELECT EXISTS (
				SELECT 1 FROM entities 
				WHERE type = $1 AND external_id = $2
			)
		`, rel.SubjectType, rel.SubjectID).Scan(&subjectExists)

		if err != nil {
			log.Printf("Error checking subject existence: %v", err)
		}

		err = s.graph.Pool.QueryRow(ctx, `
			SELECT EXISTS (
				SELECT 1 FROM entities 
				WHERE type = $1 AND external_id = $2
			)
		`, rel.ObjectType, rel.ObjectID).Scan(&objectExists)

		if err != nil {
			log.Printf("Error checking object existence: %v", err)
		}

		// Create missing entities
		var created []string

		if !subjectExists {
			_, err := s.graph.CreateEntity(ctx, rel.SubjectType, rel.SubjectID, map[string]interface{}{
				"name":         rel.SubjectID,
				"auto_created": true,
			})
			if err != nil {
				log.Printf("Failed to auto-create subject entity: %v", err)
			} else {
				created = append(created, fmt.Sprintf("%s:%s", rel.SubjectType, rel.SubjectID))
			}
		}

		if !objectExists {
			_, err := s.graph.CreateEntity(ctx, rel.ObjectType, rel.ObjectID, map[string]interface{}{
				"name":         rel.ObjectID,
				"auto_created": true,
			})
			if err != nil {
				log.Printf("Failed to auto-create object entity: %v", err)
			} else {
				created = append(created, fmt.Sprintf("%s:%s", rel.ObjectType, rel.ObjectID))
			}
		}

		// Return success
		jsonResponse(w, map[string]interface{}{
			"success":          true,
			"entities_created": created,
		}, http.StatusOK)
	})
}
