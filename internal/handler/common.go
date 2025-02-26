package handler

import (
	"encoding/json"
	"net/http"
)

// UserIDKey is the context key for the user ID
type contextKey string

const UserIDKey = contextKey("userID")

type ErrorResponse struct { // TypeGen: ErrorResponse
	BaseResponse
	Error   string    `json:"error"`
	Details *[]string `json:"details,omitempty"`
	Code    *string   `json:"error_code,omitempty"`
	Link    *string   `json:"error_link,omitempty"`
}

type BaseResponse struct { // TypeGen: DefaultResponse
	Ok bool `json:"ok"`
}

// respondWithError sends an error response with a message
func respondWithError(w http.ResponseWriter, code int, message string) {
	respondWithJSON(w, code, ErrorResponse{Error: message})
}

// respondWithJSON sends a JSON response
func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	// Sets content type header
	w.Header().Set("Content-Type", "application/json")

	// Sets the HTTP status code
	w.WriteHeader(code)

	// Encodes the response
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		// If encoding fails, logs the error and sends a plain text response
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}
