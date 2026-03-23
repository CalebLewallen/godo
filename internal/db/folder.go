package db

import (
	"database/sql"
	"time"

	"github.com/CalebLewallen/godo/internal/model"
)

// GetAllFolders returns all folders ordered by parent_id, position.
func (d *DB) GetAllFolders() ([]model.Folder, error) {
	rows, err := d.Query(`
		SELECT id, name, parent_id, position, created_at, updated_at
		FROM folders
		ORDER BY parent_id, position, name
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var folders []model.Folder
	for rows.Next() {
		var f model.Folder
		var parentID sql.NullInt64
		if err := rows.Scan(&f.ID, &f.Name, &parentID, &f.Position, &f.CreatedAt, &f.UpdatedAt); err != nil {
			return nil, err
		}
		if parentID.Valid {
			id := parentID.Int64
			f.ParentID = &id
		}
		folders = append(folders, f)
	}
	return folders, rows.Err()
}

// CreateFolder inserts a new folder and returns it with the assigned ID.
func (d *DB) CreateFolder(name string, parentID *int64) (model.Folder, error) {
	now := time.Now().UTC()
	res, err := d.Exec(`
		INSERT INTO folders (name, parent_id, created_at, updated_at)
		VALUES (?, ?, ?, ?)
	`, name, parentID, now, now)
	if err != nil {
		return model.Folder{}, err
	}
	id, _ := res.LastInsertId()
	return model.Folder{
		ID:        id,
		Name:      name,
		ParentID:  parentID,
		CreatedAt: now,
		UpdatedAt: now,
	}, nil
}

// UpdateFolder updates a folder's name and parent.
func (d *DB) UpdateFolder(f model.Folder) error {
	now := time.Now().UTC()
	_, err := d.Exec(`
		UPDATE folders SET name=?, parent_id=?, position=?, updated_at=? WHERE id=?
	`, f.Name, f.ParentID, f.Position, now, f.ID)
	return err
}

// RenameFolder updates a folder's name.
func (d *DB) RenameFolder(id int64, name string) error {
	_, err := d.Exec(`UPDATE folders SET name=?, updated_at=CURRENT_TIMESTAMP WHERE id=?`, name, id)
	return err
}

// DeleteFolder deletes a folder by ID (cascades to children).
func (d *DB) DeleteFolder(id int64) error {
	_, err := d.Exec(`DELETE FROM folders WHERE id=?`, id)
	return err
}

// ReparentFolder moves a folder under newParentID (nil = root) as the last child.
func (d *DB) ReparentFolder(id int64, newParentID *int64) error {
	var maxPos int
	row := d.QueryRow(
		`SELECT COALESCE(MAX(position), -1) FROM folders WHERE parent_id IS ? AND id != ?`,
		newParentID, id,
	)
	row.Scan(&maxPos)
	now := time.Now().UTC()
	_, err := d.Exec(
		`UPDATE folders SET parent_id=?, position=?, updated_at=? WHERE id=?`,
		newParentID, maxPos+1, now, id,
	)
	return err
}

// MoveFolderUp swaps a folder with the sibling immediately above it (lower position).
func (d *DB) MoveFolderUp(id int64) error {
	return d.moveFolderDir(id, -1)
}

// MoveFolderDown swaps a folder with the sibling immediately below it (higher position).
func (d *DB) MoveFolderDown(id int64) error {
	return d.moveFolderDir(id, +1)
}

func (d *DB) moveFolderDir(id int64, dir int) error {
	// Get parent_id and current position of the target folder.
	var parentID sql.NullInt64
	var curPos int
	err := d.QueryRow(`SELECT parent_id, position FROM folders WHERE id=?`, id).Scan(&parentID, &curPos)
	if err != nil {
		return err
	}

	// Load all siblings (same parent) ordered by position, name.
	rows, err := d.Query(
		`SELECT id, position FROM folders WHERE parent_id IS ? ORDER BY position, name`,
		parentID,
	)
	if err != nil {
		return err
	}
	type sibling struct{ id int64; pos int }
	var siblings []sibling
	for rows.Next() {
		var s sibling
		rows.Scan(&s.id, &s.pos)
		siblings = append(siblings, s)
	}
	rows.Close()

	// Normalise positions to 0,1,2,… so swaps always have distinct values.
	tx, err := d.Begin()
	if err != nil {
		return err
	}
	for i, s := range siblings {
		siblings[i].pos = i
		if _, err := tx.Exec(`UPDATE folders SET position=? WHERE id=?`, i, s.id); err != nil {
			tx.Rollback()
			return err
		}
	}

	// Find this folder's index in the sibling list.
	idx := -1
	for i, s := range siblings {
		if s.id == id {
			idx = i
			break
		}
	}
	swapIdx := idx + dir
	if idx < 0 || swapIdx < 0 || swapIdx >= len(siblings) {
		tx.Rollback()
		return nil // already at boundary — no-op
	}

	// Swap positions.
	posA, posB := siblings[idx].pos, siblings[swapIdx].pos
	if _, err := tx.Exec(`UPDATE folders SET position=? WHERE id=?`, posB, siblings[idx].id); err != nil {
		tx.Rollback()
		return err
	}
	if _, err := tx.Exec(`UPDATE folders SET position=? WHERE id=?`, posA, siblings[swapIdx].id); err != nil {
		tx.Rollback()
		return err
	}
	return tx.Commit()
}
