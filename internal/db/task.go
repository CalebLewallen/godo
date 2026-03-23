package db

import (
	"database/sql"
	"time"

	"github.com/CalebLewallen/godo/internal/model"
)

// GetTasksByCompletion returns all tasks filtered by completed status.
func (d *DB) GetTasksByCompletion(completed bool) ([]model.Task, error) {
	rows, err := d.Query(`
		SELECT id, name, description, due_date, folder_id, completed, completed_at, position, created_at, updated_at
		FROM tasks
		WHERE completed = ?
		ORDER BY folder_id, position, name
	`, completed)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanTasks(rows)
}

// GetTask returns a single task by ID.
func (d *DB) GetTask(id int64) (model.Task, error) {
	row := d.QueryRow(`
		SELECT id, name, description, due_date, folder_id, completed, completed_at, position, created_at, updated_at
		FROM tasks WHERE id=?
	`, id)
	var t model.Task
	var desc, dueDate, completedAt sql.NullString
	var folderID sql.NullInt64
	err := row.Scan(&t.ID, &t.Name, &desc, &dueDate, &folderID, &t.Completed, &completedAt, &t.Position, &t.CreatedAt, &t.UpdatedAt)
	if err != nil {
		return model.Task{}, err
	}
	if desc.Valid {
		t.Description = desc.String
	}
	if dueDate.Valid {
		t.DueDate = dueDate.String
	}
	if folderID.Valid {
		id := folderID.Int64
		t.FolderID = &id
	}
	return t, nil
}

// CreateTask inserts a new task and returns it with the assigned ID.
func (d *DB) CreateTask(name string, folderID *int64) (model.Task, error) {
	now := time.Now().UTC()
	res, err := d.Exec(`
		INSERT INTO tasks (name, folder_id, created_at, updated_at)
		VALUES (?, ?, ?, ?)
	`, name, folderID, now, now)
	if err != nil {
		return model.Task{}, err
	}
	id, _ := res.LastInsertId()
	return model.Task{
		ID:        id,
		Name:      name,
		FolderID:  folderID,
		CreatedAt: now,
		UpdatedAt: now,
	}, nil
}

// SaveTask updates an existing task.
func (d *DB) SaveTask(t model.Task) error {
	now := time.Now().UTC()
	var completedAt interface{}
	if t.Completed {
		if t.CompletedAt.IsZero() {
			completedAt = now
		} else {
			completedAt = t.CompletedAt
		}
	}
	_, err := d.Exec(`
		UPDATE tasks
		SET name=?, description=?, due_date=?, folder_id=?, completed=?, completed_at=?, position=?, updated_at=?
		WHERE id=?
	`, t.Name, nullStr(t.Description), nullStr(t.DueDate), t.FolderID, t.Completed, completedAt, t.Position, now, t.ID)
	return err
}

// RenameTask updates a task's name.
func (d *DB) RenameTask(id int64, name string) error {
	_, err := d.Exec(`UPDATE tasks SET name=?, updated_at=CURRENT_TIMESTAMP WHERE id=?`, name, id)
	return err
}

// DeleteTask deletes a task by ID.
func (d *DB) DeleteTask(id int64) error {
	_, err := d.Exec(`DELETE FROM tasks WHERE id=?`, id)
	return err
}

// ReparentTask moves a task into newFolderID (nil = root) as the last item.
func (d *DB) ReparentTask(id int64, newFolderID *int64) error {
	var maxPos int
	d.QueryRow(
		`SELECT COALESCE(MAX(position), -1) FROM tasks WHERE folder_id IS ? AND id != ?`,
		newFolderID, id,
	).Scan(&maxPos)
	_, err := d.Exec(
		`UPDATE tasks SET folder_id=?, position=?, updated_at=CURRENT_TIMESTAMP WHERE id=?`,
		newFolderID, maxPos+1, id,
	)
	return err
}

// MoveTaskUp swaps a task with the sibling immediately above it.
func (d *DB) MoveTaskUp(id int64) error { return d.moveTaskDir(id, -1) }

// MoveTaskDown swaps a task with the sibling immediately below it.
func (d *DB) MoveTaskDown(id int64) error { return d.moveTaskDir(id, +1) }

func (d *DB) moveTaskDir(id int64, dir int) error {
	var folderID sql.NullInt64
	var curPos int
	if err := d.QueryRow(`SELECT folder_id, position FROM tasks WHERE id=?`, id).Scan(&folderID, &curPos); err != nil {
		return err
	}

	rows, err := d.Query(
		`SELECT id, position FROM tasks WHERE folder_id IS ? ORDER BY position, name`,
		folderID,
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

	tx, err := d.Begin()
	if err != nil {
		return err
	}
	for i, s := range siblings {
		siblings[i].pos = i
		if _, err := tx.Exec(`UPDATE tasks SET position=? WHERE id=?`, i, s.id); err != nil {
			tx.Rollback()
			return err
		}
	}

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
		return nil
	}

	posA, posB := siblings[idx].pos, siblings[swapIdx].pos
	if _, err := tx.Exec(`UPDATE tasks SET position=? WHERE id=?`, posB, siblings[idx].id); err != nil {
		tx.Rollback()
		return err
	}
	if _, err := tx.Exec(`UPDATE tasks SET position=? WHERE id=?`, posA, siblings[swapIdx].id); err != nil {
		tx.Rollback()
		return err
	}
	return tx.Commit()
}

func nullStr(s string) interface{} {
	if s == "" {
		return nil
	}
	return s
}

func scanTasks(rows *sql.Rows) ([]model.Task, error) {
	var tasks []model.Task
	for rows.Next() {
		var t model.Task
		var desc, dueDate sql.NullString
		var completedAt sql.NullTime
		var folderID sql.NullInt64
		if err := rows.Scan(&t.ID, &t.Name, &desc, &dueDate, &folderID, &t.Completed, &completedAt, &t.Position, &t.CreatedAt, &t.UpdatedAt); err != nil {
			return nil, err
		}
		if desc.Valid {
			t.Description = desc.String
		}
		if dueDate.Valid {
			t.DueDate = dueDate.String
		}
		if folderID.Valid {
			id := folderID.Int64
			t.FolderID = &id
		}
		if completedAt.Valid {
			t.CompletedAt = completedAt.Time
		}
		tasks = append(tasks, t)
	}
	return tasks, rows.Err()
}
