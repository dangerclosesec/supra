package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/dangerclosesec/supra/internal/domain"
	"github.com/dangerclosesec/supra/internal/model"
	"github.com/google/uuid"
)

type LoginInput struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginOutput struct {
	User  *model.User `json:"user"`
	Token string      `json:"token"`
}

func (s *UserService) Login(ctx context.Context, input LoginInput) (*LoginOutput, error) {
	// Find the user
	user, err := s.repo.FindByEmail(ctx, input.Email)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			return nil, domain.ErrInvalidCredentials
		}
		return nil, err
	}

	// Verify the password using the factor service
	verified, err := s.factorService.VerifyPassword(ctx, user.ID, input.Password)
	if err != nil {
		return nil, fmt.Errorf("verifying password: %w", err)
	}

	if !verified {
		return nil, domain.ErrInvalidCredentials
	}

	// Generate token
	token, err := s.tokenManager.Generate(user.ID.String(), user.Email)
	if err != nil {
		return nil, fmt.Errorf("generating token: %w", err)
	}

	return &LoginOutput{
		User:  user,
		Token: token,
	}, nil
}

type LogoutInput struct {
	UserID uuid.UUID
}

func (s *UserService) Logout(ctx context.Context, input LogoutInput) error {
	return nil
}

type MFAVerifyInput struct {
	UserID     uuid.UUID `json:"user_id"`
	FactorID   uuid.UUID `json:"factor_id"`
	FactorCode string    `json:"factor_code"`
}

// VerifyMFAAndLogin verifies the MFA code and completes the login process
func (s *UserService) VerifyMFAAndLogin(ctx context.Context, input MFAVerifyInput) (*LoginOutput, error) {
	// Verify the factor
	err := s.factorService.VerifyFactor(ctx, input.UserID, input.FactorID, input.FactorCode)
	if err != nil {
		return nil, fmt.Errorf("verifying factor: %w", err)
	}

	// Get user details
	user, err := s.repo.FindByID(ctx, input.UserID)
	if err != nil {
		return nil, fmt.Errorf("finding user: %w", err)
	}

	// Generate token
	token, err := s.tokenManager.Generate(user.ID.String(), user.Email)
	if err != nil {
		return nil, fmt.Errorf("generating token: %w", err)
	}

	return &LoginOutput{
		User:  user,
		Token: token,
	}, nil
}
