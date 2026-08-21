# Add a visible mouse cursor

## Background / root cause

The WM (a keyboard-driven X11 tiling WM in Go using `github.com/jezek/xgb`
core protocol) never set a cursor on any window. The first attempt used a
Core-X **font cursor** via `CreateGlyphCursor` on the `"cursor"` font with
glyph `68` (`XC_left_ptr`).

Two bugs were found:

1. `CreateGlyphCursor(68, 68)` draws a **solid black arrow** (mask == source),
   which is invisible on the black `-br` Xephyr root (black-on-black).
2. Fixing it to `CreateGlyphCursor(68, 69)` (correct `source, source+1` pair)
   **crashes the WM**: Xephyr's `"cursor"` font lacks glyph 69, so the server
   returns a `BadValue` X error. `setup()` returns that error, `wm.New()` fails,
   `main` calls `logger.Fatal`, and the WM **never starts** — hence a truly
   black screen with no bar and no cursor.

## Approach

Replace the font-dependent glyph cursor with a **pixmap cursor** built from
depth-1 pixmaps. This removes all dependency on the X `"cursor"` font and works
in Xephyr. Cursor creation is also made **non-fatal** so a cursor failure can
never blank the whole WM again.

## Steps

### 1. Replace `internal/x11/cursor.go`

```go
package x11

import "github.com/jezek/xgb/xproto"

// CreateArrowCursor builds a 16x16 left-pointing arrow from depth-1 pixmaps,
// so it does not depend on the X "cursor" font (which Xephyr may not have).
func (c *Connection) CreateArrowCursor() (xproto.Cursor, error) {
	const w, h uint16 = 16, 16

	src, err := xproto.NewPixmapId(c.conn)
	if err != nil {
		return 0, err
	}
	if err := xproto.CreatePixmapChecked(c.conn, 1, src, xproto.Drawable(c.RootWindow()), w, h).Check(); err != nil {
		return 0, err
	}

	mask, err := xproto.NewPixmapId(c.conn)
	if err != nil {
		return 0, err
	}
	if err := xproto.CreatePixmapChecked(c.conn, 1, mask, xproto.Drawable(c.RootWindow()), w, h).Check(); err != nil {
		return 0, err
	}

	// depth-1 GC: foreground bit 1 = "set"
	srcGC, err := xproto.NewGcontextId(c.conn)
	if err != nil {
		return 0, err
	}
	if err := xproto.CreateGCChecked(c.conn, srcGC, xproto.Drawable(src), xproto.GcForeground, []uint32{1}).Check(); err != nil {
		return 0, err
	}

	maskGC, err := xproto.NewGcontextId(c.conn)
	if err != nil {
		return 0, err
	}
	if err := xproto.CreateGCChecked(c.conn, maskGC, xproto.Drawable(mask), xproto.GcForeground, []uint32{1}).Check(); err != nil {
		return 0, err
	}

	// Draw the same arrow shape into source (color) and mask (visibility).
	for _, r := range arrowRects() {
		if err := xproto.PolyFillRectangleChecked(c.conn, xproto.Drawable(src), srcGC, []xproto.Rectangle{r}).Check(); err != nil {
			return 0, err
		}
		if err := xproto.PolyFillRectangleChecked(c.conn, xproto.Drawable(mask), maskGC, []xproto.Rectangle{r}).Check(); err != nil {
			return 0, err
		}
	}

	id, err := xproto.NewCursorId(c.conn)
	if err != nil {
		return 0, err
	}
	if err := xproto.CreatePixmapCursorChecked(
		c.conn, id, src, mask,
		0xffff, 0xffff, 0xffff, // fg = white
		0, 0, 0, // bg = black
		1, 1, // hot spot
	).Check(); err != nil {
		return 0, err
	}
	return id, nil
}

func arrowRects() []xproto.Rectangle {
	var rs []xproto.Rectangle
	for y := 0; y < 9; y++ {
		rs = append(rs, xproto.Rectangle{X: 1, Y: int16(y), Width: uint16(9 - y), Height: 1})
	}
	rs = append(rs, xproto.Rectangle{X: 1, Y: 9, Width: 3, Height: 6})
	return rs
}
```

> The pixmaps/GCs are intentionally left allocated: the cursor references them
> for its lifetime, and freeing risks an invisible cursor on some servers. The
> leak is two tiny pixmaps + two GCs per WM session.

The existing `SetCursor` helper in `cursor.go` is reused as-is:

```go
func (c *Connection) SetCursor(win xproto.Window, cursor xproto.Cursor) error {
	return xproto.ChangeWindowAttributesChecked(c.conn, win, xproto.CwCursor, []uint32{uint32(cursor)}).Check()
}
```

### 2. `internal/wm/manager.go` — `setup()` cursor block

Replace the `XCLeftPtr` / `CreateFontCursor` block with a non-fatal one:

```go
cursor, err := m.conn.CreateArrowCursor()
if err != nil {
	m.log.Warn("create cursor", zap.Error(err))
} else {
	m.cursor = cursor
	if err := m.conn.SetCursor(m.conn.RootWindow(), cursor); err != nil {
		m.log.Warn("set root cursor", zap.Error(err))
	}
}
```

(Delete the old `XCLeftPtr := uint16(68)` line and the `CreateFontCursor` call.
The `m.cursor xproto.Cursor` field already exists on `Manager`.)

### 3. Make `SetCursor` best-effort in `newBar` and `newFrame`

`internal/wm/bar.go` (~line 52):

```go
// before:
if err := conn.SetCursor(b.win, cursor); err != nil {
	return err
}
// after:
_ = conn.SetCursor(b.win, cursor)
```

`internal/wm/frame.go` (~line 30):

```go
// before:
if err := c.SetCursor(win, style.Cursor); err != nil {
	return 0, err
}
// after:
_ = c.SetCursor(win, style.Cursor)
```

`manage()` already sets `Cursor: m.cursor` in the `FrameStyle`, so frames pick
up the cursor automatically.

## Verification

1. `go build ./...` — must pass.
2. `make dev` — the **bar strip must now appear** (proves the WM started; the
   old `shape+1` glyph was fatally crashing it). A white-outlined arrow should
   follow the mouse over the root, frames, and bar.
3. Independent Xephyr check: `DISPLAY=:1 xterm` (or `xeyes`). If that also shows
   nothing, the problem is Xephyr/display, not the WM.
