// internal/repository/organization.go
package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/dangerclosesec/supra/internal/domain"
	"github.com/dangerclosesec/supra/internal/model"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type OrganizationRepository struct {
	db *gorm.DB
}

func NewOrganizationRepository(db *gorm.DB) *OrganizationRepository {
	return &OrganizationRepository{db: db}
}

func (r *OrganizationRepository) Create(ctx context.Context, org *model.Organization) error {
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Check if user already has a personal organization if this is a personal org
		if org.OrgType == model.OrgTypePersonal {
			var count int64
			if err := tx.Model(&model.Organization{}).
				Where("created_by_id = ? AND org_type = ?", org.CreatedByID, model.OrgTypePersonal).
				Count(&count).Error; err != nil {
				return fmt.Errorf("checking existing personal org: %w", err)
			}
			if count > 0 {
				return domain.ErrDuplicatePersonalOrg
			}
		}

		if err := tx.Create(org).Error; err != nil {
			return fmt.Errorf("creating organization: %w", err)
		}

		return nil
	})

	if err != nil {
		if errors.Is(err, domain.ErrDuplicatePersonalOrg) {
			return err
		}
		return fmt.Errorf("transaction failed: %w", err)
	}

	return nil
}

func (r *OrganizationRepository) CreateOrganizationUser(ctx context.Context, orgUser *model.OrganizationUser) error {
	if err := r.db.WithContext(ctx).Create(orgUser).Error; err != nil {
		return fmt.Errorf("creating organization user: %w", err)
	}
	return nil
}

func (r *OrganizationRepository) FindByID(ctx context.Context, id uuid.UUID) (*model.Organization, error) {
	var org model.Organization
	if err := r.db.WithContext(ctx).First(&org, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrOrganizationNotFound
		}
		return nil, fmt.Errorf("finding organization: %w", err)
	}
	return &org, nil
}

func (r *OrganizationRepository) FindByUser(ctx context.Context, userID uuid.UUID) ([]model.Organization, error) {
	var orgs []model.Organization
	if err := r.db.WithContext(ctx).
		Joins("JOIN organization_users ON organizations.id = organization_users.organization_id").
		Where("organization_users.user_id = ?", userID).
		Find(&orgs).Error; err != nil {
		return nil, fmt.Errorf("finding user organizations: %w", err)
	}
	return orgs, nil
}

func (r *OrganizationRepository) Update(ctx context.Context, org *model.Organization) error {
	if err := r.db.WithContext(ctx).Save(org).Error; err != nil {
		return fmt.Errorf("updating organization: %w", err)
	}
	return nil
}

func (r *OrganizationRepository) Delete(ctx context.Context, id uuid.UUID) error {
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Delete organization users first
		if err := tx.Where("organization_id = ?", id).Delete(&model.OrganizationUser{}).Error; err != nil {
			return fmt.Errorf("deleting organization users: %w", err)
		}

		// Delete organization
		if err := tx.Delete(&model.Organization{}, "id = ?", id).Error; err != nil {
			return fmt.Errorf("deleting organization: %w", err)
		}

		return nil
	})

	if err != nil {
		return fmt.Errorf("transaction failed: %w", err)
	}

	return nil
}

// DB returns the underlying database connection
func (r *OrganizationRepository) DB() *gorm.DB {
	return r.db
}
