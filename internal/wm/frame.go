package wm

import (
	"github.com/jezek/xgb/xproto"
	"github.com/joaosimsic/wm/internal/x11"
)

type FrameStyle struct {
	BorderWidth    int
	Background     uint32
	BorderActive   uint32
	BorderInactive uint32
	Cursor         xproto.Cursor
}

func newFrame(c *x11.Connection, style FrameStyle, r rect) (xproto.Window, error) {
	win, err := c.CreateWindowOR(
		c.RootWindow(),
		r.x,
		r.y,
		r.w,
		r.h,
		style.BorderWidth,
		style.Background,
		style.BorderActive)
	if err != nil {
		return 0, err
	}

	_ = c.SetCursor(win, style.Cursor)

	return win, nil
}
