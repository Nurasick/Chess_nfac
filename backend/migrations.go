package main

import (
	"fmt"
	"log"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func runMigrations(databaseURL string) error {
	m, err := migrate.New("file://db/migrations", databaseURL)
	if err != nil {
		return fmt.Errorf("migration init failed: %w", err)
	}
	defer m.Close()

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("migration up failed: %w", err)
	}

	version, dirty, err := m.Version()
	if err == nil {
		log.Printf("✓ Migrations applied (version: %d, dirty: %v)\n", version, dirty)
	}
	return nil
}
