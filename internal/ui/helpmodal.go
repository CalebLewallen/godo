package ui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

type helpEntry struct {
	section string // non-empty → render as section header
	key     string
	desc    string
}

var helpEntries = []helpEntry{
	{section: "Global"},
	{key: "ctrl+e", desc: "Toggle focus sidebar ↔ task"},
	{key: "ctrl+b", desc: "Toggle sidebar"},
	{key: "ctrl+f", desc: "Filter sidebar"},
	{key: "ctrl+p", desc: "Quick open task"},
	{key: "ctrl+n", desc: "New task (in active folder)"},
	{key: "alt+ctrl+n", desc: "New folder (prompts for name)"},
	{key: "ctrl+q", desc: "Quit"},
	{key: "?", desc: "Show this help"},

	{section: "Sidebar"},
	{key: "↑ / ↓  (or k / j)", desc: "Navigate"},
	{key: "→ / ←  (or l / h)", desc: "Expand / Collapse folder"},
	{key: "enter", desc: "Open task / Toggle folder"},
	{key: "shift+tab", desc: "Switch tabs (Todo / Completed)"},
	{key: "f2", desc: "Rename selected folder or task"},
	{key: "shift+→", desc: "Indent folder/task (nest under folder above)"},
	{key: "shift+←", desc: "Dedent folder/task (move up one level)"},
	{key: "shift+↑ / ↓", desc: "Reorder folder/task among siblings"},
	{key: "ctrl+→", desc: "Expand all folders"},
	{key: "ctrl+←", desc: "Collapse all folders"},
	{key: "ctrl+x", desc: "Delete selected folder or task"},

	{section: "Task Panel"},
	{key: "tab / shift+tab", desc: "Next / Previous field"},
	{key: "ctrl+s", desc: "Save task"},
	{key: "ctrl+d", desc: "Toggle task done / incomplete"},
	{key: "ctrl+w", desc: "Close task (prompts if unsaved)"},
	{key: "ctrl+o", desc: "Open all links in description"},
}

// HelpModal is an overlay showing all keyboard shortcuts.
type HelpModal struct {
	visible bool
	cursor  int // top visible row index
}

const (
	modalWidth   = 56
	modalVisible = 16 // max content rows shown at once
)

func (h *HelpModal) Open()  { h.visible = true; h.cursor = 0 }
func (h *HelpModal) Close() { h.visible = false }

func (h *HelpModal) ScrollUp() {
	if h.cursor > 0 {
		h.cursor--
	}
}

func (h *HelpModal) ScrollDown() {
	max := len(helpEntries) - modalVisible
	if max < 0 {
		max = 0
	}
	if h.cursor < max {
		h.cursor++
	}
}

func (h HelpModal) View(termW, termH int) string {
	var sb strings.Builder

	end := h.cursor + modalVisible
	if end > len(helpEntries) {
		end = len(helpEntries)
	}

	for _, e := range helpEntries[h.cursor:end] {
		if e.section != "" {
			sb.WriteString(fieldLabelFocusedStyle.Render(e.section) + "\n")
			continue
		}
		key := lipgloss.NewStyle().
			Foreground(colorAccent).
			Width(22).
			Render(e.key)
		desc := mutedStyle.Render(e.desc)
		sb.WriteString("  " + key + desc + "\n")
	}

	// Scroll indicator
	total := len(helpEntries)
	showing := end - h.cursor
	if total > showing {
		pct := 100 * (h.cursor + showing) / total
		sb.WriteString(mutedStyle.Render("─────────────────────────────────────────────\n"))
		sb.WriteString(helpStyle.Render("↑/↓ scroll") + "  " + mutedStyle.Render("esc close") +
			"  " + mutedStyle.Render(strings.Repeat("─", 10)) +
			"  " + mutedStyle.Render(strings.Repeat("█", pct*10/100)+strings.Repeat("░", 10-pct*10/100)) + "\n")
	} else {
		sb.WriteString(mutedStyle.Render("─────────────────────────────────────────────\n"))
		sb.WriteString(helpStyle.Render("esc  close") + "\n")
	}

	box := lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(colorFocusBorder).
		Padding(1, 2).
		Width(modalWidth).
		Render(
			titleStyle.Render("Keyboard Shortcuts") + "\n" +
				dividerStyle.Render(strings.Repeat("─", modalWidth-4)) + "\n" +
				sb.String(),
		)

	return lipgloss.Place(termW, termH, lipgloss.Center, lipgloss.Center, box)
}
