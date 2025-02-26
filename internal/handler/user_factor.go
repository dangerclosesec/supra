// internal/handler/user_factor.go
package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/dangerclosesec/supra/internal/domain"
	"github.com/dangerclosesec/supra/internal/model"
	"github.com/dangerclosesec/supra/internal/service"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type UserFactorHandler struct {
	service *service.UserFactorService
}

func NewUserFactorHandler(service *service.UserFactorService) *UserFactorHandler {
	return &UserFactorHandler{
		service: service,
	}
}

// CreateFactorRequest represents the request body for creating a new factor
type CreateFactorRequest struct {
	FactorType model.FactorType `json:"factor_type"`
	Material   string           `json:"material"`
}

// ListFactors returns all factors for the authenticated user
func (h *UserFactorHandler) ListFactors(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(UserIDKey).(string)

	uid, err := uuid.Parse(userID)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid user ID")
		return
	}

	factors, err := h.service.ListFactors(r.Context(), uid)
	if err != nil {
		h.handleError(w, err)
		return
	}

	respondWithJSON(w, http.StatusOK, factors)
}

// CreateFactor creates a new authentication factor
func (h *UserFactorHandler) CreateFactor(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(UserIDKey).(string)

	uid, err := uuid.Parse(userID)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid user ID")
		return
	}

	var req CreateFactorRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	input := service.CreateFactorInput{
		UserID:     uid,
		FactorType: req.FactorType,
		Material:   req.Material,
	}

	factor, err := h.service.CreateFactor(r.Context(), input)
	if err != nil {
		h.handleError(w, err)
		return
	}

	respondWithJSON(w, http.StatusCreated, factor)
}

// VerifyFactorRequest represents the request body for factor verification
type VerifyFactorRequest struct {
	Code string `json:"code"`
}

// VerifyFactor verifies a specific factor
func (h *UserFactorHandler) VerifyFactor(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(UserIDKey).(string)
	factorID := chi.URLParam(r, "id")

	uid, err := uuid.Parse(userID)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid user ID")
		return
	}

	fid, err := uuid.Parse(factorID)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid factor ID")
		return
	}

	var req VerifyFactorRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	if err := h.service.VerifyFactor(r.Context(), uid, fid, req.Code); err != nil {
		h.handleError(w, err)
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]string{
		"message": "Factor verified successfully",
	})
}

// RemoveFactor deactivates a specific factor
func (h *UserFactorHandler) RemoveFactor(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(UserIDKey).(string)
	factorID := chi.URLParam(r, "id")

	uid, err := uuid.Parse(userID)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid user ID")
		return
	}

	fid, err := uuid.Parse(factorID)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid factor ID")
		return
	}

	if err := h.service.RemoveFactor(r.Context(), uid, fid); err != nil {
		h.handleError(w, err)
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]string{
		"message": "Factor removed successfully",
	})
}

// handleError handles common error cases
func (h *UserFactorHandler) handleError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, domain.ErrFactorNotFound):
		respondWithError(w, http.StatusNotFound, "Factor not found")
	case errors.Is(err, domain.ErrFactorAlreadyExists):
		respondWithError(w, http.StatusConflict, "Factor already exists")
	case errors.Is(err, domain.ErrInvalidFactorType):
		respondWithError(w, http.StatusBadRequest, "Invalid factor type")
	case errors.Is(err, domain.ErrInvalidVerificationCode):
		respondWithError(w, http.StatusBadRequest, "Invalid verification code")
	case errors.Is(err, domain.ErrUnauthorized):
		respondWithError(w, http.StatusUnauthorized, "Unauthorized")
	default:
		respondWithError(w, http.StatusInternalServerError, "Internal server error")
	}
}
