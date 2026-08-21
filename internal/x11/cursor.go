package x11

import (
	"github.com/jezek/xgb"
	"github.com/jezek/xgb/xproto"
)

const w, h uint16 = 16, 16

func (c *Connection) CreateCursor() (xproto.Cursor, error) {
    src, err := createPixmap(c.conn)
    if err != nil {
        return 0, nil
    }

        mask, err := createPixmap(c.conn)
    if err != nil {
        return 0, nil
    }


    srcGC, err := xproto.NewGcontextId(c.conn)
    if err != nil {
        return 0, err
    }

    if err := xproto.CreateGCChecked(
        c.conn,
        srcGC,
        xproto.Drawable(src),
        xproto.GcBackground,
        []uint32{1},
    ).Check(); err != nil {
        return 0, err
    }

}

func (c *Connection) SetCursor(win xproto.Window, cursor xproto.Cursor) error {
	return xproto.ChangeWindowAttributesChecked(c.conn, win, xproto.CwCursor, []uint32{uint32(cursor)}).Check()
}

func createPixmap(conn *xgb.Conn) (xproto.Pixmap, error) {
    pixmap, err := xproto.NewPixmapId(conn)
    if err != nil {
        return 0, err
    }

    if err := xproto.CreatePixmapChecked(
        conn,
        1,
        pixmap,
        xproto.Drawable(c.RootWindow()),
        w,
        h,
    ).Check(); err != nil {
        return 0, err
    }

    return pixmap, nil
}

func createGC(conn *xgb.Conn, pixmap xproto.Pixmap) (xproto.Gcontext, error) {

}

// adding cursor, it was not visible, everything was black
