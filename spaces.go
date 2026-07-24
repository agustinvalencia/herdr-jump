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

	items := spaceItems(spaces, loadConfig().spaceRows())

	if id := runPicker("Spaces", items); id != "" {
		if err := focusWorkspace(id); err != nil {
			errExit("could not focus space:", err)
		}
	}
}

// spaceItems builds the cards for the spaces overlay. Kept pure (no herdr calls)
// so the ordering and configured card layout are unit-testable. rows is the
// per-line field layout (see fields.go).
func spaceItems(spaces []Workspace, rows [][]string) []item {
	// Keep the natural workspace numbering (1, 2, …) so the list matches the
	// sidebar and the prefix+1..9 binds.
	sort.SliceStable(spaces, func(i, j int) bool {
		return spaces[i].Number < spaces[j].Number
	})

	items := make([]item, 0, len(spaces))
	for _, w := range spaces {
		items = append(items, item{
			id:      w.WorkspaceID,
			status:  w.AgentStatus,
			focused: w.Focused,
			lines:   buildLines(rows, func(f string) string { return spaceCell(w, f) }),
			search:  fmt.Sprintf("%d %s %s", w.Number, w.Label, w.AgentStatus),
		})
	}
	return items
}
