package x11

import "github.com/jezek/xgb/xproto"

func (c *Connection) CreateFontCursor(source, mask uint16) (xproto.Cursor, error) {
	font, err := c.OpenFont("cursor")
	if err != nil {
		return 0, err
	}
	defer func() {
		_ = c.CloseFont(font)
	}()

	id, err := xproto.NewCursorId(c.conn)
	if err != nil {
		return 0, err
	}

	err = xproto.CreateGlyphCursorChecked(
		c.conn,
		id,
		font,
		font,
		source,
		mask,
		0, 0, 0,
		0xffff, 0xffff, 0xffff,
	).Check()
	if err != nil {
		return 0, err
	}

	return id, nil
}

func (c *Connection) SetCursor(win xproto.Window, cursor xproto.Cursor) error {
	return xproto.ChangeWindowAttributesChecked(c.conn, win, xproto.CwCursor, []uint32{uint32(cursor)}).Check()
}
