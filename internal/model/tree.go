package model

import "sort"

// NodeKind distinguishes folders from tasks in the tree.
type NodeKind int

const (
	NodeFolder NodeKind = iota
	NodeTask
)

// TreeNode is a flattened row in the sidebar tree.
type TreeNode struct {
	Kind     NodeKind
	Depth    int
	Expanded bool
	Folder   *Folder
	Task     *Task
}

// ID returns the underlying item ID.
func (n TreeNode) ID() int64 {
	if n.Kind == NodeFolder {
		return n.Folder.ID
	}
	return n.Task.ID
}

// Label returns the display label.
func (n TreeNode) Label() string {
	if n.Kind == NodeFolder {
		return n.Folder.Name
	}
	return n.Task.Name
}

// ParentID returns the container ID of this node (nil = root).
func (n TreeNode) ParentID() *int64 {
	if n.Kind == NodeFolder {
		return n.Folder.ParentID
	}
	return n.Task.FolderID
}

// BuildTree builds a flattened, depth-first tree from folders and tasks,
// interleaving them by position at each level.
func BuildTree(folders []Folder, tasks []Task, expandedFolders map[int64]bool) []TreeNode {
	var nodes []TreeNode
	appendLevel(nil, folders, tasks, expandedFolders, 0, &nodes)
	return nodes
}

// levelItem is a holder used during merge-sort at a single tree level.
type levelItem struct {
	pos      int
	name     string
	isFolder bool
	folder   *Folder
	task     *Task
}

func appendLevel(parentID *int64, folders []Folder, tasks []Task, expandedFolders map[int64]bool, depth int, nodes *[]TreeNode) {
	var items []levelItem

	for i := range folders {
		f := &folders[i]
		if ptrEq(f.ParentID, parentID) {
			items = append(items, levelItem{pos: f.Position, name: f.Name, isFolder: true, folder: f})
		}
	}
	for i := range tasks {
		t := &tasks[i]
		if ptrEq(t.FolderID, parentID) {
			items = append(items, levelItem{pos: t.Position, name: t.Name, isFolder: false, task: t})
		}
	}

	// Sort by position first, then name as a stable tiebreaker.
	sort.SliceStable(items, func(i, j int) bool {
		if items[i].pos != items[j].pos {
			return items[i].pos < items[j].pos
		}
		return items[i].name < items[j].name
	})

	for _, item := range items {
		if item.isFolder {
			f := item.folder
			expanded := expandedFolders[f.ID]
			*nodes = append(*nodes, TreeNode{Kind: NodeFolder, Depth: depth, Expanded: expanded, Folder: f})
			if expanded {
				appendLevel(&f.ID, folders, tasks, expandedFolders, depth+1, nodes)
			}
		} else {
			*nodes = append(*nodes, TreeNode{Kind: NodeTask, Depth: depth, Task: item.task})
		}
	}
}

func ptrEq(a, b *int64) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return *a == *b
}
