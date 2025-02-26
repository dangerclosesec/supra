// internal/middleware/auth.go
package middleware

// Usage example:
// func ExampleUsage() {
// 	// In your main.go or router setup:
// 	r := chi.NewRouter()

// 	// Protected routes
// 	r.Group(func(r chi.Router) {
// 		// Apply auth middleware to all routes in this group
// 		r.Use(AuthMiddleware(tokenManager))

// 		// Protected endpoints
// 		r.Get("/factors", factorHandler.ListFactors)
// 		r.Post("/factors", factorHandler.CreateFactor)
// 	})
// }

// // Example of accessing user ID in a handler:
// func ExampleHandler(w http.ResponseWriter, r *http.Request) {
// 	// Get user ID from context
// 	userID := r.Context().Value(UserIDKey).(string)

// 	// Use the user ID
// 	fmt.Printf("Request from user: %s\n", userID)
// }

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/dangerclosesec/supra/internal/auth"
)

type UserContextKey string

var UserIDKey UserContextKey = "supra_user_id"

// AuthMiddleware creates a middleware that validates JWT tokens
func AuthMiddleware(tokenManager *auth.TokenManager) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Extract token from Authorization header
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				respondWithError(w, http.StatusUnauthorized, "No authorization header")
				return
			}

			// Check Bearer prefix
			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || parts[0] != "Bearer" {
				respondWithError(w, http.StatusUnauthorized, "Invalid authorization header")
				return
			}

			// Validate token
			claims, err := tokenManager.Validate(parts[1])
			if err != nil {
				respondWithError(w, http.StatusUnauthorized, "Invalid token")
				return
			}

			// Create new context with user ID
			ctx := context.WithValue(r.Context(), UserIDKey, claims.UserID)

			// Call next handler with new context
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// respondWithError sends a JSON error response
func respondWithError(w http.ResponseWriter, code int, message string) {
	respondWithJSON(w, code, map[string]string{"error": message})
}

// respondWithJSON sends a JSON response
func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(payload)
}
