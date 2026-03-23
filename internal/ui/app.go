package ui

import (
	"github.com/CalebLewallen/godo/internal/db"
	"github.com/CalebLewallen/godo/internal/model"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type focusArea int

const (
	focusSidebar focusArea = iota
	focusTask
)

// AppModel is the root Bubble Tea model.
type AppModel struct {
	db          *db.DB
	sidebar     SidebarModel
	taskPanel   TaskPanelModel
	helpModal   HelpModal
	quickOpen   QuickOpenModal
	focus       focusArea
	sidebarOpen bool
	width       int
	height      int
}

const sidebarWidth = 32

// NewAppModel creates the root model and loads initial data.
func NewAppModel(database *db.DB) AppModel {
	m := AppModel{
		db:          database,
		sidebar:     NewSidebarModel(),
		taskPanel:   NewTaskPanelModel(database),
		quickOpen:   newQuickOpenModal(),
		focus:       focusSidebar,
		sidebarOpen: true,
	}
	m.sidebar.focused = true
	return m
}

func (m AppModel) Init() tea.Cmd {
	return m.loadTree()
}

func (m AppModel) loadTree() tea.Cmd {
	return func() tea.Msg {
		return m.reloadTreeMsg()
	}
}

func (m AppModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.resize()

	case tea.KeyMsg:
		// Quick-open modal intercepts all keys when open.
		if m.quickOpen.visible {
			cmd, outMsg := m.quickOpen.Update(msg)
			if outMsg != nil {
				return m, func() tea.Msg { return outMsg }
			}
			return m, cmd
		}

		// Help modal intercepts all keys when open.
		if m.helpModal.visible {
			switch msg.String() {
			case "esc", "?":
				m.helpModal.Close()
			case "up", "k":
				m.helpModal.ScrollUp()
			case "down", "j":
				m.helpModal.ScrollDown()
			}
			return m, nil
		}

		// Global keys
		switch {
		case key.Matches(msg, Keys.QuickOpen):
			cmd := m.quickOpen.Open(m.sidebar.incompleteTasks)
			return m, cmd

		case key.Matches(msg, Keys.ShowHelp):
			m.helpModal.Open()
			return m, nil

		case key.Matches(msg, Keys.Quit):
			return m, tea.Quit

		case key.Matches(msg, Keys.ToggleFocus):
			m.toggleFocus()
			return m, nil

		case key.Matches(msg, Keys.ToggleSidebar):
			m.sidebarOpen = !m.sidebarOpen
			m.resize()
			return m, nil

		case key.Matches(msg, Keys.Filter):
			if m.sidebarOpen {
				m.focus = focusSidebar
				m.updateFocusStyles()
				cmd := m.sidebar.StartFilter()
				return m, cmd
			}

		case key.Matches(msg, Keys.NewTask):
			folderID := m.sidebar.CurrentFolderID()
			m.focus = focusTask
			m.updateFocusStyles()
			var cmd tea.Cmd
			m.taskPanel, cmd = m.taskPanel.Update(NewTaskMsg{FolderID: folderID})
			cmds = append(cmds, cmd)
			return m, tea.Batch(cmds...)

		case key.Matches(msg, Keys.NewFolder):
			folderID := m.sidebar.CurrentFolderID()
			m.focus = focusSidebar
			m.sidebarOpen = true
			m.updateFocusStyles()
			m.resize()
			cmd := m.sidebar.StartNewFolder(folderID)
			return m, cmd

		}

		// Route to focused component
		if m.focus == focusSidebar && m.sidebarOpen {
			var cmd tea.Cmd
			m.sidebar, cmd = m.sidebar.Update(msg)
			cmds = append(cmds, cmd)
		} else {
			var cmd tea.Cmd
			m.taskPanel, cmd = m.taskPanel.Update(msg)
			cmds = append(cmds, cmd)
		}

	case OpenTaskMsg:
		m.focus = focusTask
		m.updateFocusStyles()
		var cmd tea.Cmd
		m.taskPanel, cmd = m.taskPanel.Update(msg)
		cmds = append(cmds, cmd)

	case TaskSavedMsg:
		// Reload tree after save
		var cmd tea.Cmd
		m.taskPanel, cmd = m.taskPanel.Update(msg)
		cmds = append(cmds, cmd, m.loadTree())

	case TaskClosedMsg:
		m.taskPanel.hasTask = false
		m.focus = focusSidebar
		m.updateFocusStyles()

	case TreeLoadedMsg:
		var sideCmd, panelCmd tea.Cmd
		m.sidebar, sideCmd = m.sidebar.Update(msg)
		m.taskPanel, panelCmd = m.taskPanel.Update(msg)
		m.quickOpen.SetTasks(msg.Tasks)
		cmds = append(cmds, sideCmd, panelCmd)

	case NewTaskMsg:
		m.focus = focusTask
		m.updateFocusStyles()
		var cmd tea.Cmd
		m.taskPanel, cmd = m.taskPanel.Update(msg)
		cmds = append(cmds, cmd)

	case CreateNamedFolderMsg:
		cmds = append(cmds, m.createNamedFolder(msg.Name, msg.ParentID))

	case RenameFolderMsg:
		cmds = append(cmds, m.renameFolder(msg.FolderID, msg.Name))

	case RenameTaskMsg:
		cmds = append(cmds, m.renameTask(msg.TaskID, msg.Name))

	case ReparentFolderMsg:
		cmds = append(cmds, m.reparentFolder(msg.FolderID, msg.NewParentID))

	case ReparentTaskMsg:
		cmds = append(cmds, m.reparentTask(msg.TaskID, msg.NewFolderID))

	case SwapSiblingsMsg:
		cmds = append(cmds, m.swapSiblings(msg))

	case DeleteTaskMsg:
		cmds = append(cmds, m.deleteTask(msg.TaskID))

	case DeleteFolderMsg:
		cmds = append(cmds, m.deleteFolder(msg.FolderID))

	case ErrMsg:
		// Could show error in status bar; for now ignore non-fatal

	default:
		// Forward to both
		var sideCmd, panelCmd tea.Cmd
		m.sidebar, sideCmd = m.sidebar.Update(msg)
		m.taskPanel, panelCmd = m.taskPanel.Update(msg)
		cmds = append(cmds, sideCmd, panelCmd)
	}

	return m, tea.Batch(cmds...)
}

