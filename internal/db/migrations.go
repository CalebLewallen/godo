package db

import (
	"embed"
	"fmt"
	"sort"
	"strings"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

// RunMigrations applies all pending up-migrations. It is idempotent.
func (d *DB) RunMigrations() {
	// Ensure the tracking table exists
	if _, err := d.Exec(`
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version  TEXT PRIMARY KEY,
			applied_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
		)
	`); err != nil {
		panic("godo: cannot create schema_migrations: " + err.Error())
	}

	// Load applied versions
	rows, err := d.Query(`SELECT version FROM schema_migrations`)
	if err != nil {
		panic("godo: cannot query migrations: " + err.Error())
	}
	applied := make(map[string]bool)
	for rows.Next() {
		var v string
		if err := rows.Scan(&v); err == nil {
			applied[v] = true
		}
	}
	rows.Close()

	// List migration files
	entries, err := migrationsFS.ReadDir("migrations")
	if err != nil {
		panic("godo: cannot list migrations: " + err.Error())
	}

	// Collect up-migrations in sorted order
	var upFiles []string
	for _, e := range entries {
		if strings.HasSuffix(e.Name(), ".up.sql") {
			upFiles = append(upFiles, e.Name())
		}
	}
	sort.Strings(upFiles)

	for _, fname := range upFiles {
		version := strings.TrimSuffix(fname, ".up.sql")
		if applied[version] {
			continue
		}

		data, err := migrationsFS.ReadFile("migrations/" + fname)
		if err != nil {
			panic(fmt.Sprintf("godo: cannot read migration %s: %v", fname, err))
		}

		tx, err := d.Begin()
		if err != nil {
			panic(fmt.Sprintf("godo: migration %s: begin tx: %v", version, err))
		}

		if _, err := tx.Exec(string(data)); err != nil {
			tx.Rollback()
			panic(fmt.Sprintf("godo: migration %s failed: %v", version, err))
		}

		if _, err := tx.Exec(`INSERT INTO schema_migrations (version) VALUES (?)`, version); err != nil {
			tx.Rollback()
			panic(fmt.Sprintf("godo: migration %s: record failed: %v", version, err))
		}

		if err := tx.Commit(); err != nil {
			panic(fmt.Sprintf("godo: migration %s: commit failed: %v", version, err))
		}
	}
}
