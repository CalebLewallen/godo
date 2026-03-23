package model

import "time"

// Folder represents a folder in the sidebar tree.
type Folder struct {
	ID        int64
	Name      string
	ParentID  *int64
	Position  int
	CreatedAt time.Time
	UpdatedAt time.Time
}

// Task represents a to-do item.
type Task struct {
	ID          int64
	Name        string
	Description string
	DueDate     string // stored as string for flexible display
	FolderID    *int64
	Completed   bool
	CompletedAt time.Time
	Position    int
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
