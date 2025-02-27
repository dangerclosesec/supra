// cmd/reconcile/main.go
package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/dangerclosesec/supra/internal/auth"
	"github.com/dangerclosesec/supra/internal/config"
	"github.com/dangerclosesec/supra/internal/repository"
	"github.com/dangerclosesec/supra/internal/service"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func main() {
	// Command line flags
	var (
		batchSize = flag.Int("batch-size", 100, "Number of entities to process in a batch")
		dryRun    = flag.Bool("dry-run", false, "Print what would be done without making changes")
		timeout   = flag.Duration("timeout", 30*time.Minute, "Maximum time to run reconciliation")
		entity    = flag.String("entity", "all", "Entity type to reconcile: all, users, organizations")
	)
	flag.Parse()

	// Initialize logger
	logHandler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})
	slogger := slog.New(logHandler)
	slog.SetDefault(slogger)

	// Load configuration
	cfg := config.Load()

	// Initialize database
	db, err := setupDatabase(cfg)
	if err != nil {
		slogger.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}

	// Initialize repositories
	userRepo := repository.NewUserRepository(db)
	orgRepo := repository.NewOrganizationRepository(db)

	// Initialize Supra service
	supraService, err := auth.NewSupraService(
		cfg.Supra.Host,
	)
	if err != nil {
		slogger.Error("failed to initialize Supra service", "error", err)
		os.Exit(1)
	}

	// Initialize entity sync service
	entitySyncService := service.NewEntitySyncService(supraService)

	// Initialize reconciliation service
	reconciliationService := service.NewEntityReconciliationService(
		userRepo,
		orgRepo,
		entitySyncService,
		0, // Interval doesn't matter for one-time sync
		slogger,
	)

	// Configure the reconciliation service
	reconciliationService.SetBatchSize(*batchSize)
	reconciliationService.SetDryRun(*dryRun)

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), *timeout)
	defer cancel()

	// Run reconciliation based on entity flag
	var reconcileErr error

	switch *entity {
	case "all":
		slogger.Info("reconciling all entities")
		reconcileErr = reconciliationService.ReconcileAllEntities(ctx)
	case "users":
		slogger.Info("reconciling users only")
		reconcileErr = reconciliationService.ReconcileUsers(ctx)
	case "organizations":
		slogger.Info("reconciling organizations only")
		reconcileErr = reconciliationService.ReconcileOrganizations(ctx)
	default:
		slogger.Error("unknown entity type", "entity", *entity)
		os.Exit(1)
	}

	if reconcileErr != nil {
		slogger.Error("reconciliation failed", "error", reconcileErr)
		os.Exit(1)
	}

	slogger.Info("reconciliation completed successfully")
}

func setupDatabase(cfg *config.Config) (*gorm.DB, error) {
	connString := os.Getenv("DB_URL")
	if connString == "" {
		connString = "postgres://postgres:password@localhost:5432/identity_graph"
	}

	gormConfig := &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
		NowFunc: func() time.Time {
			return time.Now().UTC()
		},
	}

	db, err := gorm.Open(postgres.Open(connString), gormConfig)
	if err != nil {
		return nil, fmt.Errorf("connecting to database: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("getting database instance: %w", err)
	}

	// Configure connection pool
	sqlDB.SetMaxOpenConns(25)
	sqlDB.SetMaxIdleConns(25)
	sqlDB.SetConnMaxLifetime(5 * time.Minute)
	sqlDB.SetConnMaxIdleTime(5 * time.Minute)

	return db, nil
}
