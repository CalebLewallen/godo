package ui

import "github.com/charmbracelet/lipgloss"

var (
	colorPrimary    = lipgloss.Color("#7C3AED") // violet
	colorAccent     = lipgloss.Color("#A78BFA")
	colorMuted      = lipgloss.Color("#6B7280")
	colorFocusBorder = lipgloss.Color("#7C3AED")
	colorBorder     = lipgloss.Color("#374151")
	colorSelected   = lipgloss.Color("#1E1B4B")
	colorText       = lipgloss.Color("#F9FAFB")
	colorSubtle     = lipgloss.Color("#9CA3AF")
	colorGreen      = lipgloss.Color("#10B981")
	colorYellow     = lipgloss.Color("#F59E0B")

	// Sidebar
	sidebarStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(colorBorder).
			Padding(0, 1)

	sidebarFocusedStyle = lipgloss.NewStyle().
				BorderStyle(lipgloss.RoundedBorder()).
				BorderForeground(colorFocusBorder).
				Padding(0, 1)

	tabStyle = lipgloss.NewStyle().
			Foreground(colorMuted).
			Padding(0, 1)

	activeTabStyle = lipgloss.NewStyle().
			Foreground(colorAccent).
			Bold(true).
			Padding(0, 1).
			BorderStyle(lipgloss.NormalBorder()).
			BorderBottom(true).
			BorderForeground(colorAccent)

	selectedRowStyle = lipgloss.NewStyle().
				Background(colorSelected).
				Foreground(colorText).
				Bold(true)

	folderStyle = lipgloss.NewStyle().Foreground(colorAccent)
	taskStyle   = lipgloss.NewStyle().Foreground(colorText)
	doneStyle   = lipgloss.NewStyle().Foreground(colorGreen)
	mutedStyle  = lipgloss.NewStyle().Foreground(colorMuted)

	// Task panel
	panelStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(colorBorder).
			Padding(0, 1)

	panelFocusedStyle = lipgloss.NewStyle().
				BorderStyle(lipgloss.RoundedBorder()).
				BorderForeground(colorFocusBorder).
				Padding(0, 1)

	fieldLabelStyle = lipgloss.NewStyle().
			Foreground(colorAccent).
			Bold(true)

	fieldLabelFocusedStyle = lipgloss.NewStyle().
				Foreground(colorPrimary).
				Bold(true)

	dividerStyle = lipgloss.NewStyle().
			Foreground(colorBorder)

	helpStyle = lipgloss.NewStyle().
			Foreground(colorMuted).
			Italic(true)

	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#EF4444")).
			Bold(true)

	titleStyle = lipgloss.NewStyle().
			Foreground(colorPrimary).
			Bold(true)

	dueDateStyle = lipgloss.NewStyle().
			Foreground(colorYellow)
)
