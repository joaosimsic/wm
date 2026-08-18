package wm

import "github.com/jezek/xgb/xproto"

type atoms struct {
	wmProtocols xproto.Atom
	wmDelete    xproto.Atom
	netWMName   xproto.Atom
}

func (m *Manager) initAtoms() error {
	var err error

	if m.atoms.wmProtocols, err = m.conn.InternAtom("WM_PROTOCOLS"); err != nil {
		return err
	}
	if m.atoms.wmDelete, err = m.conn.InternAtom("WM_DELETE_WINDOW"); err != nil {
		return err
	}
	if m.atoms.netWMName, err = m.conn.InternAtom("_NET_WM_NAME"); err != nil {
		return err
	}

	return nil
}

func (m *Manager) setWMProtocols(win xproto.Window) error {
	return m.conn.ChangeProperty32(win, m.atoms.wmProtocols, xproto.AtomAtom, []uint32{uint32(m.atoms.wmDelete)})
}
