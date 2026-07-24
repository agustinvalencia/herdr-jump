# herdr-jump

Two [herdr](https://herdr.dev) overlay pickers that browse **spaces** and
**agents** *separately* — unlike the built-in session navigator, which shows
everything together.

- **Agents picker** — every detected agent as a **card**, grouped by space in the
  same order as herdr's sidebar (workspace order, tab order preserved), with a
  status dot coloured by state (green idle · yellow working · red blocked · teal
  done). Pick one to focus it.
- **Spaces picker** — every workspace as a card with its status and pane/tab
  counts. Pick one to switch to it.

Each entry is a bordered card: a header (hotkey number + status) above the fields
you choose to show (see [Configuration](#configuration)). The highlighted card
gets a thick accent border. Both pickers open as temporary **overlay** pop-ups;
herdr restores your previous focus when you dismiss them.

## How it works

herdr exposes pop-ups as plugin **panes** with `placement = "overlay"`. Each
action (`agents`, `spaces`) asks herdr to open its picker pane via
`herdr plugin pane open --placement overlay`; the picker is a small
[bubbletea](https://github.com/charmbracelet/bubbletea) TUI that reads
`herdr agent list` / `herdr workspace list` (over `$HERDR_BIN_PATH`) and calls
`herdr agent focus` / `herdr workspace focus` on your selection.

## Install

Requires **herdr ≥ 0.7.0**.

```bash
herdr plugin install agustinvalencia/herdr-jump
```

herdr clones the repo, runs `scripts/build.sh`, and registers the actions. The
build step **prefers a local Go toolchain** (an exact build of the source) and
**falls back to downloading the latest prebuilt release binary** — so it works
**with or without Go**.

**Local development:** build and link a checkout in place:

```bash
sh scripts/build.sh
herdr plugin link /path/to/herdr-jump
```

## Keybindings

Bound in `~/.config/herdr/config.toml` (reload with `herdr server reload-config`):

```toml
[[keys.command]]
key = "prefix+a"
type = "plugin_action"
command = "agustinvalencia.herdr-jump.agents"

[[keys.command]]
key = "prefix+A"
type = "plugin_action"
command = "agustinvalencia.herdr-jump.spaces"
```

## Keys inside a picker

Modal (lazygit-style):

- **nav mode** (default): `j`/`k` or `↑`/`↓` (or `Ctrl+n`/`Ctrl+p`) move · `1`–`9`
  jump to and select that item directly · `g`/`G` top/bottom · `Enter` focuses
  the highlighted item · `/` starts filtering · `q`/`Esc` cancel.
- **filter mode** (after `/`): type to fuzzy-filter · `Enter` focuses · `Esc`
  clears the filter and returns to nav mode.

## Configuration

Settings live in `config.toml` in the plugin config dir (survives upgrades):

```bash
"$(herdr plugin config-dir agustinvalencia.herdr-jump)/config.toml"
```

See [`config.example.toml`](config.example.toml). Keys:

```toml
# Where the picker card sits in the overlay pane.
# Horizontal: left | center | right   Vertical: top | middle | bottom
# Combine them, e.g. "center", "top-left", "bottom right", "top".
align = "center"

# Cap the card width in columns so it reads as a centered card on wide panes.
# 0 disables the cap.
max_width = 96
```

The card is centered by default, so on a large multi-pane window it stays under
your eyes rather than pinned to the top-left corner. (`HERDR_JUMP_ALIGN` still
overrides `align` at runtime if set — handy for quick testing.)

### Card fields

What each card shows is configurable, mirroring herdr's own
[sidebar `rows` model](https://herdr.dev/docs/configuration/#ui-and-sidebar):
one inner list per line inside the card, listing the field tokens to render.
Cells that resolve empty are dropped; unknown tokens are ignored. Omit a section
to keep the built-in default layout.

```toml
[agents]
# Default: title-forward — what the agent is doing, then who/where, then cwd.
rows = [
  ["terminal_title_stripped"],
  ["agent", "workspace"],
  ["cwd"],
]

[spaces]
# Default: the space label with its number, then pane/tab counts.
rows = [
  ["workspace", "number"],
  ["panes", "tabs"],
]
```

| Picker | Field tokens |
|---|---|
| agents | `state_icon` · `state_text` · `agent` · `workspace` · `cwd` · `terminal_title` · `terminal_title_stripped` · `pane` · `tab` |
| spaces | `state_icon` · `state_text` · `workspace` · `number` · `panes` · `tabs` |

The header already carries the status, so `state_icon` / `state_text` are usually
only worth adding to a body row if you want them repeated there. herdr's `branch`
and `git_status` tokens aren't available — the herdr CLI doesn't expose them.

## Debugging

`HERDR_BIN_PATH=$(command -v herdr) ./bin/herdr-jump dump` prints the agents and
spaces the pickers would show, as plain text (no TUI) — handy for checking the
herdr CLI integration.
