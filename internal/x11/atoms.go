package x11

import "github.com/jezek/xgb/xproto"

func (c *Connection) InternAtom(name string) (xproto.Atom, error) {
	if c.atoms == nil {
		c.atoms = make(map[string]xproto.Atom)
	}

	if a, ok := c.atoms[name]; ok {
		return a, nil
	}

	reply, err := xproto.InternAtom(c.conn, false, uint16(len(name)), name).Reply()
	if err != nil {
		return 0, err
	}

	c.atoms[name] = reply.Atom

	return reply.Atom, nil
}

func (c *Connection) ChangeProperty(win xproto.Window, prop, typ xproto.Atom, data []byte) error {
	return xproto.ChangePropertyChecked(c.conn, xproto.PropModeReplace, win, prop, typ, 8, uint32(len(data)), data).Check()
}

func (c *Connection) GetProperty(win xproto.Window, prop, typ xproto.Atom) ([]byte, error) {
    reply, err := xproto.GetProperty(c.conn, false, win, prop, typ, 0, 1024).Reply()
    if err != nil {
        return nil, err
    }

    return reply.Value, nil
}

func (c *Connection) DeleteProperty(win xproto.Window, prop xproto.Atom) error {
    return xproto.DeletePropertyChecked(c.conn, win, prop).Check()
}
