package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// itoa is a tiny strconv.Itoa alias to keep the view code terse.
func itoa(n int) string { return strconv.Itoa(n) }

// shortenPath replaces $HOME with ~ and collapses long middles so a cwd stays
// readable in a narrow column.
func shortenPath(p string) string {
	if home, err := os.UserHomeDir(); err == nil && home != "" {
		if p == home {
			return "~"
		}
		if strings.HasPrefix(p, home+string(filepath.Separator)) {
			p = "~" + p[len(home):]
		}
	}
	return p
}

// plural formats a count with its noun, pluralised naively ("1 pane", "5 panes").
func plural(n int, noun string) string {
	if n == 1 {
		return fmt.Sprintf("%d %s", n, noun)
	}
	return fmt.Sprintf("%d %ss", n, noun)
}
