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
}

func newFrame(c *x11.Connection, style FrameStyle, r rect) (xproto.Window, error) {
	return c.CreateWindow(c.RootWindow(), r.x, r.y, r.w, r.h, style.BorderWidth, style.Background, style.BorderActive)
}
