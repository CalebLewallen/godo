package db

import "sort"

// SwapSiblings swaps the visual positions of two nodes (folders or tasks) that
// share the same parent level. parentID nil means the root level.
// It normalises all siblings first so positions are always distinct.
func (d *DB) SwapSiblings(parentID *int64, aIsFolder bool, aID int64, bIsFolder bool, bID int64) error {
	type item struct {
		id       int64
		pos      int
		isFolder bool
	}

	// Load all folders at this level.
	fRows, err := d.Query(
		`SELECT id, position FROM folders WHERE parent_id IS ? ORDER BY position, name`,
		parentID,
	)
	if err != nil {
		return err
	}
	var items []item
	for fRows.Next() {
		var it item
		it.isFolder = true
		fRows.Scan(&it.id, &it.pos)
		items = append(items, it)
	}
	fRows.Close()

	// Load all tasks at this level.
	tRows, err := d.Query(
		`SELECT id, position FROM tasks WHERE folder_id IS ? ORDER BY position, name`,
		parentID,
	)
	if err != nil {
		return err
	}
	for tRows.Next() {
		var it item
		it.isFolder = false
		tRows.Scan(&it.id, &it.pos)
		items = append(items, it)
	}
	tRows.Close()

	// Sort by position (stable: preserves DB name-ordering within ties).
	sort.SliceStable(items, func(i, j int) bool {
		return items[i].pos < items[j].pos
	})

	tx, err := d.Begin()
	if err != nil {
		return err
	}

	// Normalise to 0, 1, 2, … so every swap produces distinct values.
	for i := range items {
		items[i].pos = i
	}

	// Find the two items to swap.
	idxA, idxB := -1, -1
	for i, it := range items {
		if it.isFolder == aIsFolder && it.id == aID {
			idxA = i
		}
		if it.isFolder == bIsFolder && it.id == bID {
			idxB = i
		}
	}
	if idxA < 0 || idxB < 0 {
		tx.Rollback()
		return nil
	}
	items[idxA].pos, items[idxB].pos = items[idxB].pos, items[idxA].pos

	// Write back the two changed items (and the normalised positions for all).
	for _, it := range items {
		var err error
		if it.isFolder {
			_, err = tx.Exec(`UPDATE folders SET position=? WHERE id=?`, it.pos, it.id)
		} else {
			_, err = tx.Exec(`UPDATE tasks SET position=? WHERE id=?`, it.pos, it.id)
		}
		if err != nil {
			tx.Rollback()
			return err
		}
	}
	return tx.Commit()
}
