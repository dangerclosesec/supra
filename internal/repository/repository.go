// internal/repository/repository.go
package repository

import (
	"log/slog"

	"gorm.io/gorm"
)

// Transaction interface for handling DB transactions.
type Transaction interface {
	Commit() error
	Rollback() error
}

// gormTransaction is a wrapper for a GORM DB transaction.
type gormTransaction struct {
	tx *gorm.DB
}

// Commit finalizes the transaction.
func (t *gormTransaction) Commit() error {
	return t.tx.Commit().Error
}

// Rollback reverts the transaction.
func (t *gormTransaction) Rollback() error {
	slog.Warn("Rolling back transaction")
	return t.tx.Rollback().Error
}
