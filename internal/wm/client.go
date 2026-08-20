package wm

import "github.com/jezek/xgb/xproto"

type Client struct {
	win   xproto.Window
	frame xproto.Window
	ws    *Workspace
	r     rect
	title string
}