func (m *AppModel) toggleFocus() {
	if m.focus == focusSidebar {
		m.focus = focusTask
	} else {
		m.focus = focusSidebar
	}
	m.updateFocusStyles()
}

func (m *AppModel) updateFocusStyles() {
	m.sidebar.focused = m.focus == focusSidebar
	m.taskPanel.focused = m.focus == focusTask
}

func (m *AppModel) resize() {
	if m.sidebarOpen {
		m.sidebar.width = sidebarWidth
		m.sidebar.height = m.height
		m.taskPanel.width = m.width - sidebarWidth
		m.taskPanel.height = m.height
	} else {
		m.sidebar.width = 0
		m.sidebar.height = 0
		m.taskPanel.width = m.width
		m.taskPanel.height = m.height
	}
}


func (m *AppModel) createNamedFolder(name string, parentID *int64) tea.Cmd {
	return func() tea.Msg {
		if _, err := m.db.CreateFolder(name, parentID); err != nil {
			return ErrMsg{Err: err}
		}
		return m.reloadTreeMsg()
	}
}

func (m *AppModel) renameFolder(id int64, name string) tea.Cmd {
	return func() tea.Msg {
		if err := m.db.RenameFolder(id, name); err != nil {
			return ErrMsg{Err: err}
		}
		return m.reloadTreeMsg()
	}
}

func (m *AppModel) renameTask(id int64, name string) tea.Cmd {
	return func() tea.Msg {
		if err := m.db.RenameTask(id, name); err != nil {
			return ErrMsg{Err: err}
		}
		return m.reloadTreeMsg()
	}
}

func (m *AppModel) reparentFolder(folderID int64, newParentID *int64) tea.Cmd {
	return func() tea.Msg {
		if err := m.db.ReparentFolder(folderID, newParentID); err != nil {
			return ErrMsg{Err: err}
		}
		return m.reloadTreeMsg()
	}
}

func (m *AppModel) reparentTask(taskID int64, newFolderID *int64) tea.Cmd {
	return func() tea.Msg {
		if err := m.db.ReparentTask(taskID, newFolderID); err != nil {
			return ErrMsg{Err: err}
		}
		return m.reloadTreeMsg()
	}
}

func (m *AppModel) swapSiblings(msg SwapSiblingsMsg) tea.Cmd {
	aIsFolder := msg.A.Kind == model.NodeFolder
	bIsFolder := msg.B.Kind == model.NodeFolder
	var aID, bID int64
	if aIsFolder {
		aID = msg.A.Folder.ID
	} else {
		aID = msg.A.Task.ID
	}
	if bIsFolder {
		bID = msg.B.Folder.ID
	} else {
		bID = msg.B.Task.ID
	}
	return func() tea.Msg {
		if err := m.db.SwapSiblings(msg.ParentID, aIsFolder, aID, bIsFolder, bID); err != nil {
			return ErrMsg{Err: err}
		}
		return m.reloadTreeMsg()
	}
}

func (m *AppModel) deleteTask(id int64) tea.Cmd {
	return func() tea.Msg {
		if err := m.db.DeleteTask(id); err != nil {
			return ErrMsg{Err: err}
		}
		m.taskPanel.hasTask = false
		m.focus = focusSidebar
		m.updateFocusStyles()
		return m.reloadTreeMsg()
	}
}

func (m *AppModel) deleteFolder(id int64) tea.Cmd {
	return func() tea.Msg {
		if err := m.db.DeleteFolder(id); err != nil {
			return ErrMsg{Err: err}
		}
		return m.reloadTreeMsg()
	}
}

func (m *AppModel) reloadTreeMsg() tea.Msg {
	folders, err := m.db.GetAllFolders()
	if err != nil {
		return ErrMsg{Err: err}
	}
	incomplete, err := m.db.GetTasksByCompletion(false)
	if err != nil {
		return ErrMsg{Err: err}
	}
	completed, err := m.db.GetTasksByCompletion(true)
	if err != nil {
		return ErrMsg{Err: err}
	}
	return TreeLoadedMsg{Folders: folders, Tasks: append(incomplete, completed...)}
}

func (m AppModel) View() string {
	if m.width == 0 {
		return "Loading..."
	}

	if m.helpModal.visible {
		return m.helpModal.View(m.width, m.height)
	}

	if m.quickOpen.visible {
		return m.quickOpen.View(m.width, m.height)
	}

	if m.sidebarOpen {
		return lipgloss.JoinHorizontal(lipgloss.Top, m.sidebar.View(), m.taskPanel.View())
	}
	return m.taskPanel.View()
}

