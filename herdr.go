package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
)

// herdrBin returns the herdr binary to shell out to. herdr injects
// HERDR_BIN_PATH into every plugin command; it is the portable way to call back
// into the CLI. We prefer the CLI wrappers over raw socket JSON because every
// call we need (list/focus for agents and workspaces) has a stable CLI form.
func herdrBin() string {
	if b := os.Getenv("HERDR_BIN_PATH"); b != "" {
		return b
	}
	return "herdr"
}

// Agent is one entry from `herdr agent list`. The command returns every managed
// terminal; entries with an empty Agent are plain shells rather than detected AI
// agents, so callers filter those out.
type Agent struct {
	Agent       string `json:"agent"`
	AgentStatus string `json:"agent_status"`
	Cwd         string `json:"cwd"`
	Focused     bool   `json:"focused"`
	PaneID      string `json:"pane_id"`
	TabID       string `json:"tab_id"`
	TerminalID  string `json:"terminal_id"`
	WorkspaceID string `json:"workspace_id"`
}

// Workspace is one entry from `herdr workspace list` — a "space" in herdr's UI.
type Workspace struct {
	WorkspaceID string `json:"workspace_id"`
	Label       string `json:"label"`
	Number      int    `json:"number"`
	AgentStatus string `json:"agent_status"`
	Focused     bool   `json:"focused"`
	PaneCount   int    `json:"pane_count"`
	TabCount    int    `json:"tab_count"`
	ActiveTabID string `json:"active_tab_id"`
}

// cliEnvelope is herdr's CLI JSON output shape: {"id":..,"result":{..}}.
type cliEnvelope[T any] struct {
	Result T `json:"result"`
}

// runJSON runs a herdr CLI subcommand and decodes its JSON result payload.
func runJSON[T any](args ...string) (T, error) {
	var result T
	out, err := exec.Command(herdrBin(), args...).Output()
	if err != nil {
		return result, fmt.Errorf("herdr %v: %w", args, err)
	}
	var env cliEnvelope[T]
	if err := json.Unmarshal(out, &env); err != nil {
		return result, fmt.Errorf("decode herdr %v: %w", args, err)
	}
	return env.Result, nil
}

// listAgents returns all detected agents (Agent field non-empty), newest herdr
// ordering preserved.
func listAgents() ([]Agent, error) {
	res, err := runJSON[struct {
		Agents []Agent `json:"agents"`
	}]("agent", "list")
	if err != nil {
		return nil, err
	}
	agents := make([]Agent, 0, len(res.Agents))
	for _, a := range res.Agents {
		if a.Agent != "" {
			agents = append(agents, a)
		}
	}
	return agents, nil
}

// listWorkspaces returns all spaces.
func listWorkspaces() ([]Workspace, error) {
	res, err := runJSON[struct {
		Workspaces []Workspace `json:"workspaces"`
	}]("workspace", "list")
	if err != nil {
		return nil, err
	}
	return res.Workspaces, nil
}

// workspaceLabels maps workspace_id → human label, so the agents picker can show
// which space each agent lives in. Best-effort: a lookup miss just yields "".
func workspaceLabels() map[string]string {
	labels := map[string]string{}
	spaces, err := listWorkspaces()
	if err != nil {
		return labels
	}
	for _, w := range spaces {
		labels[w.WorkspaceID] = w.Label
	}
	return labels
}

// focusAgent brings the given agent (by terminal id) into focus.
func focusAgent(terminalID string) error {
	return exec.Command(herdrBin(), "agent", "focus", terminalID).Run()
}

// focusWorkspace switches to the given space.
func focusWorkspace(workspaceID string) error {
	return exec.Command(herdrBin(), "workspace", "focus", workspaceID).Run()
}

// openOverlay asks herdr to open one of this plugin's panes as an overlay pop-up.
// The action entry points call this; herdr creates the pane, runs its command,
// and restores the previous focus when the picker exits.
func openOverlay(entrypoint string) {
	cmd := exec.Command(herdrBin(),
		"plugin", "pane", "open",
		"--plugin", pluginID,
		"--entrypoint", entrypoint,
		"--placement", "overlay",
	)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		errExit("could not open the "+entrypoint+" overlay:", err)
	}
}
