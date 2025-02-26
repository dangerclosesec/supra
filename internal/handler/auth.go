// internal/handler/auth.go
package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/dangerclosesec/supra/internal/domain"
	"github.com/dangerclosesec/supra/internal/model"
	"github.com/dangerclosesec/supra/internal/service"
	chmw "github.com/go-chi/chi/v5/middleware"
)

type AuthHandler struct {
	userService  *service.UserService
	cacheService *service.CacheService
}

func NewAuthHandler(userService *service.UserService, cacheService *service.CacheService) *AuthHandler {
	return &AuthHandler{
		userService:  userService,
		cacheService: cacheService,
	}
}

type SignupResponse struct {
	BaseResponse
	User  *model.User `json:"user" sanitize:"user"`
	Token string      `json:"token"`
}

func (h *AuthHandler) SignupHandler(w http.ResponseWriter, r *http.Request) {
	// Validates that we're receiving a POST request
	if r.Method != http.MethodPost && r.Method != http.MethodGet {
		h.respondWithError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	if r.Method == http.MethodGet {
		nonce, err := h.userService.GenerateNonce(r.Context())
		if err != nil {
			h.respondWithError(w, http.StatusInternalServerError, "Failed to generate nonce")
			return
		}

		h.respondWithJSON(w, http.StatusOK, map[string]string{"nonce": nonce})
		return
	}

	// Check for nonce query string parameter
	nonce := r.URL.Query().Get("nonce")
	if nonce == "" {
		h.respondWithError(w, http.StatusBadRequest, "Nonce is required")
		return
	}

	// Verify nonce against cache service
	exists, err := h.cacheService.CheckNonce(r.Context(), nonce)
	if err != nil || !exists {
		h.respondWithError(w, http.StatusBadRequest, "Invalid or expired nonce")
		return
	}

	// Parses the request body
	var input service.SignupInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		h.respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	defer r.Body.Close()

	// Calls the service layer to handle the signup
	output, err := h.userService.Signup(r.Context(), input)
	if err != nil {
		slog.ErrorContext(r.Context(), "User registration error", "error", err, "requestID", chmw.GetReqID(r.Context()))
		switch {
		case errors.Is(err, domain.ErrEmailAlreadyExists):
			h.respondWithError(w, http.StatusConflict, "Email already exists")
		case errors.Is(err, domain.ErrInvalidInput):
			h.respondWithError(w, http.StatusBadRequest, err.Error())
		case errors.Is(err, domain.ErrPasswordTooWeak):
			h.respondWithError(w, http.StatusBadRequest, "Password does not meet requirements")
		default:
			h.respondWithError(w, http.StatusInternalServerError, "Internal server error")
		}
		return
	}

	// Returns successful response
	h.respondWithJSON(w, http.StatusCreated, SignupResponse{
		User:  output.User,
		Token: output.Token})
}

type LoginStatus string

const (
	LoginStatusSuccess     LoginStatus = "success"
	LoginStatusMFARequired LoginStatus = "mfa_required"
	LoginStatusFailed      LoginStatus = "login_failed"
)

type LoginResponse struct {
	BaseResponse
	Status     LoginStatus `json:"status"`
	User       *model.User `json:"user,omitempty" sanitize:"user"`
	Token      string      `json:"token,omitempty"`
	Error      string      `json:"error,omitempty"`
	MFADetails *MFADetails `json:"mfa_details,omitempty"`
}

type MFADetails struct {
	UserID           string   `json:"user_id"`
	Nonce            string   `json:"nonce"`
	AvailableFactors []string `json:"available_factors"`
}

func (h *AuthHandler) LoginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.respondWithError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var input service.LoginInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		h.respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	defer r.Body.Close()

	// First phase: Password verification
	output, err := h.userService.VerifyPassword(r.Context(), input)
	if err != nil {
		slog.ErrorContext(r.Context(), "User login error", "error", err, "requestID", chmw.GetReqID(r.Context()))
		switch {
		case errors.Is(err, domain.ErrInvalidCredentials):
			h.respondWithJSON(w, http.StatusUnauthorized, LoginResponse{
				Status: LoginStatusFailed,
				Error:  "Invalid email or password",
			})
		default:
			h.respondWithError(w, http.StatusInternalServerError, "Internal server error")
		}
		return
	}

	// Check for additional factors
	factors, err := h.userService.GetActiveFactors(r.Context(), output.User.ID)
	if err != nil {
		slog.ErrorContext(r.Context(), "Error fetching user factors", "error", err, "requestID", chmw.GetReqID(r.Context()))
		h.respondWithError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	// Filter out hashpass factor
	var additionalFactors []string
	for _, factor := range factors {
		if factor.FactorType != model.FactorHashpass && factor.IsActive {
			additionalFactors = append(additionalFactors, string(factor.FactorType))
		}
	}

	if len(additionalFactors) > 0 {
		// Generate MFA nonce
		nonce, err := h.userService.GenerateNonce(r.Context())
		if err != nil {
			slog.ErrorContext(r.Context(), "Error generating nonce", "error", err, "requestID", chmw.GetReqID(r.Context()))
			h.respondWithError(w, http.StatusInternalServerError, "Internal server error")
			return
		}

		// Cache the nonce with user ID
		err = h.cacheService.Set(r.Context(), fmt.Sprintf("mfa_nonce:%s", output.User.ID), nonce)
		if err != nil {
			slog.ErrorContext(r.Context(), "Error caching nonce", "error", err, "requestID", chmw.GetReqID(r.Context()))
			h.respondWithError(w, http.StatusInternalServerError, "Internal server error")
			return
		}

		h.respondWithJSON(w, http.StatusOK, LoginResponse{
			BaseResponse: BaseResponse{Ok: true},
			Status:       LoginStatusMFARequired,
			MFADetails: &MFADetails{
				UserID:           output.User.ID.String(),
				Nonce:            nonce,
				AvailableFactors: additionalFactors,
			},
		})
		return
	}

	// No additional factors required, proceed with login
	h.respondWithJSON(w, http.StatusOK, LoginResponse{
		BaseResponse: BaseResponse{Ok: true},
		Status:       LoginStatusSuccess,
		User:         output.User,
		Token:        output.Token,
	})
}

func (h *AuthHandler) LogoutHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.respondWithError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var input service.LogoutInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		h.respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	defer r.Body.Close()

	if err := h.userService.Logout(r.Context(), input); err != nil {
		slog.ErrorContext(r.Context(), "User logout error", "error", err, "requestID", chmw.GetReqID(r.Context()))
		h.respondWithError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	expiration := time.Now().Add(-1 * time.Hour)
	cookie := http.Cookie{Name: "token", Value: "", Expires: expiration}

	http.SetCookie(w, &cookie)

	h.respondWithJSON(w, http.StatusOK, map[string]string{"message": "User logged out successfully"})
}

func (h *AuthHandler) VerifyHandler(w http.ResponseWriter, r *http.Request) {
	var input service.VerifyInput

	query := r.URL.Query()
	input.Code = query.Get("code")
	input.UserID = query.Get("user")

	if err := h.userService.VerifyEmail(r.Context(), input); err != nil {
		slog.ErrorContext(r.Context(), "User verification error", "error", err, "requestID", chmw.GetReqID(r.Context()))
		switch {
		case errors.Is(err, domain.ErrUserNotFound):
			h.respondWithError(w, http.StatusNotFound, "User not found")
		case errors.Is(err, domain.ErrInvalidVerificationCode):
			h.respondWithError(w, http.StatusBadRequest, "Invalid verification code")
		case errors.Is(err, domain.ErrAlreadyVerified):
			h.respondWithError(w, http.StatusBadRequest, "User already verified")
		default:
			h.respondWithError(w, http.StatusInternalServerError, "Internal server error")
		}
		return
	}

	h.respondWithJSON(w, http.StatusOK, map[string]string{"message": "User verified successfully"})
}

func (h *AuthHandler) respondWithError(w http.ResponseWriter, code int, message string) {
	h.respondWithJSON(w, code, ErrorResponse{Error: message})
}

func (h *AuthHandler) respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
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
