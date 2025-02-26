// internal/model/organization.go
package model

import (
	"time"

	"github.com/google/uuid"
)

type OrganizationType string

const (
	OrgTypeEducation  OrganizationType = "education"
	OrgTypeEnterprise OrganizationType = "enterprise"
	OrgTypePersonal   OrganizationType = "personal"
	OrgTypeTeam       OrganizationType = "team"
)

type Organization struct {
	ID          uuid.UUID        `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	Name        string           `gorm:"type:string;not null"`
	OrgType     OrganizationType `gorm:"type:organization_type;not null;default:personal"`
	CreatedByID uuid.UUID        `gorm:"type:uuid;not null"`
	CreatedAt   time.Time
	UpdatedAt   time.Time

	CreatedBy User               `gorm:"foreignKey:CreatedByID"`
	Users     []OrganizationUser `gorm:"foreignKey:OrganizationID"`
}

type OrganizationUser struct {
	ID             uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	OrganizationID uuid.UUID `gorm:"type:uuid;not null"`
	UserID         uuid.UUID `gorm:"type:uuid;not null"`
	Role           string    `gorm:"type:string;not null"`
	CreatedAt      time.Time
	UpdatedAt      time.Time

	Organization Organization `gorm:"foreignKey:OrganizationID"`
	User         User         `gorm:"foreignKey:UserID"`
}
