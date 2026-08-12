package x11

import (
	"github.com/jezek/xgb"
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
