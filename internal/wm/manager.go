package wm

import (
	"fmt"

	"github.com/jezek/xgb/xproto"
	"github.com/joaosimsic/wm/internal/config"
	"github.com/joaosimsic/wm/internal/theme"
	"github.com/joaosimsic/wm/internal/x11"
	"go.uber.org/zap"
)

const oneIndexedWs = 1

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
	cursor   xproto.Cursor
	bindings map[xproto.Keycode]map[uint16]func(*Manager) error
	ratio    float64
	done     chan struct{}
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
		done:     make(chan struct{}),
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
	for {
		select {
		case <-m.done:
			return nil
		default:
		}

		ev, xerr := m.conn.WaitForEvent()
		if ev == nil && xerr == nil {
			m.log.Error("x11 Connection closed")
			return fmt.Errorf("x11 connection closed")
		}
		if xerr != nil {
			m.log.Warn("x11 error", zap.Error(xerr))
			continue
		}
		if ev == nil {
			continue
		}

		m.log.Debug("event", zap.String("type", fmt.Sprintf("%T", ev)))
		m.handle(ev)
	}
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
		m.workspaces = append(m.workspaces, &Workspace{id: i + oneIndexedWs})
	}
	m.current = m.workspaces[0]

	XCLeftPtr := uint16(68)
	cursor, err := m.conn.CreateFontCursor(XCLeftPtr, XCLeftPtr)
	if err != nil {
		return err
	}
	m.cursor = cursor

	bar, err := newBar(m.conn, m.cfg, m.pal, m.cursor)
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
		xproto.EventMaskPropertyChange

	if err := m.conn.SelectRootEvents(uint32(mask)); err != nil {
		return err
	}

	if err := m.conn.SetCursor(m.conn.RootWindow(), cursor); err != nil {
		return err
	}

	if err := m.adoptExisting(); err != nil {
		return err
	}

	m.focusFirst()
	m.redrawBar()
	return nil
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

	for wsID := range m.cfg.Workspaces {
		id := wsID + oneIndexedWs
		a[fmt.Sprintf("workspace_%d", id)] = func(m *Manager) error { return m.switchTo(id) }
		a[fmt.Sprintf("move_to_workspace_%d", id)] = func(m *Manager) error { return m.moveToWorkspace(id) }
	}

	return a
}

func (m *Manager) manage(win xproto.Window) error {
	r, err := m.geometry(win)
	if err != nil {
		return err
	}

	style := FrameStyle{
		BorderWidth:    m.cfg.BorderWidth,
		Background:     m.pal.BarBg,
		BorderActive:   m.pal.BorderActive,
		BorderInactive: m.pal.BorderInactive,
		Cursor:         m.cursor,
	}

	frame, err := newFrame(m.conn, style, r)
	if err != nil {
		return err
	}

	if err := m.conn.ReparentWindow(win, frame, m.cfg.BorderWidth, m.cfg.BorderWidth); err != nil {
		return err
	}

	if err := m.conn.SetBorderWidth(win, 0); err != nil {
		return err
	}

	if err := m.conn.SelectWindowEvents(win, xproto.EventMaskPropertyChange); err != nil {
		return err
	}

	if err := m.setWMProtocols(win); err != nil {
		return err
	}

	c := &Client{
		win:   win,
		frame: frame,
		ws:    m.current,
	}
	m.clients[win] = c
	m.current.add(c)

	m.updateTitle(c)

	if err := m.conn.MapWindow(frame); err != nil {
		return err
	}

	if err := m.conn.MapWindow(win); err != nil {
		return err
	}

	m.focusClient(c)
	m.arrange()
	return nil
}
