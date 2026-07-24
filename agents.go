package main

import "sort"

// runAgentsUI renders the agents overlay: every detected agent, grouped by space
// and coloured by status, and focuses the one you pick. herdr runs this inside
// the overlay pane opened for the "agents-picker" entry point.
func runAgentsUI() {
	agents, err := listAgents()
	if err != nil {
		errExit(err)
	}
	labels, order := workspaceInfo()

	items := agentItems(agents, labels, order)

	if id := runPicker("Agents", items); id != "" {
		if err := focusAgent(id); err != nil {
			errExit("could not focus agent:", err)
		}
	}
}

// agentItems builds the picker rows for the agents overlay. Kept pure (no herdr
// calls) so the ordering and — crucially — the focus-target id are unit-testable.
func agentItems(agents []Agent, labels map[string]string, order map[string]int) []item {
	// Match herdr's agent panel: group by space in workspace order, preserving
	// each space's native (tab) order. A stable sort by workspace number does
	// exactly that — no status-based reordering — so the list mirrors the sidebar.
	sort.SliceStable(agents, func(i, j int) bool {
		return order[agents[i].WorkspaceID] < order[agents[j].WorkspaceID]
	})

	items := make([]item, 0, len(agents))
	for _, a := range agents {
		label := labels[a.WorkspaceID]
		items = append(items, item{
			// pane_id, not terminal_id: herdr resolves `agent focus <target>` by
			// pane (a terminal_id yields agent_not_found). See focusAgent.
			id:         a.PaneID,
			glyph:      "●",
			glyphColor: statusColor(a.AgentStatus),
			primary:    a.Agent,
			badge:      label,
			detail:     shortenPath(a.Cwd),
			focused:    a.Focused,
			search:     a.Agent + " " + label + " " + a.AgentStatus + " " + a.Cwd,
		})
	}
	return items
}
