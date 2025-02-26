// internal/model/user.go
package model

import (
	"database/sql/driver"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

type UserStatus string

const (
	StatusPending   UserStatus = "pending"
	StatusActive    UserStatus = "active"
	StatusLocked    UserStatus = "locked"
	StatusSuspended UserStatus = "suspended"
)

type User struct {
	ID               uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	Email            string     `gorm:"type:citext;uniqueIndex;not null" json:"email"`
	FirstName        string     `gorm:"type:text;not null" json:"first_name"`
	LastName         string     `gorm:"type:text" json:"last_name"`
	Status           UserStatus `gorm:"type:user_status;not null;default:'pending'" json:"status"`
	NotificationType string     `gorm:"type:text;not null;default:'email'" json:"notification_type"`
	Experience       Experience `gorm:"type:text[];not null;default:'{\"other\"}'" json:"experience"`
	Theme            string     `gorm:"type:text;not null;default:'light'" json:"theme"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
}

// Experience is a custom type that implements the sql.Scanner and driver.Valuer interfaces
type Experience []string

// Scan implements the sql.Scanner interface
func (e *Experience) Scan(value interface{}) error {
	if value == nil {
		*e = []string{}
		return nil
	}

	str, ok := value.(string)
	if !ok {
		return fmt.Errorf("unsupported Scan, storing driver.Value type %T into type %T", value, e)
	}

	// Remove curly braces and split by comma
	*e = strings.Split(strings.Trim(str, "{}"), ",")
	return nil
}

// Value implements the driver.Valuer interface
func (e Experience) Value() (driver.Value, error) {
	if len(e) == 0 {
		return "{}", nil
	}

	// Join the slice into a string and wrap with curly braces
	return "{" + strings.Join(e, ",") + "}", nil
}
