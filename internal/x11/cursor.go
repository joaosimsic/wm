package x11

import (
	"github.com/jezek/xgb"
	"github.com/jezek/xgb/xproto"
)

const (
	cursorSize  uint16 = 16
	cursorDepth uint8  = 1

	white = 0xffff
	black = 0x0000

	hotspot uint16 = 1

	arrowX   int16  = 1
	arrowLen int    = 9
	shaftW   uint16 = 3
	shaftH   uint16 = 6
)

func (c *Connection) CreateCursor() (xproto.Cursor, error) {
	src, err := createPixmap(c.conn, c.RootWindow())
	if err != nil {
		return 0, err
	}
	mask, err := createPixmap(c.conn, c.RootWindow())
	if err != nil {
		return 0, err
	}

	srcGC, err := createGC(c.conn, src)
	if err != nil {
		return 0, err
	}
	maskGC, err := createGC(c.conn, mask)
	if err != nil {
		return 0, err
	}

	for _, r := range arrowRects() {
		if err := xproto.PolyFillRectangleChecked(
			c.conn,
			xproto.Drawable(src),
			srcGC,
			[]xproto.Rectangle{r},
		).Check(); err != nil {
			return 0, err
		}

		if err := xproto.PolyFillRectangleChecked(
			c.conn,
			xproto.Drawable(mask),
			maskGC,
			[]xproto.Rectangle{r},
		).Check(); err != nil {
			return 0, err
		}
	}

	id, err := xproto.NewCursorId(c.conn)
	if err != nil {
		return 0, err
	}

	if err := xproto.CreateCursorChecked(
		c.conn, id, src, mask,
		white, white, white,
		black, black, black,
		hotspot, hotspot,
	).Check(); err != nil {
		return 0, err
	}

	return id, nil
}

func (c *Connection) SetCursor(win xproto.Window, cursor xproto.Cursor) error {
	return xproto.ChangeWindowAttributesChecked(c.conn, win, xproto.CwCursor, []uint32{uint32(cursor)}).Check()
}

func createPixmap(conn *xgb.Conn, win xproto.Window) (xproto.Pixmap, error) {
	pixmap, err := xproto.NewPixmapId(conn)
	if err != nil {
		return 0, err
	}

	if err := xproto.CreatePixmapChecked(
		conn,
		cursorDepth,
		pixmap,
		xproto.Drawable(win),
		cursorSize,
		cursorSize,
	).Check(); err != nil {
		return 0, err
	}

	return pixmap, nil
}

func createGC(conn *xgb.Conn, pixmap xproto.Pixmap) (xproto.Gcontext, error) {
	gcID, err := xproto.NewGcontextId(conn)
	if err != nil {
		return 0, err
	}

	if err := xproto.CreateGCChecked(
		conn,
		gcID,
		xproto.Drawable(pixmap),
		xproto.GcBackground,
		[]uint32{1},
	).Check(); err != nil {
		return 0, err
	}

	return gcID, nil
}

func arrowRects() []xproto.Rectangle {
	var rs []xproto.Rectangle

	for y := range arrowLen {
		rs = append(rs, xproto.Rectangle{X: arrowX, Y: int16(y), Width: uint16(arrowLen - y), Height: 1})
	}
	rs = append(rs, xproto.Rectangle{X: arrowX, Y: int16(arrowLen), Width: shaftW, Height: shaftH})

	return rs
}
