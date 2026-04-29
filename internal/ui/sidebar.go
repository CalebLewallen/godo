package ui

import (
	"strings"

	"github.com/CalebLewallen/godo/internal/model"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type sidebarTab int

const (
	tabIncomplete sidebarTab = iota
	tabCompleted
)

type renameMode int

const (
	renameModeNone      renameMode = iota
	renameModeNewFolder            // creating a new folder, waiting for name
	renameModeFolder               // renaming an existing folder
	renameModeTask                 // renaming an existing task
)

// SidebarModel renders the left-side folder/task tree.
type SidebarModel struct {
	focused         bool
	width           int
	height          int
	activeTab       sidebarTab
	folders         []model.Folder
	folderMap       map[int64]*model.Folder
	incompleteTasks []model.Task
	completedTasks  []model.Task
	expandedFolders map[int64]bool
	nodes           []model.TreeNode
	cursor          int
	filterInput     textinput.Model
	filtering       bool
	filterText      string
	renaming        renameMode
	renameInput     textinput.Model
	renameID        int64  // ID of the item being renamed
	renameParent    *int64 // parent for a new folder
	renameDepth     int    // visual indent depth for new folder row
	confirmDelete   bool   // waiting for y/n confirmation to delete
	pendingFocusKind model.NodeKind
	pendingFocusID   int64
}

func NewSidebarModel() SidebarModel {
	filter := textinput.New()
	filter.Placeholder = "filter..."
	filter.CharLimit = 80

	rename := textinput.New()
	rename.CharLimit = 200

	return SidebarModel{
		expandedFolders: make(map[int64]bool),
		filterInput:     filter,
		renameInput:     rename,
		folderMap:       make(map[int64]*model.Folder),
	}
}

func (m SidebarModel) Init() tea.Cmd {
	return nil
}

func (m SidebarModel) Update(msg tea.Msg) (SidebarModel, tea.Cmd) {
	var cmds []tea.Cmd

	if m.confirmDelete {
		if msg, ok := msg.(tea.KeyMsg); ok {
			switch msg.String() {
			case "y", "Y":
				m.confirmDelete = false
				if m.cursor < len(m.nodes) {
					node := m.nodes[m.cursor]
					if node.Kind == model.NodeFolder {
						id := node.Folder.ID
						return m, func() tea.Msg { return DeleteFolderMsg{FolderID: id} }
					}
					id := node.Task.ID
					return m, func() tea.Msg { return DeleteTaskMsg{TaskID: id} }
				}
			case "n", "N", "esc":
				m.confirmDelete = false
			}
		}
		return m, nil
	}

	if m.renaming != renameModeNone {
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case "esc":
				m.renaming = renameModeNone
				m.renameInput.Blur()
				return m, nil
			case "enter":
				name := strings.TrimSpace(m.renameInput.Value())
				mode := m.renaming
				m.renaming = renameModeNone
				m.renameInput.Blur()
				if name == "" {
					return m, nil
				}
				switch mode {
				case renameModeNewFolder:
					parent := m.renameParent
					return m, func() tea.Msg { return CreateNamedFolderMsg{Name: name, ParentID: parent} }
				case renameModeFolder:
					id := m.renameID
					return m, func() tea.Msg { return RenameFolderMsg{FolderID: id, Name: name} }
				case renameModeTask:
					id := m.renameID
					return m, func() tea.Msg { return RenameTaskMsg{TaskID: id, Name: name} }
				}
			default:
				var cmd tea.Cmd
				m.renameInput, cmd = m.renameInput.Update(msg)
				return m, cmd
			}
		}
		return m, nil
	}

	if m.filtering {
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case "esc":
				m.filtering = false
				m.filterText = ""
				m.filterInput.SetValue("")
				m.filterInput.Blur()
				m.rebuildNodes()
				return m, nil
			case "enter":
				m.filtering = false
				m.filterText = m.filterInput.Value()
				m.filterInput.Blur()
				m.rebuildNodes()
				return m, nil
			}
		}
		var tiCmd tea.Cmd
		m.filterInput, tiCmd = m.filterInput.Update(msg)
		m.filterText = m.filterInput.Value()
		m.rebuildNodes()
		cmds = append(cmds, tiCmd)
		return m, tea.Batch(cmds...)
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, SidebarBindings.NextTab):
			m.cycleTab()
			m.rebuildNodes()
			m.cursor = 0

		case key.Matches(msg, SidebarBindings.Up):
			if m.cursor > 0 {
				m.cursor--
			}

		case key.Matches(msg, SidebarBindings.Down):
			if m.cursor < len(m.nodes)-1 {
				m.cursor++
			}

		case key.Matches(msg, SidebarBindings.Expand):
			if m.cursor < len(m.nodes) {
				node := m.nodes[m.cursor]
				if node.Kind == model.NodeFolder {
					m.expandedFolders[node.Folder.ID] = true
					m.rebuildNodes()
				}
			}

		case key.Matches(msg, SidebarBindings.Collapse):
			if m.cursor < len(m.nodes) {
				node := m.nodes[m.cursor]
				if node.Kind == model.NodeFolder {
					m.expandedFolders[node.Folder.ID] = false
					m.rebuildNodes()
				}
			}

		case key.Matches(msg, SidebarBindings.Enter):
			if m.cursor < len(m.nodes) {
				node := m.nodes[m.cursor]
				if node.Kind == model.NodeFolder {
					m.expandedFolders[node.Folder.ID] = !m.expandedFolders[node.Folder.ID]
					m.rebuildNodes()
				} else {
					return m, func() tea.Msg { return OpenTaskMsg{TaskID: node.Task.ID} }
				}
			}

		case key.Matches(msg, SidebarBindings.Rename):
			if m.cursor < len(m.nodes) {
				node := m.nodes[m.cursor]
				var cmd tea.Cmd
				if node.Kind == model.NodeFolder {
					cmd = m.startRename(renameModeFolder, node.Folder.ID, node.Folder.Name)
				} else {
					cmd = m.startRename(renameModeTask, node.Task.ID, node.Task.Name)
				}
				return m, cmd
			}

		case key.Matches(msg, SidebarBindings.IndentFolder):
			if m.cursor < len(m.nodes) {
				node := m.nodes[m.cursor]
				switch node.Kind {
				case model.NodeFolder:
					if cmd := m.indentFolder(node); cmd != nil {
						m.pendingFocusKind = node.Kind
						m.pendingFocusID = node.Folder.ID
						return m, cmd
					}
				case model.NodeTask:
					if cmd := m.indentTask(node); cmd != nil {
						m.pendingFocusKind = node.Kind
						m.pendingFocusID = node.Task.ID
						return m, cmd
					}
				}
			}

		case key.Matches(msg, SidebarBindings.DedentFolder):
			if m.cursor < len(m.nodes) {
				node := m.nodes[m.cursor]
				switch node.Kind {
				case model.NodeFolder:
					if node.Folder.ParentID != nil {
						newParent := m.grandparentID(node.Folder)
						folderID := node.Folder.ID
						m.pendingFocusKind = node.Kind
						m.pendingFocusID = folderID
						return m, func() tea.Msg {
							return ReparentFolderMsg{FolderID: folderID, NewParentID: newParent}
						}
					}
				case model.NodeTask:
					if node.Task.FolderID != nil {
						taskID := node.Task.ID
						m.pendingFocusKind = node.Kind
						m.pendingFocusID = taskID
						return m, func() tea.Msg {
							return ReparentTaskMsg{TaskID: taskID, NewFolderID: nil}
						}
					}
				}
			}

		case key.Matches(msg, SidebarBindings.MoveUp):
			if cmd := m.swapWithSibling(-1); cmd != nil {
				return m, cmd
			}

		case key.Matches(msg, SidebarBindings.MoveDown):
			if cmd := m.swapWithSibling(+1); cmd != nil {
				return m, cmd
			}

		case key.Matches(msg, SidebarBindings.ExpandAll):
			m.ExpandAll()

		case key.Matches(msg, SidebarBindings.CollapseAll):
			m.CollapseAll()

		case key.Matches(msg, SidebarBindings.Delete):
			if m.cursor < len(m.nodes) {
				m.confirmDelete = true
			}
		}

	case TreeLoadedMsg:
		m.folders = msg.Folders
		m.folderMap = make(map[int64]*model.Folder, len(m.folders))
		for i := range m.folders {
			m.folderMap[m.folders[i].ID] = &m.folders[i]
		}

		m.incompleteTasks = msg.Tasks
		// Filter out completed
		var incomplete, completed []model.Task
		for _, t := range msg.Tasks {
			if t.Completed {
				completed = append(completed, t)
			} else {
				incomplete = append(incomplete, t)
			}
		}
		m.incompleteTasks = incomplete
		m.completedTasks = completed
		m.rebuildNodes()

	case TaskSavedMsg:
		// Reload will happen via app-level reload
	}

	return m, tea.Batch(cmds...)
}

