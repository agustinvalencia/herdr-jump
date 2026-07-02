package main

import "sort"

// runAgentsUI renders the agents overlay: every detected agent, grouped-ish by
// space and coloured by status, and focuses the one you pick. herdr runs this
// inside the overlay pane opened for the "agents-picker" entry point.
func runAgentsUI() {
	agents, err := listAgents()
	if err != nil {
		errExit(err)
	}
	labels := workspaceLabels()

	// Sort by space label, then by status priority (blocked → working → idle),
	// so the agents that want attention float to the top of each space.
	sort.SliceStable(agents, func(i, j int) bool {
		li, lj := labels[agents[i].WorkspaceID], labels[agents[j].WorkspaceID]
		if li != lj {
			return li < lj
		}
		return statusRank(agents[i].AgentStatus) < statusRank(agents[j].AgentStatus)
	})

	items := make([]item, 0, len(agents))
	for _, a := range agents {
		label := labels[a.WorkspaceID]
		items = append(items, item{
			id:         a.TerminalID,
			glyph:      "●",
			glyphColor: statusColor(a.AgentStatus),
			primary:    a.Agent,
			badge:      label,
			detail:     shortenPath(a.Cwd),
			focused:    a.Focused,
			search:     a.Agent + " " + label + " " + a.AgentStatus + " " + a.Cwd,
		})
	}

	if id := runPicker("Agents", items); id != "" {
		if err := focusAgent(id); err != nil {
			errExit("could not focus agent:", err)
		}
	}
}

// statusRank orders statuses so the ones needing attention sort first:
// blocked (stuck) and done (finished, awaiting you) float above active/idle work.
func statusRank(status string) int {
	switch status {
	case "blocked":
		return 0
	case "done":
		return 1
	case "working":
		return 2
	case "idle":
		return 3
	default:
		return 4
	}
}
