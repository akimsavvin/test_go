package storage

import (
	"database/sql"
	"errors"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/pgx"
)

func Migrate(db *sql.DB) error {
	driver, err := pgx.WithInstance(db, &pgx.Config{
		MigrationsTable: "migrations",
	})
	if err != nil {
		return err
	}

	m, err := migrate.NewWithDatabaseInstance("file://migrations", "pgx", driver)
	if err != nil {
		return err
	}

	if err = m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return err
	}

	return nil
}
