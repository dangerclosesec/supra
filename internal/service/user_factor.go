// internal/service/user_factor.go
package service

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/dangerclosesec/supra/internal/auth"
	"github.com/dangerclosesec/supra/internal/domain"
	"github.com/dangerclosesec/supra/internal/model"
	"github.com/dangerclosesec/supra/internal/repository"
	"github.com/google/uuid"
)

type UserFactorService struct {
	repo repository.UserFactorRepositoryIface
}

func NewUserFactorService(repo repository.UserFactorRepositoryIface) *UserFactorService {
	return &UserFactorService{
		repo: repo,
	}
}

type CreateFactorInput struct {
	UserID     uuid.UUID
	FactorType model.FactorType
	Material   string
}

func (s *UserFactorService) CreateFactor(ctx context.Context, input CreateFactorInput) (*model.UserFactor, error) {
	factor := &model.UserFactor{
		UserID:     input.UserID,
		FactorType: input.FactorType,
		Material:   input.Material,
		IsActive:   true,
	}

	if err := s.repo.Create(ctx, factor); err != nil {
		return nil, fmt.Errorf("creating factor: %w", err)
	}

	return factor, nil
}

// VerifyPassword checks if the provided password matches the stored hash for the user
func (s *UserFactorService) VerifyPassword(ctx context.Context, userID uuid.UUID, password string) (bool, error) {
	// Find the password factor
	factor, err := s.repo.FindByUserAndType(ctx, userID, model.FactorHashpass)
	if err != nil {
		return false, fmt.Errorf("finding password factor: %w", err)
	}

	if !factor.IsActive {
		return false, domain.ErrInvalidCredentials
	}

	// Verify the password
	hasher := auth.NewPasswordHasher()
	verified, err := hasher.Verify(password, factor.Material)
	if err != nil {
		return false, fmt.Errorf("verifying password: %w", err)
	}

	if verified {
		// Update last used timestamp
		now := time.Now()
		factor.LastUsedAt = &now
		if err := s.repo.Update(ctx, factor); err != nil {
			// Log the error but don't fail the verification
			log.Printf("failed to update last used timestamp: %v", err)
		}
	}

	return verified, nil
}

func (s *UserFactorService) VerifyFactor(ctx context.Context, userID, factorID uuid.UUID, code string) error {
	factor, err := s.repo.FindByID(ctx, factorID)
	if err != nil {
		return fmt.Errorf("finding factor: %w", err)
	}

	if !factor.IsActive {
		return domain.ErrInactiveFactor
	}

	// Check if the factor belongs to the user
	if factor.UserID != userID {
		return domain.ErrInvalidFactor
	}

	// Process the different factor types
	var verificationErr error

	// Implement verification logic for each factor type
	switch factor.FactorType {
	case model.FactorEmail, model.FactorTOTP:
		verificationErr = s.verifyTOTP(ctx, factor, code)
	default:
		return domain.ErrInvalidVerificationCode
	}

	if verificationErr != nil {
		return verificationErr
	}

	// Update the factor with the verification result
	now := time.Now()
	factor.VerifiedAt = &now
	factor.LastUsedAt = &now

	if err := s.repo.Update(ctx, factor); err != nil {
		return fmt.Errorf("updating factor: %w", err)
	}

	return nil
}

func (s *UserFactorService) verifyTOTP(ctx context.Context, factor *model.UserFactor, code string) error {
	// Implement email verification logic here
	if factor.Material != code {
		return domain.ErrInvalidVerificationCode
	}
	return nil
}

func (s *UserFactorService) ListFactors(ctx context.Context, userID uuid.UUID) ([]model.UserFactor, error) {
	return s.repo.ListByUser(ctx, userID)
}

func (s *UserFactorService) DeleteFactor(ctx context.Context, factorID uuid.UUID) error {
	factor, err := s.repo.FindByID(ctx, factorID)
	if err != nil {
		return fmt.Errorf("finding factor: %w", err)
	}

	if err := s.repo.Delete(ctx, factor); err != nil {
		return fmt.Errorf("deleting factor: %w", err)
	}

	return nil
}

func (s *UserFactorService) RemoveFactor(ctx context.Context, userID, factorID uuid.UUID) error {
	return s.repo.RemoveFactor(ctx, userID, factorID)
}
