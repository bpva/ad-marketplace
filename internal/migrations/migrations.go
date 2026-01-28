package migrations

import (
	"embed"
	"errors"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
)

//go:embed sql/*.sql
var fs embed.FS

func Run(dbURL string) error {
	source, err := iofs.New(fs, "sql")
	if err != nil {
		return fmt.Errorf("create source: %w", err)
	}

	m, err := migrate.NewWithSourceInstance("iofs", source, dbURL)
	if err != nil {
		return fmt.Errorf("create migrate: %w", err)
	}

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		srcErr, dbErr := m.Close()
		return errors.Join(fmt.Errorf("run migrations: %w", err), srcErr, dbErr)
	}

	srcErr, dbErr := m.Close()
	return errors.Join(srcErr, dbErr)
}
