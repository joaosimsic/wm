package x11

import (
	"github.com/jezek/xgb"
	"github.com/jezek/xgb/xproto"
)

type Connection struct {
	conn *xgb.Conn
}

func Connect() (*Connection, error) {
	conn, err := xgb.NewConn()
	if err != nil {
		return nil, err
	}

	return &Connection{
		conn: conn,
	}, nil
}

func (c *Connection) Close() {
	c.conn.Close()
}

func (c *Connection) Raw() *xgb.Conn {
	return c.conn
}

func (c *Connection) Screen() *xproto.ScreenInfo {
	return xproto.Setup(c.conn).DefaultScreen(c.conn)
}

func (c *Connection) RootWindow() xproto.Window {
	return c.Screen().Root
}

func (c *Connection) NewWindowID() (xproto.Window, error) {
	return xproto.NewWindowId(c.conn)
}

func (c *Connection) WaitForEvent() (xgb.Event, xgb.Error) {
	return c.conn.WaitForEvent()
}

func (c *Connection) Colormap() xproto.Colormap {
	return c.Screen().DefaultColormap
}

func (c *Connection) BlackPixel() uint32 { return uint32(c.Screen().BlackPixel) }

func (c *Connection) WhitePixel() uint32 { return uint32(c.Screen().WhitePixel) }

func (c *Connection) AllocColor(r, g, b uint16) (uint32, error) {
	reply, err := xproto.AllocColor(c.conn, c.Colormap(), r, g, b).Reply()
	if err != nil {
		return 0, err
	}

	return uint32(reply.Pixel), nil
}
