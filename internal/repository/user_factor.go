package repository

import (
	"context"
	"fmt"

	"github.com/dangerclosesec/supra/internal/model"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// UserFactorRepositoryIface defines the interface for the user factor repository.
type UserFactorRepositoryIface interface {
	Begin(ctx context.Context) (Transaction, error)
	Create(ctx context.Context, factor *model.UserFactor) error
	FindByID(ctx context.Context, id uuid.UUID) (*model.UserFactor, error)
	FindByUserID(ctx context.Context, userID uuid.UUID) ([]model.UserFactor, error)
	FindByUserAndType(ctx context.Context, userID uuid.UUID, factorType model.FactorType) (*model.UserFactor, error)
	ListByUser(ctx context.Context, userID uuid.UUID) ([]model.UserFactor, error)
	Update(ctx context.Context, factor *model.UserFactor) error
	Delete(ctx context.Context, factor *model.UserFactor) error
	DeleteByID(ctx context.Context, id uuid.UUID) error
	RemoveFactor(ctx context.Context, userID uuid.UUID, factorID uuid.UUID) error
	FindAllByUser(ctx context.Context, userID uuid.UUID) ([]model.UserFactor, error)
	FindActiveByUser(ctx context.Context, userID uuid.UUID) ([]*model.UserFactor, error)
}

// UserFactorRepository implements UserFactorRepositoryIface.
type UserFactorRepository struct {
	db *gorm.DB
}

// NewUserFactorRepository initializes a new repository instance.
func NewUserFactorRepository(db *gorm.DB) *UserFactorRepository {
	return &UserFactorRepository{db: db}
}

// Begin starts a new transaction.
func (r *UserFactorRepository) Begin(ctx context.Context) (Transaction, error) {
	tx := r.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return nil, tx.Error
	}
	return &gormTransaction{tx: tx}, nil
}

// Create inserts a new user factor.
func (r *UserFactorRepository) Create(ctx context.Context, factor *model.UserFactor) error {
	if err := r.db.WithContext(ctx).Create(factor).Error; err != nil {
		return fmt.Errorf("creating user factor: %w", err)
	}
	return nil
}

// FindByID retrieves a user factor by ID.
func (r *UserFactorRepository) FindByID(ctx context.Context, id uuid.UUID) (*model.UserFactor, error) {
	var factor model.UserFactor
	if err := r.db.WithContext(ctx).First(&factor, "id = ?", id).Error; err != nil {
		return nil, fmt.Errorf("finding user factor: %w", err)
	}
	return &factor, nil
}

// FindByUserID retrieves all user factors for a user.
func (r *UserFactorRepository) FindByUserID(ctx context.Context, userID uuid.UUID) ([]model.UserFactor, error) {
	var factors []model.UserFactor
	if err := r.db.WithContext(ctx).Find(&factors, "user_id = ?", userID).Error; err != nil {
		return nil, fmt.Errorf("finding user factors: %w", err)
	}
	return factors, nil
}

// FindByUserAndType retrieves a specific user factor by type.
func (r *UserFactorRepository) FindByUserAndType(ctx context.Context, userID uuid.UUID, factorType model.FactorType) (*model.UserFactor, error) {
	var factor model.UserFactor
	if err := r.db.WithContext(ctx).First(&factor, "user_id = ? AND factor_type = ?", userID, factorType).Error; err != nil {
		return nil, fmt.Errorf("finding user factor: %w", err)
	}
	return &factor, nil
}

// ListByUser retrieves all user factors for a user.
func (r *UserFactorRepository) ListByUser(ctx context.Context, userID uuid.UUID) ([]model.UserFactor, error) {
	var factors []model.UserFactor
	if err := r.db.WithContext(ctx).Find(&factors, "user_id = ?", userID).Error; err != nil {
		return nil, fmt.Errorf("finding user factors: %w", err)
	}
	return factors, nil
}

// Update modifies an existing user factor.
func (r *UserFactorRepository) Update(ctx context.Context, factor *model.UserFactor) error {
	if err := r.db.WithContext(ctx).Save(factor).Error; err != nil {
		return fmt.Errorf("updating user factor: %w", err)
	}
	return nil
}

// Delete removes a user factor.
func (r *UserFactorRepository) Delete(ctx context.Context, factor *model.UserFactor) error {
	if err := r.db.WithContext(ctx).Delete(factor).Error; err != nil {
		return fmt.Errorf("deleting user factor: %w", err)
	}
	return nil
}

// DeleteByID removes a user factor by ID.
func (r *UserFactorRepository) DeleteByID(ctx context.Context, id uuid.UUID) error {
	if err := r.db.WithContext(ctx).Delete(&model.UserFactor{}, "id = ?", id).Error; err != nil {
		return fmt.Errorf("deleting user factor: %w", err)
	}
	return nil
}

// RemoveFactor removes a specific user factor by user and factor ID.
func (r *UserFactorRepository) RemoveFactor(ctx context.Context, userID uuid.UUID, factorID uuid.UUID) error {
	if err := r.db.WithContext(ctx).Delete(&model.UserFactor{}, "user_id = ? AND id = ?", userID, factorID).Error; err != nil {
		return fmt.Errorf("removing user factor: %w", err)
	}
	return nil
}

// FindAllByUser retrieves all user factors for a given user.
func (r *UserFactorRepository) FindAllByUser(ctx context.Context, userID uuid.UUID) ([]model.UserFactor, error) {
	var factors []model.UserFactor
	if err := r.db.WithContext(ctx).Find(&factors, "user_id = ?", userID).Error; err != nil {
		return nil, fmt.Errorf("finding user factors: %w", err)
	}
	return factors, nil
}

// FindActiveByUser retrieves active user factors for a given user.
func (r *UserFactorRepository) FindActiveByUser(ctx context.Context, userID uuid.UUID) ([]*model.UserFactor, error) {
	var factors []*model.UserFactor
	if err := r.db.WithContext(ctx).Find(&factors, "user_id = ? AND is_active = ?", userID, true).Error; err != nil {
		return nil, fmt.Errorf("finding active user factors: %w", err)
	}
	return factors, nil
}
