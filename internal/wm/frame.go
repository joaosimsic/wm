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

func newFrame(c *x11.Connection, style FrameStyle, x, y, w, h int) (xproto.Window, error) {
	return c.CreateWindow(c.RootWindow(), x, y, w, h, style.BorderWidth, style.Background, style.BorderActive)
}
