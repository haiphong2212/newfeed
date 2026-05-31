package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/newfeed/community-news/services/shared/platform"
)

func main() {
	ctx := context.Background()
	cfg := platform.Load("migrate", ":0", ":0")
	db, err := platform.NewPostgres(ctx, cfg.PostgresDSN)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	root := os.Getenv("MIGRATIONS_DIR")
	if root == "" {
		root = "database/migrations"
	}
	if err := ensureMigrationTable(ctx, db); err != nil {
		log.Fatal(err)
	}
	entries, err := os.ReadDir(root)
	if err != nil {
		log.Fatal(err)
	}
	sort.Slice(entries, func(i, j int) bool { return entries[i].Name() < entries[j].Name() })
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		version := entry.Name()
		applied, err := migrationApplied(ctx, db, version)
		if err != nil {
			log.Fatal(err)
		}
		if applied {
			continue
		}
		if err := applyMigration(ctx, db, root, version); err != nil {
			log.Fatal(err)
		}
		log.Printf("applied migration %s", version)
	}
}

type execer interface {
	Exec(context.Context, string, ...any) (pgconn.CommandTag, error)
	QueryRow(context.Context, string, ...any) pgx.Row
}

func ensureMigrationTable(ctx context.Context, db execer) error {
	_, err := db.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version TEXT PRIMARY KEY,
			applied_at TIMESTAMPTZ NOT NULL DEFAULT now()
		)
	`)
	return err
}

func migrationApplied(ctx context.Context, db execer, version string) (bool, error) {
	var existing string
	err := db.QueryRow(ctx, `SELECT version FROM schema_migrations WHERE version = $1`, version).Scan(&existing)
	if errors.Is(err, pgx.ErrNoRows) {
		return false, nil
	}
	return err == nil, err
}

func applyMigration(ctx context.Context, db *pgxpool.Pool, root, version string) error {
	dir := filepath.Join(root, version)
	files, err := os.ReadDir(dir)
	if err != nil {
		return err
	}
	sort.Slice(files, func(i, j int) bool { return files[i].Name() < files[j].Name() })
	tx, err := db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	for _, file := range files {
		if file.IsDir() || !strings.HasSuffix(file.Name(), ".sql") {
			continue
		}
		data, err := os.ReadFile(filepath.Join(dir, file.Name()))
		if err != nil {
			return err
		}
		if strings.TrimSpace(string(data)) == "" {
			continue
		}
		if _, err := tx.Exec(ctx, string(data)); err != nil {
			return fmt.Errorf("%s/%s: %w", version, file.Name(), err)
		}
	}
	if _, err := tx.Exec(ctx, `INSERT INTO schema_migrations(version) VALUES($1)`, version); err != nil {
		return err
	}
	return tx.Commit(ctx)
}
