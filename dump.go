package main

import "fmt"

// runPreview renders the picker View at a forced size (WIDTHxHEIGHT, default
// 120x40) so the card layout and alignment can be eyeballed without an overlay
// pane. It runs the real agentItems path over sample data.
func runPreview(args []string) {
	w, h := 120, 40
	if len(args) > 0 {
		fmt.Sscanf(args[0], "%dx%d", &w, &h)
	}
	agents := []Agent{
		{Agent: "claude", AgentStatus: "working", PaneID: "w3:pA", WorkspaceID: "w3", Focused: true,
			Cwd: "/Users/eaguval/repositories/work/foundation-model/network-foundation-model", TerminalTitleStripped: "Analyse epic status and plan next work"},
		{Agent: "claude", AgentStatus: "idle", PaneID: "w6:p4", WorkspaceID: "w6",
			Cwd: "/Users/eaguval/repositories/personal/workflow/cuaderno", TerminalTitleStripped: "Redesign cuaderno-app UI"},
		{Agent: "claude", AgentStatus: "blocked", PaneID: "w6:p7", WorkspaceID: "w6",
			Cwd: "/Users/eaguval/Documents/notebook", TerminalTitleStripped: "Schedule end-of-day train booking"},
	}
	labels := map[string]string{"w3": "NFM", "w6": "Notes"}
	order := map[string]int{"w3": 3, "w6": 6}
	p := newPicker("Agents", agentItems(agents, labels, order, defaultAgentRows))
	p.width, p.height = w, h
	fmt.Print(p.View())
}

// runDump prints the agents and spaces the pickers would show, as plain text.
// It exercises the same herdr CLI calls without a TUI, so it works over a plain
// pipe — handy for verifying the integration.
func runDump() {
	labels := workspaceLabels()

	agents, err := listAgents()
	if err != nil {
		errExit(err)
	}
	fmt.Printf("AGENTS (%d)\n", len(agents))
	for _, a := range agents {
		// Bracketed id is the focus target we hand to `herdr agent focus`.
		fmt.Printf("  %-8s %-10s %-10s %s  [%s]\n",
			a.AgentStatus, a.Agent, labels[a.WorkspaceID], shortenPath(a.Cwd), a.PaneID)
	}

	spaces, err := listWorkspaces()
	if err != nil {
		errExit(err)
	}
	fmt.Printf("SPACES (%d)\n", len(spaces))
	for _, w := range spaces {
		fmt.Printf("  #%d %-10s %-8s %s · %s  [%s]\n",
			w.Number, w.Label, w.AgentStatus, plural(w.PaneCount, "pane"), plural(w.TabCount, "tab"), w.WorkspaceID)
	}
}
