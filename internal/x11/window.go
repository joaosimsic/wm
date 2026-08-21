package x11

import "github.com/jezek/xgb/xproto"

func (c *Connection) CreateWindow(parent xproto.Window, x, y, w, h, bw int, bg, border uint32) (xproto.Window, error) {
	win, err := c.NewWindowID()
	if err != nil {
		return 0, err
	}

	err = xproto.CreateWindowChecked(
		c.conn,
		c.Screen().RootDepth,
		win,
		parent,
		int16(x),
		int16(y),
		uint16(w),
		uint16(h),
		uint16(bw),
		xproto.WindowClassInputOutput,
		c.Screen().RootVisual,
		xproto.CwBackPixel|xproto.CwBorderPixel,
		[]uint32{bg, border}).Check()
	if err != nil {
		return 0, err
	}

	return win, nil
}

func (c *Connection) CreateWindowOR(parent xproto.Window, x, y, w, h, bw int, bg, border uint32) (xproto.Window, error) {
	win, err := c.NewWindowID()
	if err != nil {
		return 0, err
	}

	CwOverrideRedirectValue := uint32(1)
	values := []uint32{bg, border, CwOverrideRedirectValue}

	err = xproto.CreateWindowChecked(
		c.conn,
		c.Screen().RootDepth,
		win,
		parent,
		int16(x),
		int16(y),
		uint16(w),
		uint16(h),
		uint16(bw),
		xproto.WindowClassInputOutput,
		c.Screen().RootVisual,
		xproto.CwBackPixel|xproto.CwBorderPixel|xproto.CwOverrideRedirect,
		values).Check()
	if err != nil {
		return 0, err
	}

	return win, nil
}

func (c *Connection) ReparentWindow(win, parent xproto.Window, x, y int) error {
	return xproto.ReparentWindowChecked(c.conn, win, parent, int16(x), int16(y)).Check()
}

func (c *Connection) MapWindow(win xproto.Window) error {
	return xproto.MapWindowChecked(c.conn, win).Check()
}

func (c *Connection) UnmapWindow(win xproto.Window) error {
	return xproto.UnmapWindowChecked(c.conn, win).Check()
}

func (c *Connection) DestroyWindow(win xproto.Window) error {
	return xproto.DestroyWindowChecked(c.conn, win).Check()
}

func (c *Connection) QueryTree(win xproto.Window) (xproto.Window, []xproto.Window, error) {
	reply, err := xproto.QueryTree(c.conn, win).Reply()
	if err != nil {
		return 0, nil, err
	}

	return reply.Parent, reply.Children, nil
}

func (c *Connection) MoveResize(win xproto.Window, x, y, w, h int) error {
	return xproto.ConfigureWindowChecked(
		c.conn,
		win,
		xproto.ConfigWindowX|xproto.ConfigWindowY|xproto.ConfigWindowWidth|xproto.ConfigWindowHeight,
		[]uint32{uint32(x), uint32(y), uint32(w), uint32(h)}).Check()
}

func (c *Connection) SetBorderWidth(win xproto.Window, width int) error {
	return xproto.ConfigureWindowChecked(c.conn, win, xproto.ConfigWindowBorderWidth, []uint32{uint32(width)}).Check()
}

func (c *Connection) SetBorderColor(win xproto.Window, pixel uint32) error {
	return xproto.ChangeWindowAttributesChecked(c.conn, win, xproto.CwBorderPixel, []uint32{pixel}).Check()
}

func (c *Connection) GetGeometry(win xproto.Window) (x, y, w, h int, err error) {
	reply, err := xproto.GetGeometry(c.conn, xproto.Drawable(win)).Reply()
	if err != nil {
		return 0, 0, 0, 0, err
	}

	return int(reply.X), int(reply.Y), int(reply.Width), int(reply.Height), nil
}

func (c *Connection) GetAttributes(win xproto.Window) (*xproto.GetWindowAttributesReply, error) {
	return xproto.GetWindowAttributes(c.conn, win).Reply()
}
