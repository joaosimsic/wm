# WM — Refinement Notes

Review of the current state of `internal/wm` (post rect-refactor). These are the
remaining issues that block building or correct runtime behavior. No code changes
are included here — this is a checklist of what needs to change.

## 1. Compile blocker

`internal/wm/layout.go` imports `github.com/jezek/xgb/xproto` but never uses it.
This fails `make build` with `imported and not used`.

- Remove the `xproto` import from `layout.go` (only `x11` is used there).

## 2. `m.done` is never initialized

`Manager.quit()` closes `m.done`, and `Run()` selects on it, but `New()` never
allocates the channel. `quit` would panic on a nil channel (`close(nil)`).

- Initialize `done: make(chan struct{})` in `New()` alongside the other fields.

## 3. `manage()` is skeletal

`Manager.manage()` only computes geometry, creates a frame, and reparents. It
never:

- zeroes the child border width
- selects `PropertyChange` events on the child
- sets `WM_PROTOCOLS` (so `WM_DELETE_WINDOW` never works)
- creates and registers the `Client`
- adds the client to the current workspace
- updates the title
- maps the child and the frame
- focuses and arranges

Without this, windows are reparented into an unmapped, untracked frame and never
appear or tile.

## 4. MapRequest recursion

Root has `SubstructureRedirect`, so every `MapWindow` issued by the WM itself
(for the frame or the client) generates another `MapRequest`. Without a guard
and override-redirect frames, `manage` re-enters and loops indefinitely on the
first window.

Two-part fix:

- **Frames must be override-redirect.** Add a `CreateWindowOR` primitive to
  `internal/x11` (sets `CwOverrideRedirect`), and have `newFrame` use it. OR
  windows are mapped directly, so mapping a frame no longer yields a
  `MapRequest`.
- **Guard `onMapRequest`.** If the window is already in `m.clients`, just map it
  (the server intercepted our own `MapWindow`) and return instead of re-managing.
  Keep the existing `OverrideRedirect` fast path.

## 5. Optional

- `setup()` does not call `adoptExisting()`, so windows opened before the WM
  starts are never managed. Call it before `focusFirst()`/`redrawBar()`.
- The bar window is mapped before `SelectRootEvents` sets the redirect, so it is
  not re-captured — but note this ordering is load-bearing. If the bar is ever
  remapped at runtime it should also be override-redirect (or guarded in
  `onMapRequest`).
