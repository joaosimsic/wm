package x11

import "github.com/jezek/xgb/xproto"

const (
	RevertToNone        = 0
	RevertToPointerRoot = 1
	RevertToParent      = 2
)

func (c *Connection) ChangeWindowAttributes(win xproto.Window, mask uint32) error {
	return xproto.ChangeWindowAttributesChecked(c.conn, win, xproto.CwEventMask, []uint32{mask}).Check()
}

func (c *Connection) SelectRootEvents(mask uint32) error {
    return c.ChangeWindowAttributes(c.RootWindow(), mask)
}

func (c *Connection) SelectWindowEvents(win xproto.Window, mask uint32) error {
    return c.ChangeWindowAttributes(win, mask)
}

func (c *Connection) SetInputFocus(win xproto.Window, revertTo byte) error {
    return xproto.SetInputFocusChecked(c.conn, revertTo, win, xproto.TimeCurrentTime).Check()
}
