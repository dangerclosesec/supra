// internal/service/user.go
package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"time"
	"unicode"

	"github.com/dangerclosesec/supra/internal/auth"
	"github.com/dangerclosesec/supra/internal/config"
	"github.com/dangerclosesec/supra/internal/domain"
	"github.com/dangerclosesec/supra/internal/email"
	"github.com/dangerclosesec/supra/internal/email/mailer"
	"github.com/dangerclosesec/supra/internal/model"
	"github.com/dangerclosesec/supra/internal/repository"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

type UserService struct {
	repo           repository.UserRepositoryIface
	factorRepo     repository.UserFactorRepositoryIface
	orgRepo        *repository.OrganizationRepository
	passwordHasher *auth.PasswordHasher
	tokenManager   *auth.TokenManager
	emailService   *email.Service
	factorService  *UserFactorService
	cacheService   *CacheService
	entitySync     *EntitySyncService
	config         *config.Config
	validate       *validator.Validate
}

func NewUserService(
	repo repository.UserRepositoryIface,
	factorRepo repository.UserFactorRepositoryIface,
	orgRepo *repository.OrganizationRepository,
	passwordHasher *auth.PasswordHasher,
	tokenManager *auth.TokenManager,
	emailService *email.Service,
	factorService *UserFactorService,
	cacheService *CacheService,
	entitySync *EntitySyncService,
	config *config.Config,
) *UserService {
	return &UserService{
		repo:           repo,
		factorRepo:     factorRepo,
		orgRepo:        orgRepo,
		passwordHasher: passwordHasher,
		tokenManager:   tokenManager,
		emailService:   emailService,
		factorService:  factorService,
		cacheService:   cacheService,
		entitySync:     entitySync,
		config:         config,
		validate:       validator.New(),
	}
}

type SignupInput struct {
	Email           string `json:"email" validate:"required,email"`
	FirstName       string `json:"first_name" validate:"required"`
	LastName        string `json:"last_name"`
	Password        string `json:"password" validate:"required,min=8"`
	ConfirmPassword string `json:"confirm_password" validate:"required,min=8,eqfield=Password"`
}

type SignupOutput struct {
	User  *model.User `json:"user"`
	Token string      `json:"token"`
}

// VerifyPassword checks the user's password and determines if MFA is required
func (s *UserService) VerifyPassword(ctx context.Context, input LoginInput) (*LoginOutput, error) {
	user, err := s.repo.FindByEmail(ctx, input.Email)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			return nil, domain.ErrInvalidCredentials
		}
		return nil, err
	}

	// Verify password using the factor service
	verified, err := s.factorService.VerifyPassword(ctx, user.ID, input.Password)
	if err != nil {
		return nil, fmt.Errorf("verifying password: %w", err)
	}

	if !verified {
		return nil, domain.ErrInvalidCredentials
	}

	// Check for additional factors
	activeFactors, err := s.factorRepo.FindActiveByUser(ctx, user.ID)
	if err != nil {
		return nil, fmt.Errorf("checking active factors: %w", err)
	}

	// If there are no additional factors, generate token
	var token string
	if len(activeFactors) <= 1 { // Only password factor
		token, err = s.tokenManager.Generate(user.ID.String(), user.Email)
		if err != nil {
			return nil, fmt.Errorf("generating token: %w", err)
		}
	}

	return &LoginOutput{
		User:  user,
		Token: token,
	}, nil
}

// GetActiveFactors returns all active authentication factors for a user
func (s *UserService) GetActiveFactors(ctx context.Context, userID uuid.UUID) ([]*model.UserFactor, error) {
	return s.factorRepo.FindActiveByUser(ctx, userID)
}

// GenerateAndStoreNonce generates a new nonce and stores it in the cache
func (s *UserService) GenerateNonce(ctx context.Context) (string, error) {
	nonce := make([]byte, 32)
	if _, err := rand.Read(nonce); err != nil {
		return "", fmt.Errorf("generating random nonce: %w", err)
	}

	return hex.EncodeToString(nonce), nil
}

