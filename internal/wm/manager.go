package wm

import (
	"fmt"

	"github.com/jezek/xgb/xproto"
	"github.com/joaosimsic/wm/internal/config"
	"github.com/joaosimsic/wm/internal/theme"
	"github.com/joaosimsic/wm/internal/x11"
	"go.uber.org/zap"
)

type Manager struct {
	conn *x11.Connection
	cfg  *config.Config
	pal  *theme.Palette
	log  *zap.Logger

	atoms   atoms
	kbd     *x11.Keyboard
	modMask uint16

	clients    map[xproto.Window]*Client
	workspaces []*Workspace
	current    *Workspace
	focused    *Client

	bar      *Bar
	bindings map[xproto.Keycode]map[uint16]func(*Manager) error
	ratio    float64
	running  bool
}

func New(conn *x11.Connection, cfg *config.Config, pal *theme.Palette, log *zap.Logger) (*Manager, error) {
	m := &Manager{
		conn:     conn,
		cfg:      cfg,
		pal:      pal,
		log:      log,
		clients:  make(map[xproto.Window]*Client),
		bindings: make(map[xproto.Keycode]map[uint16]func(*Manager) error),
		ratio:    cfg.SplitRatio,
	}

	if err := m.setup(); err != nil {
		m.Close()
		return nil, err
	}

	return m, nil
}

func (m *Manager) Close() {
	if m.bar != nil {
		m.bar.Close()
		m.bar = nil
	}

	if m.conn != nil && m.kbd != nil {
		for code, byMask := range m.bindings {
			for mods := range byMask {
				_ = m.conn.UngrabAllCombos(m.conn.RootWindow(), code, mods, m.kbd.NumLock)
			}
		}
	}

	for _, c := range m.clients {
		_ = m.conn.DestroyWindow(c.frame)
	}

	clear(m.clients)
	clear(m.bindings)
}

func (m *Manager) Run() error {
	m.running = true

	for m.running {
		ev, xerr := m.conn.WaitForEvent()
		if xerr != nil {
			m.log.Warn("x11 error", zap.Error(xerr))
			continue
		}
		if ev == nil {
			continue
		}

		m.handle(ev)
	}

	return nil
}

func (m *Manager) setup() error {
	if err := m.initAtoms(); err != nil {
		return err
	}

	kbd, err := m.conn.Keyboard()
	if err != nil {
		return err
	}
	m.kbd = kbd

	mods, err := x11.ParseModifierString(m.cfg.Mod)
	if err != nil {
		return err
	}
	m.modMask = mods

	for i := range m.cfg.Workspaces {
		m.workspaces = append(m.workspaces, &Workspace{id: i})
	}
	m.current = m.workspaces[0]

	bar, err := newBar(m.conn, m.cfg, m.pal)
	if err != nil {
		return err
	}
	m.bar = bar

	if err := m.bindKeys(); err != nil {
		return err
	}

	mask := xproto.EventMaskSubstructureRedirect |
		xproto.EventMaskSubstructureNotify |
		xproto.EventMaskStructureNotify |
		xproto.EventMaskStructureNotify

	if err := m.conn.SelectRootEvents(uint32(mask)); err != nil {
		return err
	}

	return m.adoptExisting()
}

func (m *Manager) adoptExisting() error {
	_, children, err := m.conn.QueryTree(m.conn.RootWindow())
	if err != nil {
		return err
	}

	for _, win := range children {
		attrs, err := m.conn.GetAttributes(win)
		if err != nil {
			continue
		}
		if attrs.OverrideRedirect || attrs.MapState == xproto.MapStateUnmapped {
			continue
		}

		if err := m.manage(win); err != nil {
			m.log.Warn("adopt", zap.Uint32("win", uint32(win)), zap.Error(err))
		}
	}

	return nil
}

func (m *Manager) bindKeys() error {
	actions := m.actions()

	for name, spec := range m.cfg.Keys {
		fn, ok := actions[name]
		if !ok {
			m.log.Warn("unknown action", zap.String("action", name))
			continue
		}

		c, err := parseCombo(spec)
		if err != nil {
			return err
		}

		mods := m.modMask | c.mods

		keycodes, ok := m.kbd.Keymap.KeycodesForKeysym(c.sym)
		if !ok {
			return fmt.Errorf("action %q: no keycode for keysym %q", name, spec)
		}

		for _, code := range keycodes {
			if err := m.conn.GrabAllCombos(m.conn.RootWindow(), code, mods, m.kbd.NumLock); err != nil {
				return fmt.Errorf("action %q: %w", name, err)
			}

			if m.bindings[code] == nil {
				m.bindings[code] = make(map[uint16]func(*Manager) error)
			}

			m.bindings[code][mods] = fn
		}
	}

	return nil
}

func (m *Manager) actions() map[string]func(*Manager) error {
	a := map[string]func(*Manager) error{
		"spawn":         (*Manager).spawn,
		"close":         (*Manager).closeClient,
		"quit":          (*Manager).quit,
		"focus_next":    (*Manager).focusNext,
		"focus_prev":    (*Manager).focusPrev,
		"swap_next":     (*Manager).swapNext,
		"swap_prev":     (*Manager).swapPrev,
		"focus_master":  (*Manager).focusMaster,
		"grow_master":   (*Manager).growMaster,
		"shrink_master": (*Manager).shrinkMaster,
	}

	for wsId := range m.cfg.Workspaces {
		a[fmt.Sprintf("workspace_%d", wsId)] = func(m *Manager) error { return m.switchTo(wsId) }
		a[fmt.Sprintf("move_to_workspace_%d", wsId)] = func(m *Manager) error { return m.moveTpWorkspace(wsId) }
	}

    return a
}

func (m *Manager) manage(win xproto.Window) error {
    geo, err := m.conn.GetGeometry(win)
    if err != nil {
        return err
    }

    style := FrameStyle{
        BorderWidth: m.cfg.BorderWidth,
        Background: m.pal.BarBg,
        BorderActive: m.pal.BorderActive,
        BorderInactive: m.pal.BorderInactive,
    }

    frame, err := newFrame(m.conn, style, int(geo.X), int(geo.Y), int(geo.Width), int(geo.Height))
    if err != nil {
        return err
    }

    if err := m.conn.ReparentWindow(win, frame, 0, 0); err != nil {
        return err
    }

    return nil
}
