package database

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"sort"
	"strings"
)

//go:embed migrations/*.sql fixtures/*.sql
var scripts embed.FS

// Setup applies each versioned script once. Service restarts do not reapply fixtures.
func Setup(ctx context.Context, db *sql.DB) error {
	if _, err := db.ExecContext(ctx, `CREATE TABLE IF NOT EXISTS setup_versions (name text PRIMARY KEY, applied_at timestamptz NOT NULL DEFAULT now())`); err != nil {
		return err
	}
	files := []string{}
	for _, directory := range []string{"migrations", "fixtures"} {
		entries, err := scripts.ReadDir(directory)
		if err != nil {
			return err
		}
		for _, entry := range entries {
			if strings.HasSuffix(entry.Name(), ".sql") {
				files = append(files, directory+"/"+entry.Name())
			}
		}
	}
	sort.Strings(files)
	// Migrations must precede fixtures, regardless of lexical directory order.
	sort.SliceStable(files, func(i, j int) bool {
		return strings.HasPrefix(files[i], "migrations/") && strings.HasPrefix(files[j], "fixtures/")
	})
	for _, name := range files {
		if err := apply(ctx, db, name); err != nil {
			return fmt.Errorf("%s: %w", name, err)
		}
	}
	return nil
}

func apply(ctx context.Context, db *sql.DB, name string) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	// Lock the ledger so parallel setup invocations cannot both seed the same database.
	if _, err := tx.ExecContext(ctx, `LOCK TABLE setup_versions IN EXCLUSIVE MODE`); err != nil {
		return err
	}
	var applied bool
	if err := tx.QueryRowContext(ctx, `SELECT EXISTS (SELECT 1 FROM setup_versions WHERE name = $1)`, name).Scan(&applied); err != nil {
		return err
	}
	if applied {
		return tx.Commit()
	}
	contents, err := scripts.ReadFile(name)
	if err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, string(contents)); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `INSERT INTO setup_versions (name) VALUES ($1)`, name); err != nil {
		return err
	}
	return tx.Commit()
}
