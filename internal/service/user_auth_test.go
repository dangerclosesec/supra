package service_test

import (
	"context"
	"testing"
	"time"

	"github.com/dangerclosesec/supra/internal/auth"
	"github.com/dangerclosesec/supra/internal/mocks"
	"github.com/dangerclosesec/supra/internal/model"
	"github.com/dangerclosesec/supra/internal/service"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestUserLogin(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	userID := uuid.New()
	testUser := &model.User{
		ID:        userID,
		Email:     "test@example.com",
		FirstName: "Test",
		LastName:  "User",
		Status:    model.StatusActive,
	}

	hasher := auth.NewPasswordHasher()
	hashedPassword, _ := hasher.Hash("correct_password")

	passwordFactor := &model.UserFactor{
		ID:         uuid.New(),
		UserID:     userID,
		FactorType: model.FactorHashpass,
		Material:   hashedPassword,
		IsActive:   true,
	}

	totpFactor := &model.UserFactor{
		ID:         uuid.New(),
		UserID:     userID,
		FactorType: model.FactorTOTP,
		Material:   "ABCDEFGHIJKLMNOP",
		IsActive:   true,
	}

	t.Run("successful MFA login flow", func(t *testing.T) {
		userRepo := mocks.NewMockUserRepositoryIface(ctrl)
		factorRepo := mocks.NewMockUserFactorRepositoryIface(ctrl)

		// Initial password verification expectations
		gomock.InOrder(
			userRepo.EXPECT().
				FindByEmail(gomock.Any(), testUser.Email).
				Return(testUser, nil),

			factorRepo.EXPECT().
				FindByUserAndType(gomock.Any(), userID, model.FactorHashpass).
				Return(passwordFactor, nil),

			factorRepo.EXPECT().
				Update(gomock.Any(), gomock.Any()).
				Return(nil),

			factorRepo.EXPECT().
				FindActiveByUser(gomock.Any(), userID).
				Return([]*model.UserFactor{passwordFactor, totpFactor}, nil),
		)

		// MFA verification expectations
		gomock.InOrder(
			factorRepo.EXPECT().
				FindByID(gomock.Any(), totpFactor.ID).
				Return(totpFactor, nil),

			factorRepo.EXPECT().
				Update(gomock.Any(), gomock.Any()).
				Return(nil),

			userRepo.EXPECT().
				FindByID(gomock.Any(), userID).
				Return(testUser, nil),
		)

		svc := service.NewUserService(
			userRepo,
			factorRepo,
			nil,
			hasher,
			auth.NewTokenManager("test_secret", time.Hour),
			nil,
			service.NewUserFactorService(factorRepo),
			service.NewCacheService(service.CacheConfig{
				TTL:         5 * time.Minute,
				CleanupFreq: time.Minute,
			}),
			nil,
		)

		// Phase 1: Password verification
		result, err := svc.VerifyPassword(context.Background(), service.LoginInput{
			Email:    testUser.Email,
			Password: "correct_password",
		})

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Empty(t, result.Token, "Token should be empty when MFA is required")

		// Phase 2: MFA verification
		finalResult, err := svc.VerifyMFAAndLogin(context.Background(), service.MFAVerifyInput{
			UserID:     userID,
			FactorID:   totpFactor.ID,
			FactorCode: totpFactor.Material,
		})

		assert.NoError(t, err)
		assert.NotNil(t, finalResult)
		assert.NotEmpty(t, finalResult.Token, "Token should be present after successful MFA")
		assert.Equal(t, testUser.ID, finalResult.User.ID)
	})
}