func (m *SidebarModel) cycleTab() {
	if m.activeTab == tabIncomplete {
		m.activeTab = tabCompleted
	} else {
		m.activeTab = tabIncomplete
	}
}

func (m *SidebarModel) rebuildNodes() {
	var tasks []model.Task
	if m.activeTab == tabIncomplete {
		tasks = m.incompleteTasks
	} else {
		tasks = m.completedTasks
	}

	if m.activeTab == tabCompleted {
		// Flat list — no folder grouping for completed tasks.
		filter := strings.ToLower(m.filterText)
		m.nodes = m.nodes[:0]
		for i := range tasks {
			t := &tasks[i]
			if filter == "" || strings.Contains(strings.ToLower(t.Name), filter) {
				m.nodes = append(m.nodes, model.TreeNode{Kind: model.NodeTask, Depth: 0, Task: t})
			}
		}
	} else if m.filterText != "" {
		filter := strings.ToLower(m.filterText)
		var filtered []model.Task
		for _, t := range tasks {
			if strings.Contains(strings.ToLower(t.Name), filter) {
				filtered = append(filtered, t)
			}
		}
		var filteredFolders []model.Folder
		for _, f := range m.folders {
			if strings.Contains(strings.ToLower(f.Name), filter) {
				filteredFolders = append(filteredFolders, f)
			}
		}
		// Also include folders that contain matching tasks
		folderHasMatch := make(map[int64]bool)
		for _, t := range filtered {
			if t.FolderID != nil {
				folderHasMatch[*t.FolderID] = true
			}
		}
		for _, f := range m.folders {
			if folderHasMatch[f.ID] {
				alreadyIn := false
				for _, ef := range filteredFolders {
					if ef.ID == f.ID {
						alreadyIn = true
						break
					}
				}
				if !alreadyIn {
					filteredFolders = append(filteredFolders, f)
					m.expandedFolders[f.ID] = true
				}
			}
		}
		m.nodes = model.BuildTree(filteredFolders, filtered, m.expandedFolders)
	} else {
		m.nodes = model.BuildTree(m.folders, tasks, m.expandedFolders)
	}

	if m.cursor >= len(m.nodes) && len(m.nodes) > 0 {
		m.cursor = len(m.nodes) - 1
	}

	if m.pendingFocusID != 0 {
		for i, n := range m.nodes {
			if n.Kind == m.pendingFocusKind {
				if n.Kind == model.NodeFolder && n.Folder.ID == m.pendingFocusID {
					m.cursor = i
					break
				} else if n.Kind == model.NodeTask && n.Task.ID == m.pendingFocusID {
					m.cursor = i
					break
				}
			}
		}
		m.pendingFocusID = 0
	}
}

