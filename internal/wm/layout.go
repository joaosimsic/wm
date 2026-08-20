package wm

import (
	"github.com/jezek/xgb/xproto"
	"github.com/joaosimsic/wm/internal/x11"
)

type rect struct {
	x, y, w, h int
}

func (m *Manager) arrange() {
	sw := int(m.conn.Screen().WidthInPixels)
	sh := int(m.conn.Screen().HeightInPixels)
	g := int(m.cfg.Gap)
	area := rect{
		x: g,
		y: m.cfg.BarHeight + g,
		w: sw - 2*g,
		h: sh - m.cfg.BarHeight - 2*g,
	}

	list := m.current.list
	switch len(list) {
	case 0:
		return
	case 1:
		m.place(list[0], area)
	default:
		mw := int(float64(area.w-g) * m.ratio)

		m.place(list[0], rect{area.x, area.y, mw, area.h})

		stack := rect{area.x + mw + g, area.y, area.w - mw - g, area.h}
		n := len(list) - 1
		rh := (stack.h - (n-1)*g) / n

		for i, c := range list[1:] {
			m.place(c, rect{stack.x, stack.y + i*(rh+g), stack.w, rh})
		}
	}
}

func (m *Manager) place(c *Client, r rect) {
	c.r = r
	_ = m.conn.MoveResize(c.frame, r.x, r.y, r.w, r.h)
}

func (m *Manager) updateBorders() {
	for _, c := range m.clients {
		// maybe a cache here
		if c.ws != m.current {
			continue
		}

		color := m.pal.BorderInactive
		if c == m.focused {
			color = m.pal.BorderActive
		}

		_ = m.conn.SetBorderColor(c.frame, color)
	}
}

func (m *Manager) focusClient(c *Client) {
	if c == nil {
		return
	}

	m.focused = c
	_ = m.conn.SetInputFocus(c.win, x11.RevertToPointerRoot)
	m.updateBorders()
	m.redrawBar()
}

func (m *Manager) hide(c *Client) { _ = m.conn.UnmapWindow(c.frame) }
func (m *Manager) show(c *Client) { _ = m.conn.MapWindow(c.frame) }

func (m *Manager) geometry(win xproto.Window) (rect, error) {
	x, y, w, h, err := m.conn.GetGeometry(win)
	if err != nil {
		return rect{}, err
	}

	return rect{x, y, w, h}, nil
}
