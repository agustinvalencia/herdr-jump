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
		{id: "a", primary: "alpha", search: "alpha"},
		{id: "b", primary: "bravo", search: "bravo"},
		{id: "c", primary: "charlie", search: "charlie"},
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
