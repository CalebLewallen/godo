package ui

import "github.com/charmbracelet/bubbles/key"

// GlobalKeys holds application-wide key bindings.
type GlobalKeys struct {
	ShowHelp      key.Binding
	ToggleFocus   key.Binding
	ToggleSidebar key.Binding
	Filter        key.Binding
	Quit          key.Binding
	NewTask       key.Binding
	NewFolder     key.Binding
	QuickOpen     key.Binding
}

var Keys = GlobalKeys{
	ShowHelp: key.NewBinding(
		key.WithKeys("?"),
		key.WithHelp("?", "help"),
	),
	ToggleFocus: key.NewBinding(
		key.WithKeys("ctrl+e"),
		key.WithHelp("ctrl+e", "toggle focus"),
	),
	ToggleSidebar: key.NewBinding(
		key.WithKeys("ctrl+b"),
		key.WithHelp("ctrl+b", "toggle sidebar"),
	),
	Filter: key.NewBinding(
		key.WithKeys("ctrl+f"),
		key.WithHelp("ctrl+f", "filter"),
	),
	Quit: key.NewBinding(
		key.WithKeys("ctrl+q"),
		key.WithHelp("ctrl+q", "quit"),
	),
	NewTask: key.NewBinding(
		key.WithKeys("ctrl+n"),
		key.WithHelp("ctrl+n", "new task"),
	),
	NewFolder: key.NewBinding(
		key.WithKeys("alt+ctrl+n"),
		key.WithHelp("alt+ctrl+n", "new folder"),
	),
	QuickOpen: key.NewBinding(
		key.WithKeys("ctrl+p"),
		key.WithHelp("ctrl+p", "quick open task"),
	),
}

// SidebarKeys holds sidebar-specific bindings.
type SidebarKeys struct {
	NextTab      key.Binding
	Up           key.Binding
	Down         key.Binding
	Expand       key.Binding
	Collapse     key.Binding
	Enter        key.Binding
	IndentFolder key.Binding
	DedentFolder key.Binding
	MoveUp       key.Binding
	MoveDown     key.Binding
	Rename       key.Binding
	ExpandAll    key.Binding
	CollapseAll  key.Binding
	Delete       key.Binding
}

var SidebarBindings = SidebarKeys{
	NextTab: key.NewBinding(
		key.WithKeys("shift+tab"),
		key.WithHelp("shift+tab", "next tab"),
	),
	Up: key.NewBinding(
		key.WithKeys("up", "k"),
		key.WithHelp("↑/k", "up"),
	),
	Down: key.NewBinding(
		key.WithKeys("down", "j"),
		key.WithHelp("↓/j", "down"),
	),
	Expand: key.NewBinding(
		key.WithKeys("right", "l"),
		key.WithHelp("→/l", "expand"),
	),
	Collapse: key.NewBinding(
		key.WithKeys("left", "h"),
		key.WithHelp("←/h", "collapse"),
	),
	Enter: key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("enter", "open"),
	),
	IndentFolder: key.NewBinding(
		key.WithKeys("shift+right"),
		key.WithHelp("shift+→", "indent folder"),
	),
	DedentFolder: key.NewBinding(
		key.WithKeys("shift+left"),
		key.WithHelp("shift+←", "dedent folder"),
	),
	MoveUp: key.NewBinding(
		key.WithKeys("shift+up"),
		key.WithHelp("shift+↑", "move up"),
	),
	MoveDown: key.NewBinding(
		key.WithKeys("shift+down"),
		key.WithHelp("shift+↓", "move down"),
	),
	Rename: key.NewBinding(
		key.WithKeys("f2"),
		key.WithHelp("f2", "rename"),
	),
	ExpandAll: key.NewBinding(
		key.WithKeys("ctrl+right"),
		key.WithHelp("ctrl+→", "expand all"),
	),
	CollapseAll: key.NewBinding(
		key.WithKeys("ctrl+left"),
		key.WithHelp("ctrl+←", "collapse all"),
	),
	Delete: key.NewBinding(
		key.WithKeys("ctrl+x"),
		key.WithHelp("ctrl+x", "delete"),
	),
}

// TaskKeys holds task-panel-specific bindings.
type TaskKeys struct {
	NextField   key.Binding
	PrevField   key.Binding
	Save        key.Binding
	Close       key.Binding
	OpenLinks   key.Binding
	MarkDone    key.Binding
}

var TaskBindings = TaskKeys{
	NextField: key.NewBinding(
		key.WithKeys("tab"),
		key.WithHelp("tab", "next field"),
	),
	PrevField: key.NewBinding(
		key.WithKeys("shift+tab"),
		key.WithHelp("shift+tab", "prev field"),
	),
	Save: key.NewBinding(
		key.WithKeys("ctrl+s"),
		key.WithHelp("ctrl+s", "save"),
	),
	Close: key.NewBinding(
		key.WithKeys("ctrl+w"),
		key.WithHelp("ctrl+w", "close"),
	),
	OpenLinks: key.NewBinding(
		key.WithKeys("ctrl+o"),
		key.WithHelp("ctrl+o", "open links"),
	),
	MarkDone: key.NewBinding(
		key.WithKeys("ctrl+d"),
		key.WithHelp("ctrl+d", "mark done"),
	),
}
