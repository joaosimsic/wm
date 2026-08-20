package wm

import (
	"os/exec"

	"github.com/jezek/xgb/xproto"
)

func (m *Manager) spawn() error {
	cmd := exec.Command(m.cfg.Terminal)
	return cmd.Start()
}

func (m *Manager) closeClient() error {
	c := m.focused
	if c == nil {
		return nil
	}

	ev := xproto.ClientMessageEvent{
		Format: 32,
		Window: c.win,
		Type:   m.atoms.wmProtocols,
		Data: xproto.ClientMessageDataUnionData32New([]uint32{
			uint32(m.atoms.wmDelete),
			uint32(xproto.TimeCurrentTime),
			0, 0, 0,
		}),
	}

	return m.conn.SendEvent(c.win, ev)
}

func (m *Manager) quit() error {
	select {
	case <-m.done:
	default:
		close(m.done)
	}
	return nil
}

func (m *Manager) focusNext() error { return m.focusStep(1) }

func (m *Manager) focusPrev() error { return m.focusStep(-1) }

func (m *Manager) focusStep(dir int) error {
	l := m.current.list
	if len(l) == 0 {
		return nil
	}

	i := m.current.index(m.focused)
	i = max(0, i)
	i = (i + dir + len(l)) % len(l)

	m.focusClient(l[i])
	return nil
}

func (m *Manager) swap(delta int) error {
	i := m.current.index(m.focused)
	l := m.current.list
	j := i + delta
	if j < 0 || j >= len(l)-1 {
		return nil
	}

	l[i], l[j] = l[j], l[i]
	m.arrange()
	return nil
}

func (m *Manager) swapNext() error {
	return m.swap(1)
}

func (m *Manager) swapPrev() error {
	return m.swap(-1)
}

func (m *Manager) focusMaster() error {
	if len(m.current.list) > 0 {
		m.focusClient(m.current.list[0])
	}
	return nil
}

func (m *Manager) growMaster() error {
	m.ratio = clamp(m.ratio + 0.05)
	m.arrange()
	return nil
}

func (m *Manager) shrinkMaster() error {
	m.ratio = clamp(m.ratio - 0.05)
	m.arrange()
	return nil
}

func (m *Manager) switchTo(id int) error {
	if id < 1 || id > len(m.workspaces) || m.workspaces[id-1] == m.current {
		return nil
	}

	for _, c := range m.current.list {
		m.hide(c)
	}

	m.current = m.workspaces[id-1]
	for _, c := range m.current.list {
		m.show(c)
	}

	m.focused = nil
	m.focusFirst()
	m.arrange()
	return nil
}

func (m *Manager) moveToWorkspace(id int) error {
	c := m.focused
	if c == nil || id < 1 || id > len(m.workspaces) {
		return nil
	}

	ws := m.workspaces[id-1]
	if ws == c.ws {
		return nil
	}

	m.hide(c)
	c.ws.remove(c)
	c.ws = ws
	ws.add(c)
	m.focused = nil
	m.focusFirst()
	m.arrange()
	return nil
}

func (m *Manager) focusFirst() {
	if len(m.current.list) == 0 {
		m.focused = nil
		m.updateBorders()
		m.redrawBar()
		return
	}
	m.focusClient(m.current.list[0])
}

func clamp(r float64) float64 {
	if r < 0.05 {
		return 0.05
	}
	if r > 0.95 {
		return 0.95
	}
	return r
}
