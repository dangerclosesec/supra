// internal/repository/user.go
package repository

import (
	"context"
	"fmt"

	"github.com/dangerclosesec/supra/internal/domain"
	"github.com/dangerclosesec/supra/internal/model"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type UserRepositoryIface interface {
	Begin(ctx context.Context) (Transaction, error) // For mocking support in tests

	Create(ctx context.Context, user *model.User) error
	FindByEmail(ctx context.Context, email string) (*model.User, error)
	FindByID(ctx context.Context, id uuid.UUID) (*model.User, error)
	Update(ctx context.Context, user *model.User) error
	Delete(ctx context.Context, id uuid.UUID) error
	FindAll(ctx context.Context) ([]*model.User, error)                                    // Get all users
	FindAllPaginated(ctx context.Context, offset, limit int) ([]*model.User, int64, error) // Get users with pagination
}

type UserRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{db: db}
}

// // DB returns the underlying database instance.
// func (r *UserRepository) DB() *gorm.DB {
// 	return r.db
// }

// Begin starts a new database transaction and returns a Transaction instance.
func (r *UserRepository) Begin(ctx context.Context) (Transaction, error) {
	tx := r.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return nil, tx.Error
	}
	return &gormTransaction{tx: tx}, nil
}

func (r *UserRepository) Create(ctx context.Context, user *model.User) error {
	result := r.db.WithContext(ctx).Create(user)
	if result.Error != nil {
		return fmt.Errorf("failed to create user: %w", result.Error)
	}
	return nil
}

func (r *UserRepository) FindByEmail(ctx context.Context, email string) (*model.User, error) {
	var user model.User
	result := r.db.WithContext(ctx).Where("email = ?", email).First(&user)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, domain.ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to find user: %w", result.Error)
	}

	fmt.Printf("User: %+v\n", user)
	return &user, nil
}

func (r *UserRepository) FindByID(ctx context.Context, id uuid.UUID) (*model.User, error) {
	var user model.User
	result := r.db.WithContext(ctx).First(&user, "id = ?", id)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, domain.ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to find user: %w", result.Error)
	}
	return &user, nil
}

func (r *UserRepository) Update(ctx context.Context, user *model.User) error {
	result := r.db.WithContext(ctx).Save(user)
	if result.Error != nil {
		return fmt.Errorf("failed to update user: %w", result.Error)
	}
	return nil
}

func (r *UserRepository) Delete(ctx context.Context, id uuid.UUID) error {
	result := r.db.WithContext(ctx).Delete(&model.User{}, id)
	if result.Error != nil {
		return fmt.Errorf("failed to delete user: %w", result.Error)
	}
	return nil
}

func (r *UserRepository) FindByOrganization(ctx context.Context, orgID uuid.UUID) ([]model.User, error) {
	var users []model.User
	result := r.db.WithContext(ctx).
		Joins("JOIN organization_users ON users.id = organization_users.user_id").
		Where("organization_users.organization_id = ?", orgID).
		Find(&users)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to find users: %w", result.Error)
	}
	return users, nil
}

// FindAll returns all users
func (r *UserRepository) FindAll(ctx context.Context) ([]*model.User, error) {
	var users []*model.User
	result := r.db.WithContext(ctx).Find(&users)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to find all users: %w", result.Error)
	}
	return users, nil
}

// FindAllPaginated returns a paginated list of users
func (r *UserRepository) FindAllPaginated(ctx context.Context, offset, limit int) ([]*model.User, int64, error) {
	var users []*model.User
	var count int64

	// Get total count
	if err := r.db.WithContext(ctx).Model(&model.User{}).Count(&count).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count users: %w", err)
	}

	// Get paginated users
	result := r.db.WithContext(ctx).Offset(offset).Limit(limit).Find(&users)
	if result.Error != nil {
		return nil, 0, fmt.Errorf("failed to find paginated users: %w", result.Error)
	}

	return users, count, nil
}
