package main

import "fmt"

// runPreview renders the picker View at a forced size (WIDTHxHEIGHT, default
// 120x40) so the layout and alignment can be eyeballed without an overlay pane.
func runPreview(args []string) {
	w, h := 120, 40
	if len(args) > 0 {
		fmt.Sscanf(args[0], "%dx%d", &w, &h)
	}
	items := []item{
		{id: "1", glyph: "●", glyphColor: colGreen, primary: "claude", badge: "NFM", detail: shortenPath("/Users/eaguval/repositories/work/foundation-model/network-foundation-model"), focused: true, search: "claude NFM"},
		{id: "2", glyph: "●", glyphColor: colYellow, primary: "claude", badge: "cuaderno", detail: "~/dotfiles/.config", search: "claude cuaderno"},
		{id: "3", glyph: "●", glyphColor: colRed, primary: "claude", badge: "cuaderno", detail: "~/repositories/personal/workflow/cuaderno", search: "claude cuaderno"},
	}
	p := newPicker("Agents", items)
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
		fmt.Printf("  %-8s %-10s %-10s %s  [%s]\n",
			a.AgentStatus, a.Agent, labels[a.WorkspaceID], shortenPath(a.Cwd), a.TerminalID)
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