// indentFolder makes the folder a child of the nearest folder above it in the tree.
func (m *SidebarModel) indentFolder(node model.TreeNode) tea.Cmd {
	for i := m.cursor - 1; i >= 0; i-- {
		above := m.nodes[i]
		if above.Kind == model.NodeFolder && above.Folder.ID != node.Folder.ID {
			newParentID := above.Folder.ID
			folderID := node.Folder.ID
			// Auto-expand the new parent so the moved folder is visible.
			m.expandedFolders[newParentID] = true
			return func() tea.Msg {
				return ReparentFolderMsg{FolderID: folderID, NewParentID: &newParentID}
			}
		}
	}
	return nil // no folder above — nothing to indent into
}

// swapWithSibling finds the previous (dir=-1) or next (dir=+1) sibling at the
// same depth in the flat tree and emits a SwapSiblingsMsg.
func (m *SidebarModel) swapWithSibling(dir int) tea.Cmd {
	if m.cursor >= len(m.nodes) {
		return nil
	}
	cur := m.nodes[m.cursor]
	// Search for the nearest node at the same depth in the given direction.
	if dir < 0 {
		for i := m.cursor - 1; i >= 0; i-- {
			if m.nodes[i].Depth < cur.Depth {
				return nil // hit parent boundary
			}
			if m.nodes[i].Depth == cur.Depth {
				sibling := m.nodes[i]
				parentID := cur.ParentID()
				m.pendingFocusKind = cur.Kind
				if cur.Kind == model.NodeFolder {
					m.pendingFocusID = cur.Folder.ID
				} else {
					m.pendingFocusID = cur.Task.ID
				}
				return func() tea.Msg { return SwapSiblingsMsg{A: cur, B: sibling, ParentID: parentID} }
			}
		}
	} else {
		for i := m.cursor + 1; i < len(m.nodes); i++ {
			if m.nodes[i].Depth < cur.Depth {
				return nil // hit parent boundary
			}
			if m.nodes[i].Depth == cur.Depth {
				sibling := m.nodes[i]
				parentID := cur.ParentID()
				m.pendingFocusKind = cur.Kind
				if cur.Kind == model.NodeFolder {
					m.pendingFocusID = cur.Folder.ID
				} else {
					m.pendingFocusID = cur.Task.ID
				}
				return func() tea.Msg { return SwapSiblingsMsg{A: cur, B: sibling, ParentID: parentID} }
			}
		}
	}
	return nil
}

