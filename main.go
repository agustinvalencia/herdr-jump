package main

import (
	"fmt"
	"os"
)

// pluginID must match the id in herdr-plugin.toml. It is used when asking herdr
// to open this plugin's overlay panes.
const pluginID = "agustinvalencia.herdr-jump"

const version = "0.1.0"

// main dispatches on the subcommand herdr passes per manifest entry point:
//
//   - "agents" / "spaces" are the actions a keybinding runs; each asks herdr to
//     open its picker as an overlay pane.
//   - "agents-ui" / "spaces-ui" are those pickers; herdr runs them inside the
//     overlay pane it opens, so you never run them directly.
func main() {
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "agents":
			openOverlay("agents-picker")
			return
		case "agents-ui":
			runAgentsUI()
			return
		case "spaces":
			openOverlay("spaces-picker")
			return
		case "spaces-ui":
			runSpacesUI()
			return
		case "dump":
			// Hidden: print the agent/space data as plain text (no TUI), for
			// debugging the herdr CLI integration outside an overlay pane.
			runDump()
			return
		case "preview":
			// Hidden: render the picker View at a forced WIDTHxHEIGHT (default
			// 120x40) to stdout, for eyeballing layout/alignment without a TTY.
			runPreview(os.Args[2:])
			return
		case "version", "--version", "-v", "-V":
			fmt.Println("herdr-jump", version)
			return
		}
	}
	errExit("a herdr plugin; run its actions through herdr (e.g. `herdr plugin action invoke " + pluginID + ".agents`) or `herdr-jump version`.")
}

// errExit prints a message to stderr and exits non-zero.
func errExit(args ...any) {
	fmt.Fprintln(os.Stderr, append([]any{"herdr-jump:"}, args...)...)
	os.Exit(1)
}
