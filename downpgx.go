package goose

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Down rolls back a single migration from the current version.
func DownPGX(db *pgxpool.Pool, dir string, opts ...OptionsFunc) error {
	ctx := context.Background()
	return DownContextPGX(ctx, db, dir, opts...)
}

func DownContextPGX(ctx context.Context, db *pgxpool.Pool, dir string, opts ...OptionsFunc) error {
	option := &options{}
	for _, f := range opts {
		f(option)
	}
	migrations, err := CollectMigrations(dir, minVersion, maxVersion)
	if err != nil {
		return err
	}
	if option.noVersioning {
		if len(migrations) == 0 {
			return nil
		}
		currentVersion := migrations[len(migrations)-1].Version
		// Migrate only the latest migration down.
		return downToNoVersioningPGX(ctx, db, migrations, currentVersion-1)
	}
	currentVersion, err := GetDBVersionContextPGX(ctx, db)
	if err != nil {
		return err
	}
	current, err := migrations.Current(currentVersion)
	if err != nil {
		return fmt.Errorf("migration %v: %w", currentVersion, err)
	}
	return current.DownContextPGX(ctx, db)
}

// DownToContext rolls back migrations to a specific version.
func DownToContextPGX(ctx context.Context, db *pgxpool.Pool, dir string, version int64, opts ...OptionsFunc) error {
	option := &options{}
	for _, f := range opts {
		f(option)
	}
	migrations, err := CollectMigrations(dir, minVersion, maxVersion)
	if err != nil {
		return err
	}
	if option.noVersioning {
		return downToNoVersioningPGX(ctx, db, migrations, version)
	}

	for {
		currentVersion, err := GetDBVersionContextPGX(ctx, db)
		if err != nil {
			return err
		}

		if currentVersion == 0 {
			log.Printf("goose: no migrations to run. current version: %d", currentVersion)
			return nil
		}
		current, err := migrations.Current(currentVersion)
		if err != nil {
			log.Printf("goose: migration file not found for current version (%d), error: %s", currentVersion, err)
			return err
		}

		if current.Version <= version {
			log.Printf("goose: no migrations to run. current version: %d", currentVersion)
			return nil
		}

		if err = current.DownContextPGX(ctx, db); err != nil {
			return err
		}
	}
}

func downToNoVersioningPGX(ctx context.Context, db *pgxpool.Pool, migrations Migrations, version int64) error {
	var finalVersion int64
	for i := len(migrations) - 1; i >= 0; i-- {
		if version >= migrations[i].Version {
			finalVersion = migrations[i].Version
			break
		}
		migrations[i].noVersioning = true
		if err := migrations[i].DownContextPGX(ctx, db); err != nil {
			return err
		}
	}
	log.Printf("goose: down to current file version: %d", finalVersion)
	return nil
}

func (m *Migration) DownContextPGX(ctx context.Context, db *pgxpool.Pool) error {
	if err := m.runPGX(ctx, db, false); err != nil {
		return err
	}
	return nil
}
