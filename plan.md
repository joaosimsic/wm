# WM — Project Plan

A keyboard-driven dynamic tiling window manager for X11, inspired by i3 / Hyprland.
Written in Go, using `github.com/jezek/xgb`.

## Vision

- Tiling WM with a **dynamic master/stack** layout (Hyprland/dwm style).
- **Keyboard-only** interaction — no mouse required.
- Minimal built-in status **bar**.
- **Config-driven** keybindings and appearance.

## Current State

Done:

- TOML config loader (`internal/config`) — path resolution, parsing, validation, defaults.
- Hex → RGB color parsing (`internal/config/color.go`).
- X11 connection wrapper (`internal/x11/client.go`) — connect, screen, root, colormap, color allocation.
- Palette allocation (`internal/theme/palette.go`).
- Xephyr dev harness (`cmd/wm-dev/main.go`).

Missing (everything that makes it a WM):

- Substructure redirect + event loop.
- Window reparenting / framing / borders.
- Tiling layout.
- Keyboard focus.
- Keybindings (grab + dispatch).
- Workspaces.
- Status bar.
- EWMH/ICCCM basics.

`cmd/wm/main.go` currently connects, loads config, allocates a palette, logs, and exits.

## Decisions

| Topic | Decision |
|---|---|
| Tiling model | Dynamic master/stack (one master window, rest stacked) |
| Keybindings | Config-driven from the start |
| Bar | Minimal bar included in first milestone |
| Focus | Keyboard-only (`SetInputFocus`), no mouse focus |

## Roadmap

### 1. `internal/x11/wm.go` — low-level X primitives

Add WM-oriented methods on the existing `*Connection`.

Constants (missing from `xproto`):

```go
const (
    RevertToNone        = 0
    RevertToPointerRoot = 1
    RevertToParent      = 2
)
```

**Event selection**
- `SelectRootEvents(mask uint32) error` — `ChangeWindowAttributes` on root.
- `SelectWindowEvents(win, mask uint32) error`.

**Window lifecycle / geometry**
- `CreateWindow(parent, x, y, w, h, bw, bg, border) (Window, error)`.
- `ReparentWindow`, `MapWindow`, `UnmapWindow`, `DestroyWindow`.
- `MoveResize(win, x, y, w, h)`.
- `SetBorderWidth(win, width)`, `SetBorderColor(win, pixel)`.
- `QueryTree(win) (parent, children, error)` — startup adoption.

**Properties / atoms**
- `InternAtom`, `ChangeProperty`, `GetProperty`, `DeleteProperty`.

**Focus**
- `SetInputFocus(win, revertTo byte)` — `CurrentTime`.

**Keyboard / keys**
- `GrabKey`, `UngrabKey`.
- `KeycodeRange()` from `Setup`.
- `KeyboardMapping()` from `GetKeyboardMapping`.
- `ModifierMask(name string) (uint16, bool)` — maps mod names to `KeyButMask*`.

**Graphics (bar)**
- `OpenFont`, `CloseFont`, `NewGC`, `FreeGC`.
- `FillRect`, `DrawText` (`ImageText8`), `TextWidth` (`QueryTextExtents`).

### 2. Config — keybindings & layout knobs

Extend `Config` + `Default()`:

- `Mod string` (default `"Mod4"`).
- `Keys` — map of action name → key symbol (e.g. `"spawn" = "Return"`).
- `SplitRatio float64` (default `0.5`) — master width fraction.
- `Workspaces int` (default `9`).

### 3. `internal/wm` — WM core

- `manager.go` — `Manager`, `Run()` event loop, event dispatch.
- `client.go` — `Client{win, frame, x, y, w, h, ws}`.
- `layout.go` — master/stack geometry (honors `Gap`, `BarHeight`, `SplitRatio`).
- `workspace.go` — N virtual desktops.
- `keys.go` — keysym → keycode resolution, `GrabKey`, `KeyPress` dispatch.
- `actions.go` — action implementations.
- `ewmh.go` — `WM_DELETE_WINDOW`, `WM_NAME`, `_NET_WM_NAME`, `PropertyNotify`.
- `bar.go` — top bar: workspace numbers + active title.

### 4. Wire up `cmd/wm/main.go`

Replace exit-log with `wm.New(conn, cfg, palette, logger).Run()`.
`Mod+Shift+q` returns cleanly so `wm-dev` can tear down Xephyr.

### 5. Verification

- `make build && make vet`.
- `make dev` (Xephyr `:1`), then `DISPLAY=:1 xterm &`.
- Check: borders, master/stack split, gaps, keyboard focus, workspace switch, bar.

## Default Keybindings

All under `Mod` (default `Mod4`), overridable via TOML.

| Action | Default key |
|---|---|
| `spawn` | `Return` |
| `close` | `q` |
| `quit` | `Shift+q` |
| `focus_next` / `focus_prev` | `j` / `k` |
| `swap_next` / `swap_prev` | `Shift+j` / `Shift+k` |
| `focus_master` | `m` |
| `grow_master` / `shrink_master` | `l` / `h` |
| `workspace_1..9` | `1..9` |
| `move_to_workspace_1..9` | `Shift+1..9` |

## Proposed Config Schema

```toml
terminal = "xterm"
font = "fixed"
border_width = 2
gap = 0
bar_height = 22
mod = "Mod4"
split_ratio = 0.5
workspaces = 9

[colors]
bar_bg = "#1a1a1a"
bar_fg = "#cccccc"
bar_active_bg = "#4a4a4a"
border_active = "#5f87af"
border_inactive = "#333333"

[keys]
spawn = "Return"
close = "q"
quit = "Shift+q"
focus_next = "j"
focus_prev = "k"
# ... etc
```

## Out of Scope (later)

- Config reload.
- Multi-monitor.
- i3-style split tree / resize modes.
- Full EWMH (`_NET_ACTIVE_WINDOW` for external tools).
- Mouse support.
- Status modules / scripting.

## Open Questions

1. **Keybinding combo format** — single global `Mod` + per-action keysym (current plan) vs. full combo strings per action (e.g. `"Mod4+Shift+q"`)? Affects the `[keys]` TOML schema and `keys.go` parsing.
2. **Text is 8-bit only** — `ImageText8`/`Char2b` limit the bar to Latin-1/ASCII. Acceptable, or defer text and draw workspace numbers as colored blocks only?
3. **Master/stack orientation** — master on the left (default) vs. right/top? Configurable via `layout` key or hardcode left-first?
4. **Split-ratio resizing granularity** — fixed step (e.g. `0.05`) for `grow_master`/`shrink_master`, or a configurable `resize_step`?
5. **`ModifierMask` scope** — keysym-name table + combo parsing (`"Shift+q"`) live in `wm/keys.go`, not `x11/wm.go` — confirm this separation is desired.