// indentTask moves a task into the nearest folder above it in the tree.
func (m *SidebarModel) indentTask(node model.TreeNode) tea.Cmd {
	for i := m.cursor - 1; i >= 0; i-- {
		above := m.nodes[i]
		if above.Kind == model.NodeFolder {
			newFolderID := above.Folder.ID
			taskID := node.Task.ID
			m.expandedFolders[newFolderID] = true
			return func() tea.Msg {
				return ReparentTaskMsg{TaskID: taskID, NewFolderID: &newFolderID}
			}
		}
	}
	return nil
}

// grandparentID returns the parent_id of the folder's parent (the folder's grandparent).
func (m *SidebarModel) grandparentID(f *model.Folder) *int64 {
	if f.ParentID == nil {
		return nil
	}
	if parent, ok := m.folderMap[*f.ParentID]; ok {
		return parent.ParentID
	}
	return nil
}

// StartNewFolder opens an inline name prompt for a new folder.
func (m *SidebarModel) StartNewFolder(parentID *int64) tea.Cmd {
	m.renaming = renameModeNewFolder
	m.renameParent = parentID
	// Compute indent depth from cursor position.
	m.renameDepth = 0
	if m.cursor < len(m.nodes) {
		node := m.nodes[m.cursor]
		if node.Kind == model.NodeFolder {
			m.renameDepth = node.Depth + 1
		} else {
			m.renameDepth = node.Depth
		}
	}
	m.renameInput.Placeholder = "Folder name..."
	m.renameInput.SetValue("")
	return m.renameInput.Focus()
}

// startRename opens the inline input to rename an existing folder or task.
func (m *SidebarModel) startRename(mode renameMode, id int64, current string) tea.Cmd {
	m.renaming = mode
	m.renameID = id
	m.renameInput.Placeholder = ""
	m.renameInput.SetValue(current)
	m.renameInput.CursorEnd()
	return m.renameInput.Focus()
}

// ExpandAll expands every folder in the tree.
func (m *SidebarModel) ExpandAll() {
	for _, f := range m.folders {
		m.expandedFolders[f.ID] = true
	}
	m.rebuildNodes()
}

// CollapseAll collapses every folder in the tree.
func (m *SidebarModel) CollapseAll() {
	m.expandedFolders = make(map[int64]bool)
	m.rebuildNodes()
}

// StartFilter activates the inline filter input.
func (m *SidebarModel) StartFilter() tea.Cmd {
	m.filtering = true
	return m.filterInput.Focus()
}

// CurrentFolderID returns the folder ID at the cursor (or nil).
func (m *SidebarModel) CurrentFolderID() *int64 {
	if m.cursor >= len(m.nodes) {
		return nil
	}
	node := m.nodes[m.cursor]
	if node.Kind == model.NodeFolder {
		return &node.Folder.ID
	}
	if node.Kind == model.NodeTask && node.Task.FolderID != nil {
		return node.Task.FolderID
	}
	return nil
}

