package postgres

import (
	"errors"
	"fmt"
	"strings"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/pgx/v5"
	"github.com/golang-migrate/migrate/v4/source/iofs"

	"github.com/kazah/golang-task-03/migrations"
)

// RunMigrations applies all up migrations that are embedded into the binary.
func RunMigrations(dsn string) error {
	driver, err := iofs.New(migrations.FS, ".")
	if err != nil {
		return fmt.Errorf("failed to create iofs migration driver: %w", err)
	}

	// migrate pgx driver wants "pgx5://" instead of "postgres://"
	trimmed := strings.TrimPrefix(strings.TrimPrefix(dsn, "postgresql://"), "postgres://")
	migDSN := "pgx5://" + trimmed

	m, err := migrate.NewWithSourceInstance("iofs", driver, migDSN)
	if err != nil {
		return fmt.Errorf("failed to initialize migrator: %w", err)
	}
	defer func() { _, _ = m.Close() }()

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("migration up failed: %w", err)
	}

	return nil
}
