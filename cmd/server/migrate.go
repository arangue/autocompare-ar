package main

import (
	"errors"
	"log/slog"
	"strings"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/pgx/v5"
	"github.com/golang-migrate/migrate/v4/source/iofs"

	"github.com/arangue/autocompare-ar/migrations"
)

func migrateDatabaseURL(dbURL string) string {
	switch {
	case strings.HasPrefix(dbURL, "postgresql://"):
		return "pgx5://" + strings.TrimPrefix(dbURL, "postgresql://")
	case strings.HasPrefix(dbURL, "postgres://"):
		return "pgx5://" + strings.TrimPrefix(dbURL, "postgres://")
	default:
		return dbURL
	}
}

func runMigrations(dbURL string) error {
	source, err := iofs.New(migrations.FS, ".")
	if err != nil {
		return err
	}
	m, err := migrate.NewWithSourceInstance("iofs", source, migrateDatabaseURL(dbURL))
	if err != nil {
		return err
	}
	defer m.Close() //nolint:errcheck

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return err
	}

	slog.Info("migrations applied")
	return nil
}
