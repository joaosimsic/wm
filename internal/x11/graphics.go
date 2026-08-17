package x11

import "github.com/jezek/xgb/xproto"

func (c *Connection) NewGC(fg, bg uint32) (xproto.Gcontext, error) {
	id, err := xproto.NewGcontextId(c.conn)
	if err != nil {
		return 0, err
	}

	err = xproto.CreateGCChecked(c.conn, id, xproto.Drawable(c.RootWindow()),
		xproto.GcForeground|xproto.GcBackground, []uint32{fg, bg}).Check()
	if err != nil {
		return 0, err
	}

	return id, nil
}

func (c *Connection) FreeGC(gc xproto.Gcontext) error {
	return xproto.FreeGCChecked(c.conn, gc).Check()
}

func (c *Connection) OpenFont(name string) (xproto.Font, error) {
	id, err := xproto.NewFontId(c.conn)
	if err != nil {
		return 0, err
	}

	err = xproto.OpenFontChecked(c.conn, id, uint16(len(name)), name).Check()
	if err != nil {
		return 0, err
	}

	return id, nil
}

func (c *Connection) CloseFont(font xproto.Font) error {
	return xproto.CloseFontChecked(c.conn, font).Check()
}

func (c *Connection) FillRect(win xproto.Window, gc xproto.Gcontext, x, y, w, h int) error {
	return xproto.PolyFillRectangleChecked(c.conn, xproto.Drawable(win), gc, []xproto.Rectangle{{X: int16(x), Y: int16(y), Width: uint16(w), Height: uint16(h)}}).Check()
}

func (c *Connection) DrawText(win xproto.Window, gc xproto.Gcontext, x, y int, s string) error {
	if len(s) > 255 {
		s = s[:255]
	}
	return xproto.ImageText8Checked(c.conn, byte(len(s)), xproto.Drawable(win), gc, int16(x), int16(y), s).Check()
}

func (c *Connection) TextWidth(font xproto.Font, s string) (int, error) {
	chars := make([]xproto.Char2b, len(s))
	for i := range s {
		chars[i] = xproto.Char2b{Byte1: 0, Byte2: s[i]}
	}
	reply, err := xproto.QueryTextExtents(c.conn, xproto.Fontable(font), chars, uint16(len(chars))).Reply()
	if err != nil {
		return 0, err
	}
	return int(reply.OverallWidth), nil
}
