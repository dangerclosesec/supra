package supra

import (
	"context"
	"database/sql"
	"log/slog"
)

// Config holds the configuration settings for the application.
type Config struct {
	// ctx is the context for all operations.
	ctx context.Context

	// logger is the logger used for logging messages.
	logger *slog.Logger

	// TablePrefix is the prefix for all tables.
	// Default is "supra_".
	TablePrefix string

	// db is the database connection used for migrations.
	db *sql.DB
}

func NewConfig(ctx context.Context, db *sql.DB) *Config {
	return &Config{
		ctx:         ctx,
		TablePrefix: "supra_",
	}
}

// SetTablePrefix sets the table prefix for all tables.
func (c *Config) SetTablePrefix(prefix string) {
	c.TablePrefix = prefix
}

// SetDB sets the database connection.
func (c *Config) SetDB(db *sql.DB) {
	c.db = db
}

// SetLogger sets the logger.
func (c *Config) SetLogger(logger *slog.Logger) {
	c.logger = logger
}
