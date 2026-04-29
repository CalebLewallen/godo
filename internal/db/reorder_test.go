package db

import (
	"database/sql"
	"os"
	"path/filepath"
	"testing"
)

func TestSwapSiblings(t *testing.T) {
	// Create a temporary database for testing
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test.db")

	sqlDB, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer sqlDB.Close()
	defer os.Remove(dbPath)

	db := &DB{sqlDB}

	// Create tables
	// RunMigrations will panic on error
	db.RunMigrations()

	// Insert some sample data
	_, err = db.Exec(`INSERT INTO folders (id, name, parent_id, position) VALUES (1, 'Folder 1', NULL, 0)`)
	if err != nil {
		t.Fatalf("Failed to insert folder 1: %v", err)
	}

	_, err = db.Exec(`INSERT INTO folders (id, name, parent_id, position) VALUES (2, 'Folder 2', NULL, 1)`)
	if err != nil {
		t.Fatalf("Failed to insert folder 2: %v", err)
	}

	// Swap Folder 1 and Folder 2
	err = db.SwapSiblings(nil, true, 1, true, 2)
	if err != nil {
		t.Fatalf("SwapSiblings failed: %v", err)
	}

	// Verify the updated positions
	var pos1, pos2 int
	err = db.QueryRow(`SELECT position FROM folders WHERE id = 1`).Scan(&pos1)
	if err != nil {
		t.Fatalf("Failed to query folder 1 position: %v", err)
	}
	err = db.QueryRow(`SELECT position FROM folders WHERE id = 2`).Scan(&pos2)
	if err != nil {
		t.Fatalf("Failed to query folder 2 position: %v", err)
	}

	if pos1 != 1 {
		t.Errorf("Expected Folder 1 position to be 1, got %d", pos1)
	}
	if pos2 != 0 {
		t.Errorf("Expected Folder 2 position to be 0, got %d", pos2)
	}
}
