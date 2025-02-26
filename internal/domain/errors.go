// internal/domain/errors.go
package domain

import "errors"

var (
	// General errors
	ErrNotFound     = errors.New("not found")
	ErrInvalidInput = errors.New("invalid input")

	// Cache-related errors
	ErrInvalidNonce = errors.New("invalid nonce")

	// User-related errors
	ErrUserNotFound        = errors.New("user not found")
	ErrEmailAlreadyExists  = errors.New("email already exists")
	ErrInvalidCredentials  = errors.New("invalid credentials")
	ErrPasswordTooWeak     = errors.New("password too weak")
	ErrUnauthorized        = errors.New("unauthorized")
	ErrPasswordsDoNotMatch = errors.New("passwords do not match")
	ErrInvalidPassword     = errors.New("invalid password")

	// Verification-related errors
	ErrInvalidVerificationCode = errors.New("invalid verification code")
	ErrVerificationExpired     = errors.New("verification code expired")
	ErrAlreadyVerified         = errors.New("already verified")

	// Organization-related errors
	ErrOrganizationNotFound = errors.New("organization not found")
	ErrDuplicatePersonalOrg = errors.New("user can only have one personal organization")
	ErrInvalidOrgType       = errors.New("invalid organization type")

	// Factor-related errors
	ErrFactorNotFound      = errors.New("factor not found")
	ErrInvalidFactorType   = errors.New("invalid factor type")
	ErrFactorAlreadyExists = errors.New("factor already exists")
	ErrInactiveFactor      = errors.New("factor is inactive")
	ErrInvalidFactor       = errors.New("invalid factor")
)
