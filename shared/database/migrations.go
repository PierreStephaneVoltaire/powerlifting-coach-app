package database

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/rs/zerolog/log"
)

// MigrationConfig holds configuration for running migrations
type MigrationConfig struct {
	DB             *sql.DB
	MigrationsPath string
	SchemaName     string // Each service gets its own schema
}

// RunMigrations executes database migrations with proper isolation and validation
func RunMigrations(config MigrationConfig) error {
	// Create schema if it doesn't exist
	if err := createSchema(config.DB, config.SchemaName); err != nil {
		return fmt.Errorf("failed to create schema: %w", err)
	}

	// Validate migrations exist
	if err := validateMigrationsExist(config.MigrationsPath); err != nil {
		return fmt.Errorf("migration validation failed: %w", err)
	}

	// Configure postgres driver with schema
	driver, err := postgres.WithInstance(config.DB, &postgres.Config{
		MigrationsTable:       fmt.Sprintf("%s.schema_migrations", config.SchemaName),
		DatabaseName:          "",
		SchemaName:            config.SchemaName,
		MigrationsTableQuoted: false,
	})
	if err != nil {
		return fmt.Errorf("failed to create migration driver: %w", err)
	}

	m, err := migrate.NewWithDatabaseInstance(
		fmt.Sprintf("file://%s", config.MigrationsPath),
		"postgres",
		driver,
	)
	if err != nil {
		return fmt.Errorf("failed to create migration instance: %w", err)
	}

	// Get current version and check for mismatches
	currentVersion, dirty, err := m.Version()
	if err != nil && err != migrate.ErrNilVersion {
		return fmt.Errorf("failed to get current migration version: %w", err)
	}

	// If dirty, log a warning
	if dirty {
		log.Warn().
			Uint("version", currentVersion).
			Str("schema", config.SchemaName).
			Msg("Database is in dirty state, attempting to fix")

		// Force version to clean state
		if err := m.Force(int(currentVersion)); err != nil {
			return fmt.Errorf("failed to force version: %w", err)
		}
	}

	// Count available migrations
	maxAvailableVersion := getMaxMigrationVersion(config.MigrationsPath)

	if err == migrate.ErrNilVersion {
		log.Info().
			Str("schema", config.SchemaName).
			Uint("max_version", maxAvailableVersion).
			Msg("No migrations applied yet, starting fresh")
	} else {
		log.Info().
			Str("schema", config.SchemaName).
			Uint("current", currentVersion).
			Uint("available", maxAvailableVersion).
			Msg("Current migration state")

		// If current version is higher than available migrations, we have a problem
		if currentVersion > maxAvailableVersion {
			log.Warn().
				Str("schema", config.SchemaName).
				Uint("current", currentVersion).
				Uint("available", maxAvailableVersion).
				Msg("Database version is ahead of available migrations - resetting to match available migrations")

			// Force to the max available version
			if err := m.Force(int(maxAvailableVersion)); err != nil {
				return fmt.Errorf("failed to force version to %d: %w", maxAvailableVersion, err)
			}
		}
	}

	// Run migrations
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	finalVersion, _, _ := m.Version()
	log.Info().
		Str("schema", config.SchemaName).
		Uint("version", finalVersion).
		Msg("Database migrations completed successfully")

	return nil
}

// createSchema creates a PostgreSQL schema if it doesn't exist
func createSchema(db *sql.DB, schemaName string) error {
	query := fmt.Sprintf("CREATE SCHEMA IF NOT EXISTS %s", schemaName)
	_, err := db.Exec(query)
	if err != nil {
		return fmt.Errorf("failed to create schema %s: %w", schemaName, err)
	}

	log.Debug().Str("schema", schemaName).Msg("Schema ensured")
	return nil
}

// validateMigrationsExist checks if the migrations directory exists and has files
func validateMigrationsExist(migrationsPath string) error {
	info, err := os.Stat(migrationsPath)
	if err != nil {
		return fmt.Errorf("migrations path does not exist: %s", migrationsPath)
	}

	if !info.IsDir() {
		return fmt.Errorf("migrations path is not a directory: %s", migrationsPath)
	}

	// Check for .sql files
	files, err := filepath.Glob(filepath.Join(migrationsPath, "*.sql"))
	if err != nil {
		return fmt.Errorf("failed to read migrations directory: %w", err)
	}

	if len(files) == 0 {
		return fmt.Errorf("no migration files found in %s", migrationsPath)
	}

	log.Debug().
		Str("path", migrationsPath).
		Int("count", len(files)).
		Msg("Found migration files")

	return nil
}

// getMaxMigrationVersion finds the highest migration version available
func getMaxMigrationVersion(migrationsPath string) uint {
	files, err := filepath.Glob(filepath.Join(migrationsPath, "*.up.sql"))
	if err != nil || len(files) == 0 {
		return 0
	}

	var maxVersion uint
	for _, file := range files {
		var version uint
		_, err := fmt.Sscanf(filepath.Base(file), "%d_", &version)
		if err == nil && version > maxVersion {
			maxVersion = version
		}
	}

	return maxVersion
}
