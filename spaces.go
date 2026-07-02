package main

import (
	"fmt"
	"sort"
)

// runSpacesUI renders the spaces overlay: every workspace with its status and
// pane/tab counts, and switches to the one you pick. herdr runs this inside the
// overlay pane opened for the "spaces-picker" entry point.
func runSpacesUI() {
	spaces, err := listWorkspaces()
	if err != nil {
		errExit(err)
	}

	// Keep the natural workspace numbering (1, 2, …) so the list matches the
	// sidebar and the prefix+1..9 binds.
	sort.SliceStable(spaces, func(i, j int) bool {
		return spaces[i].Number < spaces[j].Number
	})

	items := make([]item, 0, len(spaces))
	for _, w := range spaces {
		items = append(items, item{
			id:         w.WorkspaceID,
			glyph:      "●",
			glyphColor: statusColor(w.AgentStatus),
			primary:    w.Label,
			badge:      fmt.Sprintf("#%d", w.Number),
			detail:     fmt.Sprintf("%s · %s", plural(w.PaneCount, "pane"), plural(w.TabCount, "tab")),
			focused:    w.Focused,
			search:     fmt.Sprintf("%d %s %s", w.Number, w.Label, w.AgentStatus),
		})
	}

	if id := runPicker("Spaces", items); id != "" {
		if err := focusWorkspace(id); err != nil {
			errExit("could not focus space:", err)
		}
	}
}
