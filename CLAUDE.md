# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Overview

`herdr-jump` is a plugin for [herdr](https://herdr.dev), a terminal session
manager. It adds two overlay pickers — one for **spaces** (workspaces), one for
**agents** — that herdr's built-in combined navigator does not offer separately.
Single Go module, one flat `package main` in the repo root (no subpackages).

## Commands

```bash
go build -o bin/herdr-jump .   # build the binary the plugin runs
sh scripts/build.sh            # what herdr runs at install: go build, or download prebuilt if no Go
go test ./...                  # run all tests
go test -run TestNavMotion .   # run a single test
go vet ./...                   # what CI vets
herdr plugin link .            # link this checkout into a running herdr for local dev

# Inspect the herdr CLI integration without a TUI:
HERDR_BIN_PATH=$(command -v herdr) ./bin/herdr-jump dump      # agents + spaces as plain text
./bin/herdr-jump preview 120x40                              # render a picker View at a forced size
```

CI (`.github/workflows/ci.yml`) runs `go vet`, `go test`, then `sh scripts/build.sh`.

## Architecture

**Two-tier command dispatch.** `main.go` switches on `os.Args[1]`; every
subcommand corresponds to an entry in `herdr-plugin.toml`. herdr invokes the
binary twice per interaction:

1. A keybinding fires an **action** (`agents` / `spaces`). That subcommand owns
   no terminal, so it just calls `openOverlay` (`herdr.go`) → `herdr plugin pane
   open --placement overlay`, asking herdr to open the picker as a temporary
   pop-up pane.
2. herdr opens that **pane** and runs the `-ui` subcommand (`agents-ui` /
   `spaces-ui`) inside it. These render the bubbletea TUI. **Never run the `-ui`
   subcommands directly** — they only make sense inside the pane herdr opens, and
   herdr restores your previous focus when the picker exits.

`hidden` subcommands `dump` and `preview` exist only for debugging (see Commands).

**herdr CLI integration (`herdr.go`).** All data and all actions go through
shelling out to the `herdr` CLI, never a socket. herdr injects `HERDR_BIN_PATH`
into every plugin command; `herdrBin()` reads it (falling back to `herdr` on
`$PATH`). `runJSON` decodes herdr's `{"result": …}` CLI envelope. Reads:
`herdr agent list`, `herdr workspace list`. Writes: `herdr agent focus <id>`,
`herdr workspace focus <id>`. Agents with an empty `Agent` field are plain
shells and are filtered out in `listAgents`.

**Ordering must mirror herdr's sidebar.** `agents.go` stable-sorts agents by
their workspace's `Number` (grouped by space, native tab order preserved within a
space); `spaces.go` sorts spaces by `Number`. Do not reorder by status — the
lists are meant to match herdr's own panel and the `prefix+1..9` binds.

**Picker (`picker.go`).** A single modal bubbletea model shared by both pickers,
built from `[]item`. Modal like lazygit: **nav mode** (`j`/`k`, `1`–`9` jump &
select, `g`/`G`, `/` to filter, `enter`, `q`/`esc`) vs **filter mode** (typing
fuzzy-filters via `sahilm/fuzzy`; `esc` clears and returns to nav). The mode
split is what lets letters be both commands and query text. `runPicker` returns
the chosen `id` (empty = cancelled), which the caller hands to the focus command.

Each entry renders as a **card** (`renderCard`): a lipgloss-bordered box with a
header (hotkey number + status) above prerendered body lines. Selected → thick
accent border, else dim rounded. An `item` carries only `id`, `status`, prebuilt
`lines`, `focused`, and `search` — the fields are resolved upstream (see below),
not in the view. Two width gotchas: (1) lines are fit to an exact column count
with `ansi.Truncate` + manual pad (`fitWidth`) and the border is sized to that
content — **do not** use `lipgloss.Style.Width` here, it subtracts padding and
re-wraps; (2) all cards in a picker share the same body-row count, so windowing
uses one uniform card height. Eyeball layout changes with `herdr-jump preview
WxH` (renders the real `agentItems` path, no TTY needed).

**Fields (`fields.go`).** The card body is data-driven, mirroring herdr's
`[ui.sidebar.*]` `rows` model: a `rows [][]string` where each inner list is one
card line of field tokens. `agentCell` / `spaceCell` resolve a token to a styled
string (`""` when empty/unknown); `buildLines` joins a row's non-empty cells and
keeps rows positionally so every card stays the same height. Token vocabulary
follows herdr's names (`state_icon`, `state_text`, `agent`, `workspace`, `cwd`,
`terminal_title*`, `pane`, `tab` for agents; `number`, `panes`, `tabs` for
spaces); herdr's `branch`/`git_status` are omitted (not in the CLI). `agentItems`
/ `spaceItems` are kept pure (take the rows in) so layout + the pane_id focus
target are unit-testable.

**Config (`config.go`).** `config.toml` is read from `HERDR_PLUGIN_CONFIG_DIR`
(herdr's per-plugin config dir, survives upgrades). Keys: `align` (card position,
e.g. `"top-left"`, parsed loosely by `parseAlign`), `max_width` (card width cap;
`0` disables, unset → `defaultMaxWidth`), and the `[agents]` / `[spaces]` `rows`
sections driving the card layout (empty → `agentRows()`/`spaceRows()` defaults).
`HERDR_JUMP_ALIGN` overrides `align` at runtime for quick testing.

**Styling (`styles.go`).** Catppuccin Mocha palette with a blue accent, matched
to the herdr config. `statusColor` maps agent/workspace status → glyph colour
(idle=green, working=yellow, blocked=red, done=teal, else grey) to mirror herdr's
own agent-panel palette.

## Release

Pushing a `v*` tag triggers `.github/workflows/release.yml` → GoReleaser
(`.goreleaser.yml`): cross-compiles linux/darwin × amd64/arm64 and attaches
archives to a GitHub Release. **Archives are deliberately named without a
version** (`herdr-jump_{os}_{arch}.tar.gz`) so the `releases/latest/download/…`
redirect resolves them — this is how `install.sh` fetches a prebuilt binary when
a machine has no Go toolchain.

## Gotchas

- **The version string lives in three places that must stay in sync:** `version`
  in `main.go`, and `version` in `herdr-plugin.toml`, plus the pushed git tag.
  `min_herdr_version` in the manifest gates install compatibility (currently
  `0.7.0`).
- The manifest `id` (`agustinvalencia.herdr-jump`) must match the `pluginID`
  const in `main.go` — it is used to open this plugin's own overlay panes.
