package main

import "github.com/charmbracelet/lipgloss"

// Palette — Catppuccin Mocha with a blue accent, matching the herdr config
// (theme = "catppuccin", accent = "blue").
const (
	colBase    = lipgloss.Color("#1e1e2e")
	colText    = lipgloss.Color("#cdd6f4")
	colSubtext = lipgloss.Color("#a6adc8")
	colOverlay = lipgloss.Color("#6c7086")
	colAccent  = lipgloss.Color("#89b4fa") // blue
	colGreen   = lipgloss.Color("#a6e3a1")
	colYellow  = lipgloss.Color("#f9e2af")
	colRed     = lipgloss.Color("#f38ba8")
	colCrust   = lipgloss.Color("#11111b")
)

var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(colCrust).
			Background(colAccent).
			Padding(0, 1)

	countStyle  = lipgloss.NewStyle().Foreground(colOverlay)
	promptStyle = lipgloss.NewStyle().Foreground(colAccent).Bold(true)

	primaryStyle = lipgloss.NewStyle().Foreground(colText)
	badgeStyle   = lipgloss.NewStyle().Foreground(colSubtext)
	detailStyle  = lipgloss.NewStyle().Foreground(colOverlay)
	footerStyle  = lipgloss.NewStyle().Foreground(colOverlay)
	emptyStyle   = lipgloss.NewStyle().Foreground(colOverlay).Italic(true)

	// Selected row: accent bar on the left, brighter text.
	selBarStyle  = lipgloss.NewStyle().Foreground(colAccent).Bold(true)
	selTextStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#ffffff")).Bold(true)
)

// statusColor maps a herdr agent/workspace status to its glyph colour.
func statusColor(status string) lipgloss.Color {
	switch status {
	case "idle":
		return colGreen
	case "working":
		return colYellow
	case "blocked":
		return colRed
	default:
		return colOverlay
	}
}