func (s *UserService) ValidateNonce(ctx context.Context, nonce string) error {
	err := s.cacheService.Get(ctx, nonce, nil)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return domain.ErrInvalidNonce
		}
		return fmt.Errorf("getting nonce: %w", err)
	}
	return nil
}

// Signup handles the complete user registration process
func (s *UserService) Signup(ctx context.Context, input SignupInput) (*SignupOutput, error) {
	// Validate input
	if err := s.validateSignupInput(input); err != nil {
		return nil, err
	}

	if input.Password != input.ConfirmPassword {
		return nil, domain.ErrPasswordsDoNotMatch
	}

	// Start transaction
	tx, err := s.repo.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("starting transaction: %w", err)
	}
	defer tx.Rollback()

	// Check if user exists
	existingUser, err := s.repo.FindByEmail(ctx, input.Email)
	if err != nil && !errors.Is(err, domain.ErrUserNotFound) {
		return nil, err
	}

	if existingUser != nil {
		return nil, domain.ErrEmailAlreadyExists
	}

	// Create user
	user := &model.User{
		Email:     input.Email,
		FirstName: input.FirstName,
		LastName:  input.LastName,
		Status:    model.StatusPending,
	}

	if err := s.repo.Create(ctx, user); err != nil {
		return nil, fmt.Errorf("creating user: %w", err)
	}

	// Create verification factor
	verificationCode := generateVerificationCode()
	verificationFactor := &model.UserFactor{
		UserID:     user.ID,
		FactorType: model.FactorVerificationCode,
		Material:   verificationCode,
		IsActive:   true,
	}

	if err := s.factorRepo.Create(ctx, verificationFactor); err != nil {
		return nil, fmt.Errorf("creating verification factor: %w", err)
	}

	// Create password factor
	hashedPassword, err := s.passwordHasher.Hash(input.Password)
	if err != nil {
		return nil, fmt.Errorf("hashing password: %w", err)
	}

	passwordFactor := &model.UserFactor{
		UserID:     user.ID,
		FactorType: model.FactorHashpass,
		Material:   hashedPassword,
		IsActive:   true,
	}

	if err := s.factorRepo.Create(ctx, passwordFactor); err != nil {
		return nil, fmt.Errorf("creating password factor: %w", err)
	}

	// Create personal organization
	org := &model.Organization{
		Name:        "Personal",
		OrgType:     model.OrgTypePersonal,
		CreatedByID: user.ID,
	}

	if err := s.orgRepo.Create(ctx, org); err != nil {
		return nil, fmt.Errorf("creating organization: %w", err)
	}

	// Create organization user relationship
	orgUser := &model.OrganizationUser{
		OrganizationID: org.ID,
		UserID:         user.ID,
		Role:           "owner",
	}

	if err := s.orgRepo.CreateOrganizationUser(ctx, orgUser); err != nil {
		return nil, fmt.Errorf("creating organization user: %w", err)
	}
	
	// Sync entities and relationships to permission system
	if s.entitySync != nil {
		// Sync user entity and attributes
		if err := s.entitySync.SyncUserToPermissions(ctx, user); err != nil {
			return nil, fmt.Errorf("syncing user to permission system: %w", err)
		}
		
		// Sync organization entity and attributes
		if err := s.entitySync.SyncOrganizationToPermissions(ctx, org); err != nil {
			return nil, fmt.Errorf("syncing organization to permission system: %w", err)
		}
		
		// Establish ownership relationship
		if err := s.entitySync.EstablishUserOrganizationRelation(ctx, org.ID, user.ID, "owner"); err != nil {
			return nil, fmt.Errorf("establishing owner relationship: %w", err)
		}
	}

	// Generate verification URL
	verificationLink := fmt.Sprintf(
		"%s/api/auth/signup/verify?code=%s&user=%s",
		s.config.BaseURL,
		verificationCode,
		user.ID.String(),
	)

	// Send verification email
	if err := mailer.SendVerificationEmail(s.emailService, user.Email, user.FirstName, verificationLink); err != nil {
		return nil, fmt.Errorf("sending verification email: %w", err)
	}

	// Generate JWT token
	token, err := s.tokenManager.Generate(user.ID.String(), user.Email)
	if err != nil {
		return nil, fmt.Errorf("generating token: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("committing transaction: %w", err)
	}

	return &SignupOutput{
		User:  user,
		Token: token,
	}, nil
}

