package ui

import (
	"strings"
	"time"

	"github.com/CalebLewallen/godo/internal/db"
	"github.com/CalebLewallen/godo/internal/model"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type taskField int

const (
	fieldName taskField = iota
	fieldDueDate
	fieldFolder
	fieldDescription
	fieldCount
)

type closePromptChoice int

const (
	choiceNone    closePromptChoice = iota
	choiceSave
	choiceDiscard
	choiceCancel
)

// TaskPanelModel shows and edits a single task.
type TaskPanelModel struct {
	focused      bool
	width        int
	height       int
	db           *db.DB
	task         model.Task
	hasTask      bool
	activeField  taskField
	dirty     bool
	showClose bool
	nameInput    textinput.Model
	dueDateInput textinput.Model
	folderInput  textinput.Model
	descTA       textarea.Model
	folders      []model.Folder
	errMsg       string
}

func NewTaskPanelModel(database *db.DB) TaskPanelModel {
	name := textinput.New()
	name.Placeholder = "Task name..."
	name.CharLimit = 200

	due := textinput.New()
	due.Placeholder = "YYYY-MM-DD"
	due.CharLimit = 20

	folder := textinput.New()
	folder.Placeholder = "Folder name (optional)"
	folder.CharLimit = 100

	desc := textarea.New()
	desc.Placeholder = "Description (markdown supported)..."
	desc.SetWidth(40)
	desc.SetHeight(10)
	desc.ShowLineNumbers = false

	return TaskPanelModel{
		db:           database,
		nameInput:    name,
		dueDateInput: due,
		folderInput:  folder,
		descTA:       desc,
	}
}

func (m TaskPanelModel) Init() tea.Cmd {
	return nil
}

func (m TaskPanelModel) Update(msg tea.Msg) (TaskPanelModel, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.showClose {
			switch msg.String() {
			case "s":
				m.showClose = false
				return m, m.saveTask()
			case "d":
				m.showClose = false
				m.dirty = false
				return m, func() tea.Msg { return TaskClosedMsg{} }
			case "c":
				m.showClose = false
			}
			return m, nil
		}

		switch {
		case key.Matches(msg, TaskBindings.OpenLinks):
			for _, u := range extractURLs(m.descTA.Value()) {
				openURL(u)
			}
			return m, nil

		case key.Matches(msg, TaskBindings.MarkDone):
			if m.hasTask {
				m.task.Completed = !m.task.Completed
				if !m.task.Completed {
					m.task.CompletedAt = time.Time{}
				}
				return m, m.saveTask()
			}
			return m, nil


		case key.Matches(msg, TaskBindings.Save):
			return m, m.saveTask()

		case key.Matches(msg, TaskBindings.Close):
			if m.dirty {
				m.showClose = true
			} else {
				return m, func() tea.Msg { return TaskClosedMsg{} }
			}

		case key.Matches(msg, TaskBindings.NextField):
			m.blurCurrentField()
			m.activeField = (m.activeField + 1) % fieldCount
			return m, m.focusCurrentField()

		case key.Matches(msg, TaskBindings.PrevField):
			m.blurCurrentField()
			m.activeField = (m.activeField - 1 + fieldCount) % fieldCount
			return m, m.focusCurrentField()

		default:
			var cmd tea.Cmd
			m, cmd = m.updateActiveField(msg)
			cmds = append(cmds, cmd)
		}

	case OpenTaskMsg:
		t, err := m.db.GetTask(msg.TaskID)
		if err != nil {
			m.errMsg = err.Error()
			return m, nil
		}
		m.loadTask(t)
		return m, m.focusCurrentField()

	case NewTaskMsg:
		t, err := m.db.CreateTask("New Task", msg.FolderID)
		if err != nil {
			m.errMsg = err.Error()
			return m, nil
		}
		m.loadTask(t)
		m.dirty = true
		return m, m.focusCurrentField()

	case TreeLoadedMsg:
		m.folders = msg.Folders

	case TaskSavedMsg:
		// already handled locally

	}

	return m, tea.Batch(cmds...)
}

func (m *TaskPanelModel) loadTask(t model.Task) {
	m.task = t
	m.hasTask = true
	m.dirty = false
	m.showClose = false
	m.errMsg = ""
	m.activeField = fieldName

	m.nameInput.SetValue(t.Name)
	m.dueDateInput.SetValue(t.DueDate)
	// Folder name lookup
	folderName := ""
	for _, f := range m.folders {
		if t.FolderID != nil && f.ID == *t.FolderID {
			folderName = f.Name
			break
		}
	}
	m.folderInput.SetValue(folderName)
	m.descTA.SetValue(t.Description)
}

func (m *TaskPanelModel) saveTask() tea.Cmd {
	// Resolve folder ID from name
	var folderID *int64
	folderName := strings.TrimSpace(m.folderInput.Value())
	if folderName != "" {
		for _, f := range m.folders {
			if strings.EqualFold(f.Name, folderName) {
				id := f.ID
				folderID = &id
				break
			}
		}
	}

	m.task.Name = strings.TrimSpace(m.nameInput.Value())
	if m.task.Name == "" {
		m.task.Name = "Untitled"
		m.nameInput.SetValue(m.task.Name)
	}
	m.task.DueDate = strings.TrimSpace(m.dueDateInput.Value())
	m.task.FolderID = folderID
	m.task.Description = m.descTA.Value()

	if err := m.db.SaveTask(m.task); err != nil {
		m.errMsg = err.Error()
		return nil
	}
	m.dirty = false
	m.errMsg = ""
	t := m.task
	return func() tea.Msg { return TaskSavedMsg{Task: t} }
}

