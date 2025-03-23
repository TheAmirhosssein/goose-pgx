package goose

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/multierr"
)

func EnsureDBVersionContextPGX(ctx context.Context, db *pgxpool.Pool) (int64, error) {
	dbMigrations, err := store.ListMigrationsPGX(ctx, db, TableName())
	if err != nil {
		createErr := createVersionTablePGX(ctx, db)
		if createErr != nil {
			return 0, multierr.Append(err, createErr)
		}
		return 0, nil
	}
	// The most recent record for each migration specifies
	// whether it has been applied or rolled back.
	// The first version we find that has been applied is the current version.
	//
	// TODO(mf): for historic reasons, we continue to use the is_applied column,
	// but at some point we need to deprecate this logic and ideally remove
	// this column.
	//
	// For context, see:
	// https://github.com/TheAmirhosssein/goose/pull/131#pullrequestreview-178409168
	//
	// The dbMigrations list is expected to be ordered by descending ID. But
	// in the future we should be able to query the last record only.
	skipLookup := make(map[int64]struct{})
	for _, m := range dbMigrations {
		// Have we already marked this version to be skipped?
		if _, ok := skipLookup[m.VersionID]; ok {
			continue
		}
		// If version has been applied we are done.
		if m.IsApplied {
			return m.VersionID, nil
		}
		// Latest version of migration has not been applied.
		skipLookup[m.VersionID] = struct{}{}
	}
	return 0, ErrNoNextVersion
}

func createVersionTablePGX(ctx context.Context, db *pgxpool.Pool) error {
	txn, err := db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	if err := store.CreateVersionTablePGX(ctx, txn, TableName()); err != nil {
		_ = txn.Rollback(ctx)
		return err
	}
	if err := store.InsertVersionPGX(ctx, txn, TableName(), 0); err != nil {
		_ = txn.Rollback(ctx)
		return err
	}
	return txn.Commit(ctx)
}

// GetDBVersionContext is an alias for EnsureDBVersion, but returns -1 in error.
func GetDBVersionContextPGX(ctx context.Context, db *pgxpool.Pool) (int64, error) {
	version, err := EnsureDBVersionContextPGX(ctx, db)
	if err != nil {
		return -1, err
	}

	return version, nil
}
