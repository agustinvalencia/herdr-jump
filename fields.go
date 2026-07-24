package main

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// The field vocabulary mirrors herdr's own sidebar tokens
// (https://herdr.dev/docs/configuration/#ui-and-sidebar), mapped to what the
// herdr CLI actually exposes. Each token resolves to a styled cell; a card row
// is one inner list in the `rows` config, its non-empty cells joined by a dim
// separator. herdr's `branch` / `git_status` are intentionally absent — the CLI
// does not surface them.

// defaultAgentRows is the title-forward layout: what the agent is doing first,
// then who and where, then the working directory.
var defaultAgentRows = [][]string{
	{"terminal_title_stripped"},
	{"agent", "workspace"},
	{"cwd"},
}

// defaultSpaceRows: the space label with its number, then pane/tab counts.
var defaultSpaceRows = [][]string{
	{"workspace", "number"},
	{"panes", "tabs"},
}

// agentCell renders one agent field token to a styled string, or "" when the
// token is unknown or its value is empty (empty cells are dropped from a row).
func agentCell(a Agent, label, field string) string {
	switch field {
	case "state_icon":
		return stateIcon(a.AgentStatus)
	case "state_text":
		return stateText(a.AgentStatus)
	case "agent":
		return renderOrEmpty(primaryStyle, a.Agent)
	case "workspace":
		return renderOrEmpty(badgeStyle, label)
	case "cwd":
		return renderOrEmpty(detailStyle, shortenPath(a.Cwd))
	case "terminal_title":
		return renderOrEmpty(heroStyle, a.TerminalTitle)
	case "terminal_title_stripped":
		return renderOrEmpty(heroStyle, a.TerminalTitleStripped)
	case "pane":
		return renderOrEmpty(detailStyle, a.PaneID)
	case "tab":
		return renderOrEmpty(detailStyle, a.TabID)
	default:
		return ""
	}
}

// spaceCell renders one space field token to a styled string (see agentCell).
func spaceCell(w Workspace, field string) string {
	switch field {
	case "state_icon":
		return stateIcon(w.AgentStatus)
	case "state_text":
		return stateText(w.AgentStatus)
	case "workspace":
		return renderOrEmpty(primaryStyle, w.Label)
	case "number":
		return badgeStyle.Render("#" + itoa(w.Number))
	case "panes":
		return detailStyle.Render(plural(w.PaneCount, "pane"))
	case "tabs":
		return detailStyle.Render(plural(w.TabCount, "tab"))
	default:
		return ""
	}
}

// renderOrEmpty styles s, or returns "" for blank input so it drops out of its row.
func renderOrEmpty(style lipgloss.Style, s string) string {
	if strings.TrimSpace(s) == "" {
		return ""
	}
	return style.Render(s)
}

// buildLines renders each configured row to one prerendered card body line. Rows
// are kept positionally — even when every cell resolves empty — so every card in
// a picker has the same height; empty cells are dropped from within their line.
func buildLines(rows [][]string, cell func(field string) string) []string {
	sep := detailStyle.Render(" · ")
	lines := make([]string, 0, len(rows))
	for _, row := range rows {
		parts := make([]string, 0, len(row))
		for _, f := range row {
			if s := cell(f); s != "" {
				parts = append(parts, s)
			}
		}
		lines = append(lines, strings.Join(parts, sep))
	}
	return lines
}
