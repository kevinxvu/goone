package migration

import (
	"database/sql"
	"fmt"

	"github.com/pressly/goose/v3"
	"github.com/vuduongtp/go-core/pkg/logging"
	"go.uber.org/zap"
)

var gooseLogger = logging.Type("goose")

// GooseConfig contains configuration for goose migrations
type GooseConfig struct {
	Dir          string // Directory containing migration files
	TableName    string // Migration table name (default: goose_db_version)
	Dialect      string // Database dialect: mysql, postgres, sqlite3
	Verbose      bool   // Enable verbose logging
	AllowMissing bool   // Allow missing (out-of-order) migrations
}

// DefaultGooseConfig returns default configuration for goose
func DefaultGooseConfig() *GooseConfig {
	return &GooseConfig{
		Dir:          "internal/migrations",
		TableName:    "goose_db_version",
		Dialect:      "mysql",
		Verbose:      false,
		AllowMissing: false,
	}
}

// RunGoose initializes and runs goose migrations
func RunGoose(db *sql.DB, cfg *GooseConfig) error {
	if cfg == nil {
		cfg = DefaultGooseConfig()
	}

	// Set goose configuration
	if cfg.TableName != "" {
		goose.SetTableName(cfg.TableName)
	}
	if cfg.Verbose {
		goose.SetVerbose(true)
	}
	if err := goose.SetDialect(cfg.Dialect); err != nil {
		return fmt.Errorf("failed to set dialect: %w", err)
	}

	// Run migrations
	gooseLogger.Info("Running migrations",
		zap.String("dir", cfg.Dir),
		zap.String("dialect", cfg.Dialect),
		zap.String("table", cfg.TableName),
	)

	if err := goose.Up(db, cfg.Dir); err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	// Get current version
	version, err := goose.GetDBVersion(db)
	if err != nil {
		return fmt.Errorf("failed to get db version: %w", err)
	}

	gooseLogger.Info("Migrations completed successfully",
		zap.Int64("version", version),
	)

	return nil
}

// CreateMigration creates a new migration file
func CreateMigration(db *sql.DB, dir, name, migrationType string) error {
	if migrationType == "" {
		migrationType = "sql"
	}

	if err := goose.Create(db, dir, name, migrationType); err != nil {
		return fmt.Errorf("failed to create migration: %w", err)
	}

	gooseLogger.Info("Migration created successfully",
		zap.String("name", name),
		zap.String("type", migrationType),
		zap.String("dir", dir),
	)

	return nil
}

// GetStatus returns the current migration status
func GetStatus(db *sql.DB, dir, dialect string) error {
	if err := goose.SetDialect(dialect); err != nil {
		return fmt.Errorf("failed to set dialect: %w", err)
	}
	return goose.Status(db, dir)
}

// GetVersion returns the current database version
func GetVersion(db *sql.DB) (int64, error) {
	return goose.GetDBVersion(db)
}

// UpByOne migrates up by a single version
func UpByOne(db *sql.DB, dir, dialect string) error {
	if err := goose.SetDialect(dialect); err != nil {
		return fmt.Errorf("failed to set dialect: %w", err)
	}
	return goose.UpByOne(db, dir)
}

// Down rolls back a single migration
func Down(db *sql.DB, dir, dialect string) error {
	if err := goose.SetDialect(dialect); err != nil {
		return fmt.Errorf("failed to set dialect: %w", err)
	}
	return goose.Down(db, dir)
}

// Reset rolls back all migrations
func Reset(db *sql.DB, dir, dialect string) error {
	if err := goose.SetDialect(dialect); err != nil {
		return fmt.Errorf("failed to set dialect: %w", err)
	}
	return goose.Reset(db, dir)
}

// Redo rolls back the most recent migration, then runs it again
func Redo(db *sql.DB, dir, dialect string) error {
	if err := goose.SetDialect(dialect); err != nil {
		return fmt.Errorf("failed to set dialect: %w", err)
	}
	return goose.Redo(db, dir)
}

// UpTo migrates up to a specific version
func UpTo(db *sql.DB, dir, dialect string, version int64) error {
	if err := goose.SetDialect(dialect); err != nil {
		return fmt.Errorf("failed to set dialect: %w", err)
	}
	return goose.UpTo(db, dir, version)
}

// DownTo rolls back to a specific version
func DownTo(db *sql.DB, dir, dialect string, version int64) error {
	if err := goose.SetDialect(dialect); err != nil {
		return fmt.Errorf("failed to set dialect: %w", err)
	}
	return goose.DownTo(db, dir, version)
}
