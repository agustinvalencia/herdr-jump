package main

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/sahilm/fuzzy"
)

// item is one selectable row, shared by the agents and spaces pickers.
type item struct {
	id         string          // opaque id handed to the focus command
	glyph      string          // status marker
	glyphColor lipgloss.Color  // colour for the glyph
	primary    string          // main label (agent name / space label)
	badge      string          // secondary tag (space name / "#1")
	detail     string          // trailing context (cwd / pane+tab counts)
	focused    bool            // currently-focused agent/space (shown with a ▸)
	search     string          // concatenated text used for fuzzy filtering
}

// picker is the bubbletea model. It is modal (lazygit-style):
//
//   - nav mode (default): j/k or arrows move; 1-9 jump to and select that item
//     directly; g/G go to top/bottom; enter selects the highlighted one; /
//     enters filter mode; q/esc cancel.
//   - filter mode: typing fuzzy-filters the list; enter selects; esc clears the
//     filter and returns to nav mode.
//
// The mode split is what lets j/k and the number hotkeys coexist with a text
// filter: in nav mode letters are commands, in filter mode they are query text.
type picker struct {
	title     string
	items     []item
	order     []int // indices into items, after filtering
	query     string
	cursor    int
	filtering bool // true while the filter prompt is active
	width     int
	height    int
	alignH    lipgloss.Position // card placement, from config
	alignV    lipgloss.Position
	cardMax   int    // card width cap in columns; 0 = no cap
	chosenID  string // empty when the user cancelled
}

func newPicker(title string, items []item) *picker {
	cfg := loadConfig()
	h, v := cfg.alignment()
	p := &picker{
		title: title, items: items, width: 80, height: 20,
		alignH: h, alignV: v, cardMax: cfg.maxWidth(),
	}
	p.refilter()
	return p
}

func (p *picker) Init() tea.Cmd { return nil }

// searchSource adapts items to the fuzzy.Source interface.
type searchSource []item

func (s searchSource) String(i int) string { return s[i].search }
func (s searchSource) Len() int            { return len(s) }

func (p *picker) refilter() {
	p.order = p.order[:0]
	if strings.TrimSpace(p.query) == "" {
		for i := range p.items {
			p.order = append(p.order, i)
		}
	} else {
		for _, m := range fuzzy.FindFrom(p.query, searchSource(p.items)) {
			p.order = append(p.order, m.Index)
		}
	}
	if p.cursor >= len(p.order) {
		p.cursor = len(p.order) - 1
	}
	if p.cursor < 0 {
		p.cursor = 0
	}
}

func (p *picker) move(delta int) {
	if len(p.order) == 0 {
		return
	}
	p.cursor = (p.cursor + delta + len(p.order)) % len(p.order)
}

// choose records the highlighted item as the selection (a no-op on an empty list).
func (p *picker) choose() {
	if len(p.order) > 0 {
		p.chosenID = p.items[p.order[p.cursor]].id
	}
}

func (p *picker) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		p.width, p.height = msg.Width, msg.Height
	case tea.KeyMsg:
		if p.filtering {
			return p.updateFilter(msg)
		}
		return p.updateNav(msg)
	}
	return p, nil
}

// updateNav handles keys in nav mode: vim motions, number hotkeys, and mode entry.
func (p *picker) updateNav(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c", "q", "esc":
		p.chosenID = ""
		return p, tea.Quit
	case "enter":
		p.choose()
		return p, tea.Quit
	case "j", "down", "ctrl+n":
		p.move(1)
	case "k", "up", "ctrl+p":
		p.move(-1)
	case "g", "home":
		p.cursor = 0
	case "G", "end":
		if len(p.order) > 0 {
			p.cursor = len(p.order) - 1
		}
	case "/":
		p.filtering = true
	default:
		// 1-9 jump straight to and select that visible item.
		if len(msg.Runes) == 1 {
			if r := msg.Runes[0]; r >= '1' && r <= '9' {
				if idx := int(r - '1'); idx < len(p.order) {
					p.cursor = idx
					p.choose()
					return p, tea.Quit
				}
			}
		}
	}
	return p, nil
}

