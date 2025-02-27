// internal/service/entity_reconciliation.go
package service

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/dangerclosesec/supra/internal/repository"
)

// EntityReconciliationService periodically reconciles database entities with permission system
type EntityReconciliationService struct {
	userRepo     repository.UserRepositoryIface
	orgRepo      *repository.OrganizationRepository
	entitySync   *EntitySyncService
	syncInterval time.Duration
	batchSize    int
	dryRun       bool // If true, don't make changes, just log
	logger       *slog.Logger
	stopChan     chan struct{}
	stoppedChan  chan struct{}
}

// NewEntityReconciliationService creates a new reconciliation service
func NewEntityReconciliationService(
	userRepo repository.UserRepositoryIface,
	orgRepo *repository.OrganizationRepository,
	entitySync *EntitySyncService,
	syncInterval time.Duration,
	logger *slog.Logger,
) *EntityReconciliationService {
	if syncInterval == 0 {
		syncInterval = 30 * time.Minute
	}

	return &EntityReconciliationService{
		userRepo:     userRepo,
		orgRepo:      orgRepo,
		entitySync:   entitySync,
		syncInterval: syncInterval,
		batchSize:    100,
		dryRun:       false,
		logger:       logger,
		stopChan:     make(chan struct{}),
		stoppedChan:  make(chan struct{}),
	}
}

// Start begins the periodic reconciliation process
func (s *EntityReconciliationService) Start() {
	go func() {
		ticker := time.NewTicker(s.syncInterval)
		defer ticker.Stop()
		defer close(s.stoppedChan)

		for {
			select {
			case <-ticker.C:
				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
				if err := s.reconcileAll(ctx); err != nil {
					s.logger.Error("reconciliation failed", "error", err)
				}
				cancel()
			case <-s.stopChan:
				return
			}
		}
	}()
}

// Stop halts the reconciliation process
func (s *EntityReconciliationService) Stop() {
	close(s.stopChan)
	<-s.stoppedChan
}

// reconcileAll reconciles all entities
func (s *EntityReconciliationService) reconcileAll(ctx context.Context) error {
	s.logger.Info("starting full reconciliation of all entities")

	// Reconcile users
	if err := s.ReconcileUsers(ctx); err != nil {
		return fmt.Errorf("reconciling users: %w", err)
	}

	// Reconcile organizations
	if err := s.ReconcileOrganizations(ctx); err != nil {
		return fmt.Errorf("reconciling organizations: %w", err)
	}

	// We could add other entity types here
	s.logger.Info("completed full reconciliation of all entities")

	return nil
}

// ReconcileAllEntities runs a one-time reconciliation of all entities
// This can be called on-demand to sync existing data
func (s *EntityReconciliationService) ReconcileAllEntities(ctx context.Context) error {
	return s.reconcileAll(ctx)
}

// SetBatchSize sets the number of entities to process in a batch
func (s *EntityReconciliationService) SetBatchSize(size int) {
	if size > 0 {
		s.batchSize = size
	}
}

// SetDryRun sets whether to actually make changes or just log what would be done
func (s *EntityReconciliationService) SetDryRun(dryRun bool) {
	s.dryRun = dryRun
}

// ReconcileUsers reconciles all users with the permission system
func (s *EntityReconciliationService) ReconcileUsers(ctx context.Context) error {
	// In a real implementation, we'd use pagination to handle large datasets
	// For simplicity, we'll fetch all users
	users, err := s.userRepo.FindAll(ctx)
	if err != nil {
		return fmt.Errorf("fetching users: %w", err)
	}

	s.logger.Info("reconciling users", "count", len(users), "dry_run", s.dryRun)

	// Process in batches
	for i := 0; i < len(users); i += s.batchSize {
		end := i + s.batchSize
		if end > len(users) {
			end = len(users)
		}

		batch := users[i:end]
		s.logger.Info("processing user batch", "start", i, "end", end, "size", len(batch))

		for _, user := range batch {
			if s.dryRun {
				s.logger.Info("would sync user (dry run)",
					"user_id", user.ID.String(),
					"email", user.Email,
					"status", user.Status,
				)
				continue
			}

			if err := s.entitySync.SyncUserToPermissions(ctx, user); err != nil {
				s.logger.Error("failed to sync user",
					"user_id", user.ID.String(),
					"error", err,
				)
				// Continue with other users
			} else {
				s.logger.Info("successfully synced user",
					"user_id", user.ID.String(),
					"email", user.Email,
				)
			}
		}

		// Check if context is done between batches
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			// Continue processing
		}
	}

	return nil
}

// ReconcileOrganizations reconciles all organizations with the permission system
func (s *EntityReconciliationService) ReconcileOrganizations(ctx context.Context) error {
	// In a real implementation, we'd use pagination to handle large datasets
	// For simplicity, we'll fetch all organizations
	orgs, err := s.orgRepo.FindAll(ctx)
	if err != nil {
		return fmt.Errorf("fetching organizations: %w", err)
	}

	s.logger.Info("reconciling organizations", "count", len(orgs), "dry_run", s.dryRun)

	// Process in batches
	for i := 0; i < len(orgs); i += s.batchSize {
		end := i + s.batchSize
		if end > len(orgs) {
			end = len(orgs)
		}

		batch := orgs[i:end]
		s.logger.Info("processing organization batch", "start", i, "end", end, "size", len(batch))

		for _, org := range batch {
			if s.dryRun {
				s.logger.Info("would sync organization (dry run)",
					"org_id", org.ID.String(),
					"name", org.Name,
					"type", org.OrgType,
				)

				// Log members that would be synced
				members, err := s.orgRepo.FindOrganizationUsers(ctx, org.ID)
				if err == nil {
					for _, member := range members {
						s.logger.Info("would sync relationship (dry run)",
							"org_id", org.ID.String(),
							"user_id", member.UserID.String(),
							"role", member.Role,
						)
					}
				}

				continue
			}

			// Sync organization entity
			if err := s.entitySync.SyncOrganizationToPermissions(ctx, org); err != nil {
				s.logger.Error("failed to sync organization",
					"org_id", org.ID.String(),
					"error", err,
				)
				// Continue with other organizations
			} else {
				s.logger.Info("successfully synced organization",
					"org_id", org.ID.String(),
					"name", org.Name,
				)
			}

			// Now reconcile membership relationships
			members, err := s.orgRepo.FindOrganizationUsers(ctx, org.ID)
			if err != nil {
				s.logger.Error("failed to fetch organization members",
					"org_id", org.ID.String(),
					"error", err,
				)
				continue
			}

			for _, member := range members {
				if err := s.entitySync.EstablishUserOrganizationRelation(ctx, org.ID, member.UserID, member.Role); err != nil {
					s.logger.Error("failed to sync member relationship",
						"org_id", org.ID.String(),
						"user_id", member.UserID.String(),
						"role", member.Role,
						"error", err,
					)
				} else {
					s.logger.Info("successfully synced relationship",
						"org_id", org.ID.String(),
						"user_id", member.UserID.String(),
						"role", member.Role,
					)
				}
			}
		}

		// Check if context is done between batches
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			// Continue processing
		}
	}

	return nil
}
