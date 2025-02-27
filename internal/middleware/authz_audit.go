package middleware

import (
	"context"
	"net/http"

	"github.com/dangerclosesec/supra/internal/auth"
)

// AuthzAuditMiddleware enriches authorization service with HTTP request for audit logging
func AuthzAuditMiddleware(supraService *auth.SupraService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Create a request-aware service copy with the current HTTP request
			requestSupraService, err := auth.NewSupraService(
				"", // Will be ignored since we're copying the existing service
				func(s *auth.SupraService) {
					*s = *supraService            // Copy all settings
					s.SetHTTPRequest(r.Context(), r) // Set the current request
				},
			)
			
			if err != nil {
				// Fall back to the original service if we can't create a request-aware copy
				next.ServeHTTP(w, r)
				return
			}
			
			// Store the request-aware service in the context
			ctx := context.WithValue(r.Context(), "supra_service", requestSupraService)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}