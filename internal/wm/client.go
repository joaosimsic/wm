package wm

import "github.com/jezek/xgb/xproto"

type Client struct {
	win        xproto.Window
	frame      xproto.Window
	ws         *Workspace
	x, y, w, h int
	title      string
}
