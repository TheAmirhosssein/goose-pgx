package goose

import (
	"context"
	"fmt"
	"path/filepath"
	"time"

	"github.com/TheAmirhosssein/goose/v3/internal/sqlparser"
	"github.com/jackc/pgx/v5"
)

// UpContext runs an up migration.
func (m *Migration) UpContextPGX(ctx context.Context, db *pgx.Conn) error {
	if err := m.runPGX(ctx, db, true); err != nil {
		return err
	}
	return nil
}

func (m *Migration) runPGX(ctx context.Context, db *pgx.Conn, direction bool) error {
	switch filepath.Ext(m.Source) {
	case ".sql":
		f, err := baseFS.Open(m.Source)
		if err != nil {
			return fmt.Errorf("ERROR %v: failed to open SQL migration file: %w", filepath.Base(m.Source), err)
		}
		defer f.Close()

		statements, useTx, err := sqlparser.ParseSQLMigration(f, sqlparser.FromBool(direction), verbose)
		if err != nil {
			return fmt.Errorf("ERROR %v: failed to parse SQL migration file: %w", filepath.Base(m.Source), err)
		}

		start := time.Now()
		if err := runSQLMigrationPGX(ctx, db, statements, useTx, m.Version, direction, m.noVersioning); err != nil {
			return fmt.Errorf("ERROR %v: failed to run SQL migration: %w", filepath.Base(m.Source), err)
		}
		finish := truncateDuration(time.Since(start))

		if len(statements) > 0 {
			log.Printf("OK   %s (%s)", filepath.Base(m.Source), finish)
		} else {
			log.Printf("EMPTY %s (%s)", filepath.Base(m.Source), finish)
		}

	case ".go":
		panic("we don't have this feature")
	}
	return nil
}

func runSQLMigrationPGX(
	ctx context.Context,
	db *pgx.Conn,
	statements []string,
	useTx bool,
	v int64,
	direction bool,
	noVersioning bool,
) error {
	if useTx {
		// TRANSACTION.

		verboseInfo("Begin transaction")

		tx, err := db.BeginTx(ctx, pgx.TxOptions{})
		if err != nil {
			return fmt.Errorf("failed to begin transaction: %w", err)
		}

		for _, query := range statements {
			verboseInfo("Executing statement: %s\n", clearStatement(query))
			if _, err := tx.Exec(ctx, query); err != nil {
				verboseInfo("Rollback transaction")
				_ = tx.Rollback(ctx)
				return fmt.Errorf("failed to execute SQL query %q: %w", clearStatement(query), err)
			}
		}

		if !noVersioning {
			if direction {
				if err := store.InsertVersionPGX(ctx, tx, TableName(), v); err != nil {
					verboseInfo("Rollback transaction")
					_ = tx.Rollback(ctx)
					return fmt.Errorf("failed to insert new goose version: %w", err)
				}
			} else {
				if err := store.DeleteVersionPGX(ctx, tx, TableName(), v); err != nil {
					verboseInfo("Rollback transaction")
					_ = tx.Rollback(ctx)
					return fmt.Errorf("failed to delete goose version: %w", err)
				}
			}
		}

		verboseInfo("Commit transaction")
		if err := tx.Commit(ctx); err != nil {
			return fmt.Errorf("failed to commit transaction: %w", err)
		}

		return nil
	}

	// NO TRANSACTION.
	for _, query := range statements {
		verboseInfo("Executing statement: %s", clearStatement(query))
		if _, err := db.Exec(ctx, query); err != nil {
			return fmt.Errorf("failed to execute SQL query %q: %w", clearStatement(query), err)
		}
	}
	if !noVersioning {
		if direction {
			if err := store.InsertVersionNoTxPGX(ctx, db, TableName(), v); err != nil {
				return fmt.Errorf("failed to insert new goose version: %w", err)
			}
		} else {
			if err := store.DeleteVersionNoTxPGX(ctx, db, TableName(), v); err != nil {
				return fmt.Errorf("failed to delete goose version: %w", err)
			}
		}
	}

	return nil
}
