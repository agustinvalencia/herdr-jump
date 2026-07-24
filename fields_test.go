package main

import (
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"
)

// plain strips styling so tests can assert on the text a cell/line resolves to.
func plain(s string) string { return ansi.Strip(s) }

func TestAgentCellResolvesTokens(t *testing.T) {
	a := Agent{
		Agent:                 "claude",
		AgentStatus:           "working",
		Cwd:                   "/home/x/repo",
		PaneID:                "w3:pA",
		TabID:                 "w3:t1",
		TerminalTitle:         "✳ Do the thing",
		TerminalTitleStripped: "Do the thing",
	}
	cases := map[string]string{
		"agent":                   "claude",
		"workspace":               "work",
		"state_text":              "working",
		"terminal_title":          "✳ Do the thing",
		"terminal_title_stripped": "Do the thing",
		"pane":                    "w3:pA",
		"tab":                     "w3:t1",
	}
	for field, want := range cases {
		if got := plain(agentCell(a, "work", field)); got != want {
			t.Errorf("agentCell(%q) = %q, want %q", field, got, want)
		}
	}
	// cwd is home-shortened; an unknown token resolves to nothing.
	if got := plain(agentCell(a, "work", "cwd")); got != "/home/x/repo" && !strings.HasPrefix(got, "~") {
		t.Errorf("cwd cell = %q, want the (possibly ~-shortened) path", got)
	}
	if got := agentCell(a, "work", "nonsense"); got != "" {
		t.Errorf("unknown token = %q, want empty", got)
	}
}

func TestBuildLinesDropsEmptyCellsKeepsRows(t *testing.T) {
	a := Agent{Agent: "claude", AgentStatus: "idle"} // no title, no cwd
	rows := [][]string{
		{"terminal_title_stripped"}, // empty → blank line, but the row is kept
		{"agent", "workspace"},      // workspace empty → just "claude", no separator
	}
	lines := buildLines(rows, func(f string) string { return agentCell(a, "", f) })
	if len(lines) != len(rows) {
		t.Fatalf("got %d lines, want %d (rows kept positionally)", len(lines), len(rows))
	}
	if plain(lines[0]) != "" {
		t.Errorf("empty title row = %q, want blank", plain(lines[0]))
	}
	if plain(lines[1]) != "claude" {
		t.Errorf("row with one empty cell = %q, want just %q (no dangling separator)", plain(lines[1]), "claude")
	}
}

func TestAgentItemsDefaultLayoutIsTitleForward(t *testing.T) {
	a := Agent{
		Agent:                 "claude",
		AgentStatus:           "idle",
		Cwd:                   "/x",
		PaneID:                "w3:pA",
		WorkspaceID:           "w3",
		TerminalTitleStripped: "Do the thing",
	}
	items := agentItems([]Agent{a}, map[string]string{"w3": "work"}, map[string]int{"w3": 3}, defaultAgentRows)
	lines := items[0].lines
	if len(lines) != 3 {
		t.Fatalf("default layout produced %d lines, want 3", len(lines))
	}
	if plain(lines[0]) != "Do the thing" {
		t.Errorf("line 1 = %q, want the terminal title (title-forward)", plain(lines[0]))
	}
	if !strings.Contains(plain(lines[1]), "claude") || !strings.Contains(plain(lines[1]), "work") {
		t.Errorf("line 2 = %q, want agent + workspace", plain(lines[1]))
	}
}

func TestSpaceItemsOrderAndLayout(t *testing.T) {
	spaces := []Workspace{
		{WorkspaceID: "w6", Label: "Notes", Number: 6, AgentStatus: "working", PaneCount: 10, TabCount: 4},
		{WorkspaceID: "w3", Label: "NFM", Number: 3, AgentStatus: "idle", PaneCount: 1, TabCount: 1},
	}
	items := spaceItems(spaces, defaultSpaceRows)
	if items[0].id != "w3" || items[1].id != "w6" {
		t.Fatalf("space order = [%q, %q], want sorted by number [w3, w6]", items[0].id, items[1].id)
	}
	// #3 in the first row, "1 pane · 1 tab" in the second (singular).
	if !strings.Contains(plain(items[0].lines[0]), "#3") {
		t.Errorf("space line 1 = %q, want the #number", plain(items[0].lines[0]))
	}
	if got := plain(items[0].lines[1]); !strings.Contains(got, "1 pane") || !strings.Contains(got, "1 tab") {
		t.Errorf("space line 2 = %q, want singular pane/tab counts", got)
	}
}

func TestConfigRowsDefaultAndOverride(t *testing.T) {
	if got := (Config{}).agentRows(); len(got) != len(defaultAgentRows) {
		t.Errorf("unset agent rows = %v, want default", got)
	}
	custom := [][]string{{"agent"}}
	c := Config{Agents: section{Rows: custom}}
	if got := c.agentRows(); len(got) != 1 || got[0][0] != "agent" {
		t.Errorf("configured agent rows = %v, want the override %v", got, custom)
	}
	if got := (Config{}).spaceRows(); len(got) != len(defaultSpaceRows) {
		t.Errorf("unset space rows = %v, want default", got)
	}
}