func (m *TaskPanelModel) blurCurrentField() {
	switch m.activeField {
	case fieldName:
		m.nameInput.Blur()
	case fieldDueDate:
		m.dueDateInput.Blur()
	case fieldFolder:
		m.folderInput.Blur()
	case fieldDescription:
		m.descTA.Blur()
	}
}

func (m *TaskPanelModel) focusCurrentField() tea.Cmd {
	switch m.activeField {
	case fieldName:
		return m.nameInput.Focus()
	case fieldDueDate:
		return m.dueDateInput.Focus()
	case fieldFolder:
		return m.folderInput.Focus()
	case fieldDescription:
		return m.descTA.Focus()
	}
	return nil
}

func (m TaskPanelModel) updateActiveField(msg tea.KeyMsg) (TaskPanelModel, tea.Cmd) {
	var cmd tea.Cmd
	prevDesc := m.descTA.Value()

	switch m.activeField {
	case fieldName:
		m.nameInput, cmd = m.nameInput.Update(msg)
		if m.nameInput.Value() != m.task.Name {
			m.dirty = true
		}
	case fieldDueDate:
		m.dueDateInput, cmd = m.dueDateInput.Update(msg)
		if m.dueDateInput.Value() != m.task.DueDate {
			m.dirty = true
		}
	case fieldFolder:
		m.folderInput, cmd = m.folderInput.Update(msg)
		m.dirty = true
	case fieldDescription:
		m.descTA, cmd = m.descTA.Update(msg)
		if m.descTA.Value() != prevDesc {
			m.dirty = true
		}
	}
	return m, cmd
}


func (m TaskPanelModel) View() string {
	if !m.hasTask {
		style := panelStyle.Width(m.width - 2).Height(m.height - 2)
		if m.focused {
			style = panelFocusedStyle.Width(m.width - 2).Height(m.height - 2)
		}
		help := helpStyle.Render("?  help")
		return style.Render(
			titleStyle.Render("GoDo") + "\n\n" +
				mutedStyle.Render("Select a task or press ctrl+n to create one.") + "\n\n" +
				help,
		)
	}

	innerWidth := m.width - 4

	var sb strings.Builder

	// Title bar
	dirtyMark := ""
	if m.dirty {
		dirtyMark = " " + errorStyle.Render("●")
	}
	sb.WriteString(titleStyle.Render(m.task.Name) + dirtyMark + "\n")
	sb.WriteString(dividerStyle.Render(strings.Repeat("─", innerWidth)) + "\n")

	// Close prompt overlay
	if m.showClose {
		sb.WriteString("\n")
		sb.WriteString(errorStyle.Render("Unsaved changes: "))
		sb.WriteString("(s) save  (d) discard  (c) cancel\n")
		style := panelStyle.Width(m.width - 2).Height(m.height - 2)
		if m.focused {
			style = panelFocusedStyle.Width(m.width - 2).Height(m.height - 2)
		}
		return style.Render(sb.String())
	}

	// Error
	if m.errMsg != "" {
		sb.WriteString(errorStyle.Render("Error: "+m.errMsg) + "\n")
	}

	// Status
	sb.WriteString(fieldLabelStyle.Render("Status: "))
	if m.task.Completed {
		sb.WriteString(doneStyle.Render("✓ Completed") + "\n\n")
	} else {
		sb.WriteString(mutedStyle.Render("○ Incomplete") + "\n\n")
	}

	// Field: Name
	label := fieldLabelStyle
	if m.activeField == fieldName {
		label = fieldLabelFocusedStyle
	}
	sb.WriteString(label.Render("Name: "))
	sb.WriteString(m.nameInput.View() + "\n\n")

	// Field: Due Date
	label = fieldLabelStyle
	if m.activeField == fieldDueDate {
		label = fieldLabelFocusedStyle
	}
	sb.WriteString(label.Render("Due Date: "))
	sb.WriteString(m.dueDateInput.View() + "\n\n")

	// Field: Folder
	label = fieldLabelStyle
	if m.activeField == fieldFolder {
		label = fieldLabelFocusedStyle
	}
	sb.WriteString(label.Render("Folder: "))
	sb.WriteString(m.folderInput.View() + "\n\n")

	// Field: Description
	label = fieldLabelStyle
	if m.activeField == fieldDescription {
		label = fieldLabelFocusedStyle
	}
	sb.WriteString(label.Render("Description") + "\n")
	sb.WriteString(dividerStyle.Render(strings.Repeat("─", innerWidth)) + "\n")

	usedLines := 17 // title, divider, status, name, due date, folder, desc label, dividers, help
	descHeight := m.height - usedLines
	if descHeight < 3 {
		descHeight = 3
	}

	if m.activeField == fieldDescription {
		// Edit mode: show the textarea with cursor.
		m.descTA.SetWidth(innerWidth)
		m.descTA.SetHeight(descHeight)
		sb.WriteString(m.descTA.View() + "\n")
	} else {
		// Display mode: render with OSC 8 links, bypassing lipgloss wrapping
		// which can corrupt non-SGR escape sequences.
		raw := wrapLinksOSC8(m.descTA.Value())
		lines := strings.Split(raw, "\n")
		if len(lines) > descHeight {
			lines = lines[:descHeight]
		}
		sb.WriteString(strings.Join(lines, "\n") + "\n")
	}

	// Help
	sb.WriteString(dividerStyle.Render(strings.Repeat("─", innerWidth)) + "\n")
	sb.WriteString(helpStyle.Render("?  help"))

	style := panelStyle.Width(m.width - 2).Height(m.height - 2)
	if m.focused {
		style = panelFocusedStyle.Width(m.width - 2).Height(m.height - 2)
	}
	return style.Render(sb.String())
}
