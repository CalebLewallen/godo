package ui

import "github.com/CalebLewallen/godo/internal/model"

// OpenTaskMsg signals the task panel to load and show a task.
type OpenTaskMsg struct{ TaskID int64 }

// TaskSavedMsg is emitted after a task is successfully saved.
type TaskSavedMsg struct{ Task model.Task }

// TaskClosedMsg is emitted when the task panel is closed.
type TaskClosedMsg struct{}

// TreeLoadedMsg carries fresh data from the DB for the sidebar.
type TreeLoadedMsg struct {
	Folders []model.Folder
	Tasks   []model.Task
}

// NewTaskMsg requests a new task be created (optionally inside a folder).
type NewTaskMsg struct{ FolderID *int64 }

// NewFolderMsg requests a new folder be created.
type NewFolderMsg struct{ ParentID *int64 }

// CreateNamedFolderMsg requests a new folder be created with a user-supplied name.
type CreateNamedFolderMsg struct {
	Name     string
	ParentID *int64
}

// RenameFolderMsg requests a folder be renamed.
type RenameFolderMsg struct {
	FolderID int64
	Name     string
}

// RenameTaskMsg requests a task be renamed.
type RenameTaskMsg struct {
	TaskID int64
	Name   string
}

// ReparentFolderMsg requests a folder be moved under a new parent (nil = root).
type ReparentFolderMsg struct {
	FolderID    int64
	NewParentID *int64
}

// ReparentTaskMsg requests a task be moved into a different folder (nil = root).
type ReparentTaskMsg struct {
	TaskID      int64
	NewFolderID *int64
}

// SwapSiblingsMsg swaps the positions of two adjacent nodes at the same level.
type SwapSiblingsMsg struct {
	A, B     model.TreeNode
	ParentID *int64 // shared parent (nil = root)
}

// DeleteTaskMsg requests a task be deleted.
type DeleteTaskMsg struct{ TaskID int64 }

// DeleteFolderMsg requests a folder be deleted.
type DeleteFolderMsg struct{ FolderID int64 }

// ErrMsg carries a non-fatal error to display.
type ErrMsg struct{ Err error }
