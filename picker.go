package main

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
	"github.com/sahilm/fuzzy"
)

// cardGap is the blank-line gap stacked between cards.
const cardGap = 1

// item is one selectable card, shared by the agents and spaces pickers.
type item struct {
	id      string   // opaque id handed to the focus command (pane_id / workspace_id)
	status  string   // raw agent/workspace status → header state icon + text
	lines   []string // prerendered card body lines, one per configured row
	focused bool     // currently-focused agent/space (marked ▸ in the header)
	search  string   // concatenated text used for fuzzy filtering
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

	// Card geometry. cardW is the outer card (border included); a border (1) plus
	// horizontal padding (1) on each side leaves contentW columns for text.
	cardW := p.width - 4
	if p.cardMax > 0 && cardW > p.cardMax {
		cardW = p.cardMax
	}
	if cardW < 24 {
		cardW = p.width - 2
	}
	contentW := cardW - 4
	if contentW < 8 {
		contentW = 8
	}

	// Window the cards so a long list scrolls around the cursor. Each card is the
	// two borders + a header line + one line per configured body row.
	cardH := p.cardBodyLines() + 3
	slotH := cardH + cardGap
	reserve := 4 // title line, blank, footer, and a margin row
	if p.filtering {
		reserve += 2
	}
	avail := p.height - reserve
	if avail < cardH {
		avail = cardH
	}
	perPage := (avail + cardGap) / slotH // the last card needs no trailing gap
	if perPage < 1 {
		perPage = 1
	}
	start := 0
	if p.cursor >= perPage {
		start = p.cursor - perPage + 1
	}
	end := start + perPage
	if end > len(p.order) {
		end = len(p.order)
	}

	for i := start; i < end; i++ {
		if i > start {
			b.WriteString(strings.Repeat("\n", cardGap))
		}
		b.WriteString(p.renderCard(p.items[p.order[i]], i+1, i == p.cursor, contentW))
		b.WriteString("\n")
	}

	b.WriteString("\n")
	if p.filtering {
		b.WriteString(footerStyle.Render("type to filter · ↑/↓ move · enter focus · esc clear"))
	} else {
		b.WriteString(footerStyle.Render("1-9 select · j/k move · enter focus · / filter · q quit"))
	}

	// Pad every line to the card width so the block places as one left-aligned
	// unit (otherwise lipgloss.Place ragged-centers each line by its own width),
	// then place it within the pane per the configured alignment (see config.go).
	block := lipgloss.NewStyle().Width(cardW).Align(lipgloss.Left).Render(b.String())
	return lipgloss.Place(p.width, p.height, p.alignH, p.alignV, block)
}

// cardBodyLines is the body-row count, uniform across a picker's cards (they all
// share the same configured layout).
func (p *picker) cardBodyLines() int {
	for _, it := range p.items {
		return len(it.lines)
	}
	return 0
}

// renderCard draws one agent/space as a bordered card: a header line (hotkey
// number on the left, focus marker + status on the right) above the prerendered
// body lines. The selected card gets a thick accent border, others a dim rounded
// one. Lines are ansi-truncated so styling survives the cut.
func (p *picker) renderCard(it item, num int, selected bool, contentW int) string {
	numCell := countStyle.Render(itoa(num))
	if selected {
		numCell = selTextStyle.Render(itoa(num))
	}
	right := stateIcon(it.status)
	if st := stateText(it.status); st != "" {
		right += " " + st
	}
	if it.focused {
		right = lipgloss.NewStyle().Foreground(colAccent).Render("▸") + " " + right
	}
	gap := contentW - ansi.StringWidth(numCell) - ansi.StringWidth(right)
	if gap < 1 {
		gap = 1
	}
	header := numCell + strings.Repeat(" ", gap) + right

	// Fit every line to exactly contentW columns so the border sizes uniformly
	// and no line word-wraps. We size the border to the content ourselves rather
	// than via Style.Width, which would subtract padding and re-wrap.
	lines := make([]string, 0, len(it.lines)+1)
	lines = append(lines, fitWidth(header, contentW))
	for _, bl := range it.lines {
		lines = append(lines, fitWidth(bl, contentW))
	}

	border, bc := lipgloss.RoundedBorder(), colOverlay
	if selected {
		border, bc = lipgloss.ThickBorder(), colAccent
	}
	return lipgloss.NewStyle().
		Border(border).
		BorderForeground(bc).
		Padding(0, 1).
		Render(strings.Join(lines, "\n"))
}

// fitWidth truncates (with an ellipsis) or right-pads a possibly-styled string to
// exactly w display columns, counting cells not bytes so ANSI styling survives.
func fitWidth(s string, w int) string {
	s = ansi.Truncate(s, w, "…")
	if pad := w - ansi.StringWidth(s); pad > 0 {
		s += strings.Repeat(" ", pad)
	}
	return s
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