func (m SidebarModel) View() string {
	innerWidth := m.width - 4 // account for border + padding

	// Tabs
	tab1 := tabStyle.Render("My Todo's")
	tab2 := tabStyle.Render("Completed")
	if m.activeTab == tabIncomplete {
		tab1 = activeTabStyle.Render("My Todo's")
	} else {
		tab2 = activeTabStyle.Render("Completed")
	}
	tabs := lipgloss.JoinHorizontal(lipgloss.Top, tab1, " ", tab2)

	var sb strings.Builder
	sb.WriteString(tabs)
	sb.WriteString("\n")
	sb.WriteString(dividerStyle.Render(strings.Repeat("─", innerWidth)))
	sb.WriteString("\n")

	if m.filtering {
		sb.WriteString(fieldLabelStyle.Render("Filter: "))
		sb.WriteString(m.filterInput.View())
		sb.WriteString("\n")
		sb.WriteString(dividerStyle.Render(strings.Repeat("─", innerWidth)))
		sb.WriteString("\n")
	}

	// Tree rows
	visibleHeight := m.height - 6 // reserve for tabs, divider, help
	if visibleHeight < 1 {
		visibleHeight = 1
	}

	// Scroll window
	start := 0
	if m.cursor >= visibleHeight {
		start = m.cursor - visibleHeight + 1
	}
	end := start + visibleHeight
	if end > len(m.nodes) {
		end = len(m.nodes)
	}

	if len(m.nodes) == 0 {
		sb.WriteString(mutedStyle.Render("  (empty)"))
		sb.WriteString("\n")
	}

	for i := start; i < end; i++ {
		node := m.nodes[i]
		var row string
		if m.renaming != renameModeNone && i == m.cursor && m.renaming != renameModeNewFolder {
			// Render inline rename input in place of the label.
			indent := strings.Repeat("  ", node.Depth)
			var icon string
			if node.Kind == model.NodeFolder {
				icon = folderStyle.Render("▶ /")
			} else {
				icon = taskStyle.Render("*")
			}
			row = indent + icon + " " + m.renameInput.View()
		} else {
			row = renderNode(node, innerWidth)
			if i == m.cursor && m.focused {
				row = selectedRowStyle.Width(innerWidth).Render(row)
			}
		}
		sb.WriteString(row)
		sb.WriteString("\n")

		// For new folder mode, inject an input row immediately after the cursor row.
		if m.renaming == renameModeNewFolder && i == m.cursor {
			indent := strings.Repeat("  ", m.renameDepth)
			newRow := indent + folderStyle.Render("▶ /") + " " + m.renameInput.View()
			sb.WriteString(newRow)
			sb.WriteString("\n")
		}
	}

	// Help line / confirm prompt
	if m.confirmDelete {
		sb.WriteString(errorStyle.Render("Delete? ") + helpStyle.Render("(y) yes  (n) no"))
	} else {
		sb.WriteString(helpStyle.Render("?  help"))
	}

	content := sb.String()

	style := sidebarStyle.Width(m.width - 2).Height(m.height - 2)
	if m.focused {
		style = sidebarFocusedStyle.Width(m.width - 2).Height(m.height - 2)
	}
	return style.Render(content)
}

func renderNode(node model.TreeNode, maxWidth int) string {
	indent := strings.Repeat("  ", node.Depth)
	var icon, label string

	if node.Kind == model.NodeFolder {
		if node.Expanded {
			icon = folderStyle.Render("▼ /")
		} else {
			icon = folderStyle.Render("▶ /")
		}
		label = folderStyle.Render(node.Folder.Name)
	} else {
		if node.Task.Completed {
			icon = doneStyle.Render("✓")
			label = doneStyle.Render(node.Task.Name)
		} else {
			icon = taskStyle.Render("*")
			label = taskStyle.Render(node.Task.Name)
		}
	}

	text := indent + icon + " " + label

	// Truncate if needed
	visLen := lipgloss.Width(text)
	if visLen > maxWidth {
		// crude truncation
		runes := []rune(node.Label())
		for len(indent)+1+1+len(runes) > maxWidth && len(runes) > 0 {
			runes = runes[:len(runes)-1]
		}
		label = string(runes) + "…"
		if node.Kind == model.NodeFolder {
			text = indent + folderStyle.Render("▶ /") + " " + folderStyle.Render(label)
		} else {
			text = indent + taskStyle.Render("*") + " " + taskStyle.Render(label)
		}
	}

	return text
}
