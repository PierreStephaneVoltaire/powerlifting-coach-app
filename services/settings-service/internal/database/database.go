package database

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/lib/pq"
	shareddb "github.com/PierreStephaneVoltaire/powerlifting-coach-app/shared/database"
)

type DB struct {
	*sql.DB
}

func New(databaseURL string) (*DB, error) {
	db, err := sql.Open("postgres", databaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)
	db.SetConnMaxIdleTime(2 * time.Minute)

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &DB{db}, nil
}

func (db *DB) RunMigrations(migrationsPath string) error {
	return shareddb.RunMigrations(shareddb.MigrationConfig{
		DB:             db.DB,
		MigrationsPath: migrationsPath,
		SchemaName:     "settings_service",
	})
}
