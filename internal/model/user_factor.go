// internal/model/user_factor.go
package model

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type FactorType string

const (
	FactorWebauthn         FactorType = "webauthn"
	FactorPasskey          FactorType = "passkey"
	FactorHashpass         FactorType = "hashpass"
	FactorPubkey           FactorType = "pubkey"
	FactorTOTP             FactorType = "totp"
	FactorOpenID           FactorType = "openid"
	FactorSMS              FactorType = "sms"
	FactorEmail            FactorType = "email"
	FactorU2F              FactorType = "u2f"
	FactorBackupCode       FactorType = "backup_code"
	FactorVerificationCode FactorType = "verification_code"
)

type UserFactor struct {
	ID                      uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	UserID                  uuid.UUID  `gorm:"type:uuid;not null;column:user_id"`
	FactorType              FactorType `gorm:"type:user_factor_type;not null"`
	Material                string     `gorm:"type:text"`
	IsActive                bool       `gorm:"default:true"`
	VerifiedAt              *time.Time
	LastUsedAt              *time.Time
	FederatedAuthProvider   string `gorm:"type:text"`
	FederatedAuthExternalID string `gorm:"type:text"`
	CreatedAt               time.Time
	UpdatedAt               time.Time

	User User `gorm:"foreignKey:UserID"`
}

// BeforeCreate hook for UserFactor
func (uf *UserFactor) BeforeCreate(tx *gorm.DB) error {
	if uf.ID == uuid.Nil {
		uf.ID = uuid.New()
	}

	// Validates FactorType is a known value
	validTypes := map[FactorType]bool{
		FactorWebauthn:         true,
		FactorPasskey:          true,
		FactorHashpass:         true,
		FactorPubkey:           true,
		FactorTOTP:             true,
		FactorOpenID:           true,
		FactorSMS:              true,
		FactorEmail:            true,
		FactorU2F:              true,
		FactorBackupCode:       true,
		FactorVerificationCode: true,
	}

	if !validTypes[uf.FactorType] {
		return fmt.Errorf("invalid factor type: %s", uf.FactorType)
	}

	return nil
}
