package main

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/charmbracelet/lipgloss"
)

// Config is the plugin's user configuration, read from config.toml in the herdr
// plugin config dir (`herdr plugin config-dir agustinvalencia.herdr-jump`).
type Config struct {
	// Align places the picker card in the overlay pane. Combine a horizontal
	// word (left/center/right) and a vertical word (top/middle/bottom), e.g.
	// "center", "top-left", "bottom right", "top". Missing axes default to
	// center.
	Align string `toml:"align"`
	// MaxWidth caps the card width in columns so it reads as a centered card on
	// wide panes. 0 disables the cap (the card grows to fit the pane). Nil in
	// the file means "unset" → the built-in default applies.
	MaxWidth *int `toml:"max_width"`
	// Agents / Spaces control which fields each card shows, mirroring herdr's
	// [ui.sidebar.*] rows model. Empty → the built-in default layout.
	Agents section `toml:"agents"`
	Spaces section `toml:"spaces"`
}

// section is a per-picker layout: a `rows` array where each inner list is one
// line inside the card, listing the field tokens to show (see fields.go).
type section struct {
	Rows [][]string `toml:"rows"`
}

const defaultMaxWidth = 96

// loadConfig reads config.toml (if present) over the defaults. HERDR_JUMP_ALIGN,
// when set, still overrides the align key — handy for quick testing — but
// config.toml is the primary knob.
func loadConfig() Config {
	c := Config{Align: "center"}
	if dir := os.Getenv("HERDR_PLUGIN_CONFIG_DIR"); dir != "" {
		var fc Config
		if _, err := toml.DecodeFile(filepath.Join(dir, "config.toml"), &fc); err == nil {
			if strings.TrimSpace(fc.Align) != "" {
				c.Align = fc.Align
			}
			if fc.MaxWidth != nil {
				c.MaxWidth = fc.MaxWidth
			}
			if len(fc.Agents.Rows) > 0 {
				c.Agents = fc.Agents
			}
			if len(fc.Spaces.Rows) > 0 {
				c.Spaces = fc.Spaces
			}
		}
	}
	if s := strings.TrimSpace(os.Getenv("HERDR_JUMP_ALIGN")); s != "" {
		c.Align = s
	}
	return c
}

// alignment resolves the configured align string to lipgloss positions.
func (c Config) alignment() (lipgloss.Position, lipgloss.Position) {
	return parseAlign(c.Align)
}

// maxWidth returns the card width cap: the configured value when set (0 = no
// cap), otherwise the built-in default.
func (c Config) maxWidth() int {
	if c.MaxWidth == nil {
		return defaultMaxWidth
	}
	return *c.MaxWidth
}

// agentRows / spaceRows return the configured card layout, or the built-in
// default when the file leaves the section empty.
func (c Config) agentRows() [][]string {
	if len(c.Agents.Rows) > 0 {
		return c.Agents.Rows
	}
	return defaultAgentRows
}

func (c Config) spaceRows() [][]string {
	if len(c.Spaces.Rows) > 0 {
		return c.Spaces.Rows
	}
	return defaultSpaceRows
}

// parseAlign turns a free-form spec ("center", "top-left", "bottom right") into
// horizontal and vertical positions; unrecognised or missing axes stay centered.
func parseAlign(spec string) (lipgloss.Position, lipgloss.Position) {
	h, v := lipgloss.Center, lipgloss.Center
	for _, tok := range strings.FieldsFunc(strings.ToLower(spec), func(r rune) bool {
		return r == '-' || r == '_' || r == ' '
	}) {
		switch tok {
		case "left":
			h = lipgloss.Left
		case "right":
			h = lipgloss.Right
		case "top":
			v = lipgloss.Top
		case "bottom":
			v = lipgloss.Bottom
		case "center", "middle", "centre":
			// no-op: the matching axis stays centered
		}
	}
	return h, v
}
