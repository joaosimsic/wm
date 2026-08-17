package x11

import (
	"sync"

	"github.com/jezek/xgb"
	"github.com/jezek/xgb/xproto"
)

type Connection struct {
	conn  *xgb.Conn
	atoms map[string]xproto.Atom
	mu    sync.Mutex
}

type RGB struct {
	R, G, B uint8
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

func (c *Connection) AllocColorRGB(rgb RGB) (uint32, error) {
	return c.AllocColor(scale16(rgb.R), scale16(rgb.G), scale16(rgb.B))
}

func scale16(v uint8) uint16 {
	return uint16(v)<<8 | uint16(v)
}
