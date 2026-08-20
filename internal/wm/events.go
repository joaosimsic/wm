package wm

import (
	"github.com/jezek/xgb"
	"github.com/jezek/xgb/xproto"
	"go.uber.org/zap"
)

func (m *Manager) handle(ev xgb.Event) {
	switch e := ev.(type) {
	case xproto.MapRequestEvent:
		m.onMapRequest(e)
	case xproto.ConfigureRequestEvent:
		m.onConfigureRequest(e)
	case xproto.DestroyNotifyEvent:
		m.onDestroyNotify(e)
	case xproto.UnmapNotifyEvent:
		m.onUnmapNotify(e)
	case xproto.KeyPressEvent:
		m.onKeyPress(e)
	case xproto.PropertyNotifyEvent:
		m.onPropertyNotify(e)
	case xproto.ExposeEvent:
		m.redrawBar()
	}
}

func (m *Manager) onMapRequest(e xproto.MapRequestEvent) {
	attrs, err := m.conn.GetAttributes(e.Window)
	if err != nil {
		return
	}
	if attrs.OverrideRedirect {
		_ = m.conn.MapWindow(e.Window)
		return
	}

	if err := m.manage(e.Window); err != nil {
		m.log.Warn("manage", zap.Uint32("win", uint32(e.Window)), zap.Error(err))
	}
}

func (m *Manager) onConfigureRequest(e xproto.ConfigureRequestEvent) {
	if c, ok := m.clients[e.Window]; ok {
		_ = m.conn.MoveResize(c.frame, c.x, c.y, c.w, c.h)
		return
	}

	_ = m.conn.MoveResize(e.Window, int(e.X), int(e.Y), int(e.Width), int(e.Height))
}

func (m *Manager) onDestroyNotify(e xproto.DestroyNotifyEvent) {
	c, ok := m.clients[e.Window]
	if !ok {
		return
	}

	delete(m.clients, e.Window)
	c.ws.remove(c)

	if m.focused == c {
		m.focused = nil
	}

	m.focusFirst()
	m.arrange()
}

func (m *Manager) onUnmapNotify(e xproto.UnmapNotifyEvent) {
	c, ok := m.clients[e.Window]
	if !ok {
		return
	}

	_ = m.conn.UnmapWindow(c.frame)
}

func (m *Manager) onKeyPress(e xproto.KeyPressEvent) {
	byMask, ok := m.bindings[e.Detail]
	if !ok {
		return
	}

	state := e.State &^ (m.kbd.NumLock | xproto.KeyButMaskLock)
	fn, ok := byMask[state]
	if !ok {
		return
	}

	if err := fn(m); err != nil {
		m.log.Warn("action", zap.Error(err))
	}
}

func (m *Manager) onPropertyNotify(e xproto.PropertyNotifyEvent) {
	c, ok := m.clients[e.Window]
	if !ok {
		return
	}

	if e.Atom == xproto.AtomWmName || e.Atom == m.atoms.netWMName {
		m.updateTitle(c)
	}
}
