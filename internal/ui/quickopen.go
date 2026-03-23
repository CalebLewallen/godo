package ui

import (
	"strings"

	"github.com/CalebLewallen/godo/internal/model"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const (
	quickOpenWidth   = 60
	quickOpenMaxRows = 8
)

// QuickOpenModal is a VS Code-style task switcher overlay.
type QuickOpenModal struct {
	visible  bool
	input    textinput.Model
	allTasks []model.Task
	results  []model.Task
	cursor   int
}

func newQuickOpenModal() QuickOpenModal {
	ti := textinput.New()
	ti.Placeholder = "Type a task name..."
	ti.CharLimit = 200

	return QuickOpenModal{input: ti}
}

func (q *QuickOpenModal) Open(tasks []model.Task) tea.Cmd {
	q.visible = true
	q.allTasks = tasks
	q.cursor = 0
	q.input.SetValue("")
	q.filter()
	return q.input.Focus()
}

func (q *QuickOpenModal) Close() {
	q.visible = false
	q.input.Blur()
}

func (q *QuickOpenModal) SetTasks(tasks []model.Task) {
	q.allTasks = tasks
	if q.visible {
		q.filter()
	}
}

func (q *QuickOpenModal) filter() {
	query := strings.ToLower(strings.TrimSpace(q.input.Value()))
	q.results = q.results[:0]
	for _, t := range q.allTasks {
		if query == "" || strings.Contains(strings.ToLower(t.Name), query) {
			q.results = append(q.results, t)
		}
	}
	if q.cursor >= len(q.results) {
		q.cursor = max(0, len(q.results)-1)
	}
}

// Update handles key input while the modal is open. Returns the modal, a cmd,
// and an optional OpenTaskMsg (non-nil when the user confirms a selection).
func (q *QuickOpenModal) Update(msg tea.KeyMsg) (tea.Cmd, tea.Msg) {
	switch msg.String() {
	case "esc":
		q.Close()
		return nil, nil

	case "enter":
		if len(q.results) > 0 {
			id := q.results[q.cursor].ID
			q.Close()
			return nil, OpenTaskMsg{TaskID: id}
		}
		return nil, nil

	case "up", "ctrl+p":
		if q.cursor > 0 {
			q.cursor--
		}
		return nil, nil

	case "down", "ctrl+n":
		if q.cursor < len(q.results)-1 {
			q.cursor++
		}
		return nil, nil
	}

	var cmd tea.Cmd
	q.input, cmd = q.input.Update(msg)
	q.filter()
	return cmd, nil
}

func (q QuickOpenModal) View(termW, termH int) string {
	var sb strings.Builder

	// Input row
	sb.WriteString(q.input.View())
	sb.WriteString("\n")
	sb.WriteString(dividerStyle.Render(strings.Repeat("─", quickOpenWidth-4)))
	sb.WriteString("\n")

	if len(q.results) == 0 {
		sb.WriteString(mutedStyle.Render("  no matching tasks"))
		sb.WriteString("\n")
	} else {
		shown := min(quickOpenMaxRows, len(q.results))
		start := 0
		if q.cursor >= shown {
			start = q.cursor - shown + 1
		}
		for i := start; i < start+shown; i++ {
			line := "  " + q.results[i].Name
			if i == q.cursor {
				line = selectedRowStyle.Width(quickOpenWidth - 4).Render(line)
			} else {
				line = taskStyle.Render(line)
			}
			sb.WriteString(line + "\n")
		}
	}

	sb.WriteString(dividerStyle.Render(strings.Repeat("─", quickOpenWidth-4)))
	sb.WriteString("\n")
	sb.WriteString(helpStyle.Render("↑/↓ navigate  enter open  esc close"))

	box := lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(colorFocusBorder).
		Padding(1, 2).
		Width(quickOpenWidth).
		Render(titleStyle.Render("Open Task") + "\n" + sb.String())

	return lipgloss.Place(termW, termH, lipgloss.Center, lipgloss.Top, box,
		lipgloss.WithWhitespaceChars(" "),
		lipgloss.WithWhitespaceForeground(lipgloss.Color("#00000000")),
	)
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