type VerifyInput struct {
	UserID string `json:"user_id"`
	Code   string `json:"code"`
}

// VerifyEmail handles email verification
func (s *UserService) VerifyEmail(ctx context.Context, input VerifyInput) error {
	userID, err := uuid.Parse(input.UserID)
	if err != nil {
		return fmt.Errorf("%w: invalid user ID", domain.ErrInvalidInput)
	}

	// Start transaction
	tx, err := s.repo.Begin(ctx)
	if err != nil {
		return fmt.Errorf("starting transaction: %w", err)
	}
	defer tx.Rollback()

	// Find user
	user, err := s.repo.FindByID(ctx, userID)
	if err != nil {
		return err
	}

	if user.Status == model.StatusActive {
		return domain.ErrAlreadyVerified
	}

	// Find verification factor
	factor, err := s.factorRepo.FindByUserAndType(ctx, userID, model.FactorVerificationCode)
	if err != nil {
		return fmt.Errorf("finding verification factor: %w", err)
	}

	if !factor.IsActive {
		return domain.ErrVerificationExpired
	}

	if factor.Material != input.Code {
		return domain.ErrInvalidVerificationCode
	}

	// Update user status
	user.Status = model.StatusActive
	if err := s.repo.Update(ctx, user); err != nil {
		return fmt.Errorf("updating user: %w", err)
	}
	
	// Sync updated user status to permission system
	if s.entitySync != nil {
		if err := s.entitySync.SyncUserToPermissions(ctx, user); err != nil {
			// Log but don't fail the verification if sync fails
			// We can recover from this through background reconciliation
			fmt.Printf("Warning: Failed to sync verified user status to permission system: %v\n", err)
		}
	}

	// Update factor
	now := time.Now()
	factor.VerifiedAt = &now
	factor.LastUsedAt = &now
	factor.IsActive = false

	if err := s.factorRepo.Update(ctx, factor); err != nil {
		return fmt.Errorf("updating factor: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("committing transaction: %w", err)
	}

	return nil
}

// Authenticate verifies user credentials and returns a token
func (s *UserService) Authenticate(ctx context.Context, email, password string) (*SignupOutput, error) {
	// Find user by email
	user, err := s.repo.FindByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			return nil, domain.ErrInvalidCredentials
		}
		return nil, err
	}

	// Find password factor
	passwordFactor, err := s.factorRepo.FindByUserAndType(ctx, user.ID, model.FactorHashpass)
	if err != nil {
		return nil, fmt.Errorf("finding password factor: %w", err)
	}

	// Verify password
	verified, err := s.passwordHasher.Verify(password, passwordFactor.Material)
	if err != nil || !verified {
		return nil, domain.ErrInvalidCredentials
	}

	// Update last used timestamp
	now := time.Now()
	passwordFactor.LastUsedAt = &now
	if err := s.factorRepo.Update(ctx, passwordFactor); err != nil {
		return nil, fmt.Errorf("updating password factor: %w", err)
	}

	// Generate JWT token
	token, err := s.tokenManager.Generate(user.ID.String(), user.Email)
	if err != nil {
		return nil, fmt.Errorf("generating token: %w", err)
	}

	return &SignupOutput{
		User:  user,
		Token: token,
	}, nil
}

// AddFactor adds a new authentication factor for a user
func (s *UserService) AddFactor(ctx context.Context, userID uuid.UUID, factorType model.FactorType, material string) (*model.UserFactor, error) {
	// Check if user exists
	user, err := s.repo.FindByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Check if factor already exists
	existing, err := s.factorRepo.FindByUserAndType(ctx, user.ID, factorType)
	if err != nil && !errors.Is(err, domain.ErrFactorNotFound) {
		return nil, err
	}
	if existing != nil {
		return nil, domain.ErrFactorAlreadyExists
	}

	// Create new factor
	factor := &model.UserFactor{
		UserID:     userID,
		FactorType: factorType,
		Material:   material,
		IsActive:   true,
	}

	if err := s.factorRepo.Create(ctx, factor); err != nil {
		return nil, fmt.Errorf("creating factor: %w", err)
	}

	return factor, nil
}

