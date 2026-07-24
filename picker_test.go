package main

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// runes builds a KeyMsg for typed characters (letters, digits, "/").
func runes(s string) tea.KeyMsg { return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(s)} }

func sampleItems() []item {
	return []item{
		{id: "a", status: "idle", lines: []string{"alpha"}, search: "alpha"},
		{id: "b", status: "working", lines: []string{"bravo"}, search: "bravo"},
		{id: "c", status: "blocked", lines: []string{"charlie"}, search: "charlie"},
	}
}

// send feeds a key to the model and returns it back as *picker.
func send(p *picker, msg tea.Msg) *picker {
	m, _ := p.Update(msg)
	return m.(*picker)
}

func TestNavMotion(t *testing.T) {
	p := newPicker("T", sampleItems())
	if p.cursor != 0 {
		t.Fatalf("initial cursor = %d, want 0", p.cursor)
	}
	send(p, runes("j"))
	if p.cursor != 1 {
		t.Fatalf("after j, cursor = %d, want 1", p.cursor)
	}
	send(p, runes("k"))
	if p.cursor != 0 {
		t.Fatalf("after k, cursor = %d, want 0", p.cursor)
	}
	send(p, runes("G"))
	if p.cursor != 2 {
		t.Fatalf("after G, cursor = %d, want 2 (bottom)", p.cursor)
	}
	send(p, runes("g"))
	if p.cursor != 0 {
		t.Fatalf("after g, cursor = %d, want 0 (top)", p.cursor)
	}
}

func TestNumberSelectsDirectly(t *testing.T) {
	p := send(newPicker("T", sampleItems()), runes("3"))
	if p.chosenID != "c" {
		t.Fatalf("pressing 3 chose %q, want c", p.chosenID)
	}
}

func TestOutOfRangeNumberIgnored(t *testing.T) {
	p := send(newPicker("T", sampleItems()), runes("9"))
	if p.chosenID != "" {
		t.Fatalf("pressing 9 on a 3-item list chose %q, want none", p.chosenID)
	}
}

func TestQuitCancels(t *testing.T) {
	p := send(newPicker("T", sampleItems()), runes("q"))
	if p.chosenID != "" {
		t.Fatalf("q chose %q, want none", p.chosenID)
	}
}

func TestFilterModeNarrowsAndSelects(t *testing.T) {
	p := newPicker("T", sampleItems())
	send(p, runes("/"))
	if !p.filtering {
		t.Fatal("'/' did not enter filter mode")
	}
	send(p, runes("b"))
	send(p, runes("r"))
	send(p, runes("a"))
	if len(p.order) != 1 {
		t.Fatalf("filter 'bra' matched %d items, want 1", len(p.order))
	}
	p = send(p, tea.KeyMsg{Type: tea.KeyEnter})
	if p.chosenID != "b" {
		t.Fatalf("enter after filter chose %q, want b", p.chosenID)
	}
}

func TestDigitsTypeInFilterMode(t *testing.T) {
	p := newPicker("T", sampleItems())
	send(p, runes("/"))
	send(p, runes("2"))
	if p.chosenID != "" {
		t.Fatalf("digit in filter mode selected %q, want none", p.chosenID)
	}
	if p.query != "2" {
		t.Fatalf("digit in filter mode gave query %q, want \"2\"", p.query)
	}
}

func TestEscLeavesFilterMode(t *testing.T) {
	p := newPicker("T", sampleItems())
	send(p, runes("/"))
	send(p, runes("z")) // no matches
	send(p, tea.KeyMsg{Type: tea.KeyEsc})
	if p.filtering {
		t.Fatal("esc did not leave filter mode")
	}
	if p.query != "" || len(p.order) != 3 {
		t.Fatalf("esc did not restore full list: query=%q order=%d", p.query, len(p.order))
	}
}

func TestParseAlign(t *testing.T) {
	cases := map[string][2]lipgloss.Position{
		"center":       {lipgloss.Center, lipgloss.Center},
		"top-left":     {lipgloss.Left, lipgloss.Top},
		"bottom right": {lipgloss.Right, lipgloss.Bottom},
		"top":          {lipgloss.Center, lipgloss.Top},
		"":             {lipgloss.Center, lipgloss.Center}, // empty → default center
	}
	for spec, want := range cases {
		h, v := parseAlign(spec)
		if h != want[0] || v != want[1] {
			t.Errorf("align %q = (%v,%v), want (%v,%v)", spec, h, v, want[0], want[1])
		}
	}
}

func TestConfigMaxWidthDefaultAndOverride(t *testing.T) {
	if got := (Config{}).maxWidth(); got != defaultMaxWidth {
		t.Fatalf("unset max_width = %d, want default %d", got, defaultMaxWidth)
	}
	zero := 0
	if got := (Config{MaxWidth: &zero}).maxWidth(); got != 0 {
		t.Fatalf("max_width=0 = %d, want 0 (no cap)", got)
	}
}

func TestViewRendersNumbersNoPanic(t *testing.T) {
	v := newPicker("Agents", sampleItems()).View()
	if !strings.Contains(v, "alpha") || !strings.Contains(v, "1") {
		t.Fatalf("view missing numbered items:\n%s", v)
	}
}

// TestAgentItemsFocusTargetIsPaneID guards the herdr API contract: the id we hand
// to `herdr agent focus` must be the pane_id, not the terminal_id (which herdr
// rejects with agent_not_found). Regression test for issue #1.
func TestAgentItemsFocusTargetIsPaneID(t *testing.T) {
	agents := []Agent{
		{Agent: "claude", PaneID: "w3:pA", TerminalID: "term_abc", WorkspaceID: "w3"},
	}
	labels := map[string]string{"w3": "work"}
	order := map[string]int{"w3": 3}

	items := agentItems(agents, labels, order, defaultAgentRows)
	if len(items) != 1 {
		t.Fatalf("got %d items, want 1", len(items))
	}
	if items[0].id != "w3:pA" {
		t.Fatalf("focus target id = %q, want the pane_id %q", items[0].id, "w3:pA")
	}
}

// TestAgentItemsGroupedByWorkspaceOrder keeps the list in herdr's sidebar order:
// grouped by space (workspace number), native tab order preserved within a space.
func TestAgentItemsGroupedByWorkspaceOrder(t *testing.T) {
	agents := []Agent{
		{Agent: "a", PaneID: "w6:p1", WorkspaceID: "w6"},
		{Agent: "b", PaneID: "w3:p1", WorkspaceID: "w3"},
		{Agent: "c", PaneID: "w6:p2", WorkspaceID: "w6"},
	}
	order := map[string]int{"w3": 3, "w6": 6}

	items := agentItems(agents, map[string]string{}, order, defaultAgentRows)
	got := []string{items[0].id, items[1].id, items[2].id}
	want := []string{"w3:p1", "w6:p1", "w6:p2"} // w3 first; w6 pair keeps input order
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("order = %v, want %v", got, want)
		}
	}
}