// updateFilter handles keys in filter mode: text entry and selection.
func (p *picker) updateFilter(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c":
		p.chosenID = ""
		return p, tea.Quit
	case "esc":
		// Leave filter mode and restore the full list.
		p.filtering = false
		p.query = ""
		p.refilter()
	case "enter":
		p.choose()
		return p, tea.Quit
	case "up", "ctrl+p":
		p.move(-1)
	case "down", "ctrl+n":
		p.move(1)
	case "backspace":
		if len(p.query) > 0 {
			// Trim one rune, not one byte, to stay UTF-8 safe.
			r := []rune(p.query)
			p.query = string(r[:len(r)-1])
			p.refilter()
		}
	case "space":
		p.query += " "
		p.refilter()
	default:
		if len(msg.Runes) > 0 {
			p.query += string(msg.Runes)
			p.refilter()
		}
	}
	return p, nil
}

func (p *picker) View() string {
	var b strings.Builder

	// Header: title chip + match count.
	b.WriteString(titleStyle.Render(p.title))
	b.WriteString("  ")
	b.WriteString(countStyle.Render(itoa(len(p.order)) + "/" + itoa(len(p.items))))
	b.WriteString("\n\n")

	// Filter line — only shown while filtering.
	if p.filtering {
		b.WriteString(promptStyle.Render("/") + " " + p.query + promptStyle.Render("▏"))
		b.WriteString("\n\n")
	}

	if len(p.order) == 0 {
		msg := "No matches."
		if len(p.items) == 0 {
			msg = "Nothing to show."
		}
		b.WriteString("  " + emptyStyle.Render(msg) + "\n")
	}

	// Windowed list so long lists scroll around the cursor. Reserve extra rows
	// for the filter line when it is showing.
	reserve := 6
	if p.filtering {
		reserve = 8
	}
	visible := p.height - reserve
	if visible < 1 {
		visible = 1
	}
	start := 0
	if p.cursor >= visible {
		start = p.cursor - visible + 1
	}
	end := start + visible
	if end > len(p.order) {
		end = len(p.order)
	}

	// Card width: cap it so the content reads as a centered card on a wide pane
	// rather than sprawling full-width. Column widths derive from the card, not
	// the pane, leaving room for the leading number column.
	cardW := p.width - 4
	if p.cardMax > 0 && cardW > p.cardMax {
		cardW = p.cardMax
	}
	if cardW < 24 {
		cardW = p.width
	}
	primaryW, badgeW := 14, 12
	detailW := cardW - primaryW - badgeW - 12
	if detailW < 10 {
		detailW = 10
	}

	for i := start; i < end; i++ {
		it := p.items[p.order[i]]
		selected := i == p.cursor

		glyph := lipgloss.NewStyle().Foreground(it.glyphColor).Render(it.glyph)
		marker := " "
		if it.focused {
			marker = lipgloss.NewStyle().Foreground(colAccent).Render("▸")
		}

		// Right-aligned 1-based position. 1-9 double as select hotkeys.
		numStr := itoa(i + 1)
		if len(numStr) < 2 {
			numStr = " " + numStr
		}

		primary := pad(it.primary, primaryW)
		badge := pad(it.badge, badgeW)
		detail := truncate(it.detail, detailW)

		var prefix, num, primaryR string
		if selected {
			prefix = selBarStyle.Render("┃")
			num = selTextStyle.Render(numStr)
			primaryR = selTextStyle.Render(primary)
		} else {
			prefix = " "
			num = countStyle.Render(numStr)
			primaryR = primaryStyle.Render(primary)
		}
		row := primaryR + " " + badgeStyle.Render(badge) + " " + detailStyle.Render(detail)
		b.WriteString(prefix + num + " " + marker + " " + glyph + " " + row + "\n")
	}

	b.WriteString("\n")
	if p.filtering {
		b.WriteString(footerStyle.Render("  type to filter · ↑/↓ move · enter focus · esc clear"))
	} else {
		b.WriteString(footerStyle.Render("  1-9 select · j/k move · enter focus · / filter · q quit"))
	}

	// Pad every line to the card width first, so the block centers as one
	// left-aligned unit — otherwise lipgloss.Place centers each line by its own
	// width and the left edge goes ragged. Then place the card within the pane
	// per the configured alignment (see config.go).
	card := lipgloss.NewStyle().Width(cardW).Align(lipgloss.Left).Render(b.String())
	return lipgloss.Place(p.width, p.height, p.alignH, p.alignV, card)
}

// runPicker runs a picker to completion inside the overlay pane and returns the
// chosen id (empty if cancelled).
func runPicker(title string, items []item) string {
	model, err := tea.NewProgram(newPicker(title, items), tea.WithAltScreen()).Run()
	if err != nil {
		errExit("picker failed:", err)
	}
	return model.(*picker).chosenID
}
