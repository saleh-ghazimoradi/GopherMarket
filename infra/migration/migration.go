package migration

import (
	"database/sql"
	"embed"
	"errors"
	"fmt"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
)

//go:embed *.sql
var fs embed.FS

type Migrator struct {
	m *migrate.Migrate
}

func (m *Migrator) Up() error {
	if err := m.m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("migrate up: %w", err)
	}
	return nil
}

func (m *Migrator) Down() error {
	if err := m.m.Down(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("migrate down: %w", err)
	}
	return nil
}

func (m *Migrator) Steps(n int) error {
	if err := m.m.Steps(n); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("migration steps: %w", err)
	}
	return nil
}

func (m *Migrator) Close() error {
	if m.m == nil {
		return nil
	}

	sourceErr, driverErr := m.m.Close()
	if sourceErr != nil {
		return fmt.Errorf("close migration source: %w", sourceErr)
	}

	if driverErr != nil {
		return fmt.Errorf("close migration driver: %w", driverErr)
	}
	return nil
}

func NewMigrator(db *sql.DB, dbName string) (*Migrator, error) {
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return nil, fmt.Errorf("create postgres driver: %w", err)
	}

	source, err := iofs.New(fs, ".")
	if err != nil {
		return nil, fmt.Errorf("load migration files: %w", err)
	}

	m, err := migrate.NewWithInstance("iofs", source, dbName, driver)
	if err != nil {
		return nil, fmt.Errorf("init migration instance: %w", err)
	}

	return &Migrator{m: m}, nil
}