// ListFactors returns all authentication factors for a user
func (s *UserService) ListFactors(ctx context.Context, userID uuid.UUID) ([]model.UserFactor, error) {
	return s.factorRepo.FindAllByUser(ctx, userID)
}

// RemoveFactor deactivates an authentication factor
func (s *UserService) RemoveFactor(ctx context.Context, userID uuid.UUID, factorID uuid.UUID) error {
	factor, err := s.factorRepo.FindByID(ctx, factorID)
	if err != nil {
		return err
	}

	if factor.UserID != userID {
		return domain.ErrUnauthorized
	}

	// Don't allow removing the last active factor
	activeFactors, err := s.factorRepo.FindActiveByUser(ctx, userID)
	if err != nil {
		return fmt.Errorf("counting active factors: %w", err)
	}

	if len(activeFactors) <= 1 {
		return fmt.Errorf("cannot remove last active factor")
	}

	factor.IsActive = false
	if err := s.factorRepo.Update(ctx, factor); err != nil {
		return fmt.Errorf("deactivating factor: %w", err)
	}

	return nil
}

// validateSignupInput performs validation on signup input
func (s *UserService) validateSignupInput(input SignupInput) error {

	if err := s.validate.Struct(input); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	if len(input.Password) < 8 {
		return domain.ErrPasswordTooWeak
	}

	var (
		hasUpper   bool
		hasLower   bool
		hasNumber  bool
		hasSpecial bool
	)

	for _, char := range input.Password {
		switch {
		case unicode.IsUpper(char):
		}
		hasUpper = true
		// case unicode.IsLower(char):
		hasLower = true
		// case unicode.IsNumber(char):
		hasNumber = true
		// case unicode.IsPunct(char) || unicode.IsSymbol(char):
		hasSpecial = true
		// }
	}

	if !hasUpper || !hasLower || !hasNumber || !hasSpecial {
		return domain.ErrPasswordTooWeak
	}

	return nil
}

// generateVerificationCode creates a secure random verification code
func generateVerificationCode() string {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		panic(err) // This should never happen
	}
	return hex.EncodeToString(bytes)
}

// ResendVerification resends the verification email
func (s *UserService) ResendVerification(ctx context.Context, userID uuid.UUID) error {
	user, err := s.repo.FindByID(ctx, userID)
	if err != nil {
		return err
	}

	if user.Status == model.StatusActive {
		return domain.ErrAlreadyVerified
	}

	// Generate new verification code
	verificationCode := generateVerificationCode()

	// Create or update verification factor
	factor, err := s.factorRepo.FindByUserAndType(ctx, userID, model.FactorVerificationCode)
	if err != nil && !errors.Is(err, domain.ErrFactorNotFound) {
		return err
	}

	if factor == nil {
		factor = &model.UserFactor{
			UserID:     userID,
			FactorType: model.FactorVerificationCode,
			Material:   verificationCode,
			IsActive:   true,
		}
		if err := s.factorRepo.Create(ctx, factor); err != nil {
			return fmt.Errorf("creating verification factor: %w", err)
		}
	} else {
		factor.Material = verificationCode
		factor.IsActive = true
		factor.VerifiedAt = nil
		if err := s.factorRepo.Update(ctx, factor); err != nil {
			return fmt.Errorf("updating verification factor: %w", err)
		}
	}

	// Generate verification URL
	verificationLink := fmt.Sprintf(
		"%s/api/verify?code=%s&user=%s",
		s.config.BaseURL,
		verificationCode,
		user.ID.String(),
	)

	// Send verification email
	if err := mailer.SendVerificationEmail(s.emailService, user.Email, user.FirstName, verificationLink); err != nil {
		return fmt.Errorf("sending verification email: %w", err)
	}

	return nil
}
