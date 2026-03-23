package db

import (
	"database/sql"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

// DB wraps sql.DB with app-specific methods.
type DB struct {
	*sql.DB
}

// Open creates the data directory if needed, opens the SQLite DB, and sets WAL mode.
func Open() *DB {
	dir := dataDir()
	if err := os.MkdirAll(dir, 0755); err != nil {
		panic("godo: cannot create data dir: " + err.Error())
	}

	path := filepath.Join(dir, "godo.db")
	sqlDB, err := sql.Open("sqlite", path)
	if err != nil {
		panic("godo: cannot open db: " + err.Error())
	}

	if _, err := sqlDB.Exec("PRAGMA journal_mode=WAL"); err != nil {
		panic("godo: cannot set WAL mode: " + err.Error())
	}
	if _, err := sqlDB.Exec("PRAGMA foreign_keys=ON"); err != nil {
		panic("godo: cannot enable foreign keys: " + err.Error())
	}

	return &DB{sqlDB}
}

// DataDir returns the path to the godo data directory.
func dataDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		panic("godo: cannot determine home dir: " + err.Error())
	}
	return filepath.Join(home, ".local", "share", "godo")
}

// DataDirPath is exported for uninstall use.
func DataDirPath() string {
	return dataDir()
}
