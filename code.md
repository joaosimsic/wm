# WM core — implementation code (roadmap steps 2–4)

Decisions: global `mod` + per-action combos like `"Shift+q"`, 8-bit bar text, master left, fixed 0.05 resize step, combo parsing in `wm/keys.go`.

## 1. `internal/x11` additions

`window.go` — append:

```go
func (c *Connection) GetGeometry(win xproto.Window) (x, y, w, h int, err error) {
	reply, err := xproto.GetGeometry(c.conn, xproto.Drawable(win)).Reply()
	if err != nil {
		return 0, 0, 0, 0, err
	}
	return int(reply.X), int(reply.Y), int(reply.Width), int(reply.Height), nil
}

func (c *Connection) GetAttributes(win xproto.Window) (*xproto.GetWindowAttributesReply, error) {
	return xproto.GetWindowAttributes(c.conn, win).Reply()
}
```

`atoms.go` — append (adds `xgb` import):

```go
func (c *Connection) ChangeProperty32(win xproto.Window, prop, typ xproto.Atom, values []uint32) error {
	data := make([]byte, 4*len(values))
	for i, v := range values {
		data[i*4] = byte(v)
		data[i*4+1] = byte(v >> 8)
		data[i*4+2] = byte(v >> 16)
		data[i*4+3] = byte(v >> 24)
	}
	return xproto.ChangePropertyChecked(c.conn, xproto.PropModeReplace, win, prop, typ, 32, uint32(len(values)), data).Check()
}

func (c *Connection) SendEvent(win xproto.Window, ev xgb.Event) error {
	return xproto.SendEventChecked(c.conn, false, win, 0, string(ev.Bytes())).Check()
}
```

## 2. `internal/config`

`config.go` — add fields:

```go
	Mod        string            `toml:"mod"`
	SplitRatio float64           `toml:"split_ratio"`
	Workspaces int               `toml:"workspaces"`
	Keys       map[string]string `toml:"keys"`
```

`defaults.go`:

```go
package config

import "fmt"

func Default() Config {
	return Config{
		Terminal:    "xterm",
		Font:        "fixed",
		BorderWidth: 2,
		Gap:         0,
		BarHeight:   22,
		Mod:         "Mod4",
		SplitRatio:  0.5,
		Workspaces:  9,
		Keys:        DefaultKeys(),
		Colors:      DefaultColors(),
	}
}

func DefaultKeys() map[string]string {
	keys := map[string]string{
		"spawn":         "Return",
		"close":         "q",
		"quit":          "Shift+q",
		"focus_next":    "j",
		"focus_prev":    "k",
		"swap_next":     "Shift+j",
		"swap_prev":     "Shift+k",
		"focus_master":  "m",
		"grow_master":   "l",
		"shrink_master": "h",
	}
	for i := 1; i <= 9; i++ {
		keys[fmt.Sprintf("workspace_%d", i)] = fmt.Sprintf("%d", i)
		keys[fmt.Sprintf("move_to_workspace_%d", i)] = fmt.Sprintf("Shift+%d", i)
	}
	return keys
}
```

`validate.go` — append:

```go
	if cfg.Mod == "" {
		errs = append(errs, fmt.Errorf("mod missing, using default %q", def.Mod))
		cfg.Mod = def.Mod
	}

	if cfg.SplitRatio <= 0 || cfg.SplitRatio >= 1 {
		errs = append(errs, fmt.Errorf("split_ratio must be between 0 and 1: %f, using default %f", cfg.SplitRatio, def.SplitRatio))
		cfg.SplitRatio = def.SplitRatio
	}

	if cfg.Workspaces < 1 {
		errs = append(errs, fmt.Errorf("workspaces must be at least 1: %d, using default %d", cfg.Workspaces, def.Workspaces))
		cfg.Workspaces = def.Workspaces
	}

	if len(cfg.Keys) == 0 {
		errs = append(errs, fmt.Errorf("keys missing, using defaults"))
		cfg.Keys = def.Keys
	}
```

## 3. `internal/wm` — new files

`keys.go`:

```go
package wm

import (
	"fmt"
	"strings"

	"github.com/jezek/xgb/xproto"
	"github.com/joaosimsic/wm/internal/x11"
)

var namedKeysyms = map[string]xproto.Keysym{
	"Return":    0xff0d,
	"Tab":       0xff09,
	"space":     0x20,
	"Escape":    0xff1b,
	"BackSpace": 0xff08,
	"Delete":    0xffff,
	"Home":      0xff50,
	"End":       0xff57,
	"Left":      0xff51,
	"Right":     0xff53,
	"Up":        0xff52,
	"Down":      0xff54,
	"Page_Up":   0xff55,
	"Page_Down": 0xff56,
}

func keysymFromName(name string) (xproto.Keysym, bool) {
	if len(name) == 1 {
		return xproto.Keysym(name[0]), true
	}
	s, ok := namedKeysyms[name]
	return s, ok
}

type combo struct {
	mods uint16
	sym  xproto.Keysym
}

func parseCombo(s string) (combo, error) {
	var c combo
	parts := strings.Split(s, "+")
	for _, p := range parts[:len(parts)-1] {
		m, err := x11.ParseModifiers(p)
		if err != nil {
			return c, fmt.Errorf("combo %q: %w", s, err)
		}
		c.mods |= m
	}
	sym, ok := keysymFromName(parts[len(parts)-1])
	if !ok {
		return c, fmt.Errorf("combo %q: unknown keysym %q", s, parts[len(parts)-1])
	}
	c.sym = sym
	return c, nil
}
```

`client.go`:

```go
package wm

import "github.com/jezek/xgb/xproto"

type Client struct {
	win   xproto.Window
	frame xproto.Window
	ws    *Workspace
	x, y, w, h int
	title string
}
```

`workspace.go`:

```go
package wm

type Workspace struct {
	id   int
	list []*Client
}

func (w *Workspace) add(c *Client) {
	w.list = append(w.list, c)
}

func (w *Workspace) remove(c *Client) {
	for i, e := range w.list {
		if e == c {
			w.list = append(w.list[:i], w.list[i+1:]...)
			return
		}
	}
}

func (w *Workspace) index(c *Client) int {
	for i, e := range w.list {
		if e == c {
			return i
		}
	}
	return -1
}
```

`layout.go`:

```go
package wm

import "github.com/joaosimsic/wm/internal/x11"

type rect struct {
	x, y, w, h int
}

func (m *Manager) arrange() {
	sw := int(m.conn.Screen().WidthInPixels)
	sh := int(m.conn.Screen().HeightInPixels)
	g := m.cfg.Gap
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
	c.x, c.y, c.w, c.h = r.x, r.y, r.w, r.h
	_ = m.conn.MoveResize(c.frame, r.x, r.y, r.w, r.h)
}

func (m *Manager) updateBorders() {
	for _, c := range m.clients {
		if c.ws != m.current {
			continue
		}
		pixel := m.pal.BorderInactive
		if c == m.focused {
			pixel = m.pal.BorderActive
		}
		_ = m.conn.SetBorderColor(c.frame, pixel)
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
```

`actions.go`:

```go
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
	m.running = false
	return nil
}

func (m *Manager) focusNext() error {
	return m.focusStep(1)
}

func (m *Manager) focusPrev() error {
	return m.focusStep(-1)
}

func (m *Manager) focusStep(dir int) error {
	l := m.current.list
	if len(l) == 0 {
		return nil
	}
	i := m.current.index(m.focused)
	if i < 0 {
		i = 0
	}
	i = (i + dir + len(l)) % len(l)
	m.focusClient(l[i])
	return nil
}

func (m *Manager) swapNext() error {
	i := m.current.index(m.focused)
	l := m.current.list
	if i < 0 || i >= len(l)-1 {
		return nil
	}
	l[i], l[i+1] = l[i+1], l[i]
	m.arrange()
	return nil
}

func (m *Manager) swapPrev() error {
	i := m.current.index(m.focused)
	l := m.current.list
	if i < 1 {
		return nil
	}
	l[i], l[i-1] = l[i-1], l[i]
	m.arrange()
	return nil
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
	if id < 1 || id > len(m.works) || m.works[id-1] == m.current {
		return nil
	}
	for _, c := range m.current.list {
		m.hide(c)
	}
	m.current = m.works[id-1]
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
	if c == nil || id < 1 || id > len(m.works) {
		return nil
	}
	ws := m.works[id-1]
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
```

`ewmh.go`:

```go
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
```

`bar.go`:

```go
package wm

import (
	"strconv"

	"github.com/jezek/xgb/xproto"
	"github.com/joaosimsic/wm/internal/config"
	"github.com/joaosimsic/wm/internal/theme"
	"github.com/joaosimsic/wm/internal/x11"
)

type Bar struct {
	win      xproto.Window
	gcBg     xproto.Gcontext
	gc       xproto.Gcontext
	gcActive xproto.Gcontext
	font     xproto.Font
}

func newBar(conn *x11.Connection, cfg *config.Config, pal *theme.Palette) (*Bar, error) {
	win, err := conn.CreateWindow(conn.RootWindow(), 0, 0, int(conn.Screen().WidthInPixels), cfg.BarHeight, 0, pal.BarBg, pal.BarBg)
	if err != nil {
		return nil, err
	}
	gcBg, err := conn.NewGC(pal.BarBg, pal.BarBg)
	if err != nil {
		return nil, err
	}
	gc, err := conn.NewGC(pal.BarFg, pal.BarBg)
	if err != nil {
		return nil, err
	}
	gcActive, err := conn.NewGC(pal.BarFg, pal.BarActiveBg)
	if err != nil {
		return nil, err
	}
	font, err := conn.OpenFont(cfg.Font)
	if err != nil {
		return nil, err
	}
	if err := conn.SelectWindowEvents(win, xproto.EventMaskExposure); err != nil {
		return nil, err
	}
	if err := conn.MapWindow(win); err != nil {
		return nil, err
	}
	return &Bar{win: win, gcBg: gcBg, gc: gc, gcActive: gcActive, font: font}, nil
}

func (m *Manager) redrawBar() {
	b := m.bar
	sw := int(m.conn.Screen().WidthInPixels)
	ty := m.cfg.BarHeight/2 + 4
	_ = m.conn.FillRect(b.win, b.gcBg, 0, 0, sw, m.cfg.BarHeight)
	x := 0
	for _, ws := range m.works {
		label := strconv.Itoa(ws.id)
		w, err := m.conn.TextWidth(b.font, label)
		if err != nil {
			w = 8
		}
		cell := w + 12
		gc := b.gc
		if ws == m.current {
			gc = b.gcActive
		}
		_ = m.conn.FillRect(b.win, gc, x, 0, cell, m.cfg.BarHeight)
		_ = m.conn.DrawText(b.win, gc, x+6, ty, label)
		x += cell
	}
	if m.focused != nil && m.focused.title != "" {
		_ = m.conn.DrawText(b.win, b.gc, x+12, ty, m.focused.title)
	}
}
```

`manager.go`:

```go
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

	clients map[xproto.Window]*Client
	works   []*Workspace
	current *Workspace
	focused *Client

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
		return nil, err
	}
	return m, nil
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

	for i := 1; i <= m.cfg.Workspaces; i++ {
		m.works = append(m.works, &Workspace{id: i})
	}
	m.current = m.works[0]

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
		xproto.EventMaskPropertyChange
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
	for i := 1; i <= m.cfg.Workspaces; i++ {
		id := i
		a[fmt.Sprintf("workspace_%d", i)] = func(m *Manager) error { return m.switchTo(id) }
		a[fmt.Sprintf("move_to_workspace_%d", i)] = func(m *Manager) error { return m.moveToWorkspace(id) }
	}
	return a
}

func (m *Manager) manage(win xproto.Window) error {
	geo, err := m.conn.GetGeometry(win)
	if err != nil {
		return err
	}
	style := FrameStyle{
		BorderWidth:    m.cfg.BorderWidth,
		Background:     m.pal.BarBg,
		BorderActive:   m.pal.BorderActive,
		BorderInactive: m.pal.BorderInactive,
	}
	frame, err := newFrame(m.conn, style, int(geo.X), int(geo.Y), int(geo.Width), int(geo.Height))
	if err != nil {
		return err
	}
	if err := m.conn.ReparentWindow(win, frame, 0, 0); err != nil {
		return err
	}
	_ = m.conn.SetBorderWidth(win, 0)
	_ = m.conn.SelectWindowEvents(win, xproto.EventMaskPropertyChange)
	if err := m.setWMProtocols(win); err != nil {
		return err
	}
	c := &Client{win: win, frame: frame, ws: m.current}
	m.clients[win] = c
	m.current.add(c)
	m.updateTitle(c)
	_ = m.conn.MapWindow(win)
	_ = m.conn.MapWindow(frame)
	m.focusClient(c)
	m.arrange()
	return nil
}

func (m *Manager) updateTitle(c *Client) {
	data, err := m.conn.GetProperty(c.win, m.atoms.netWMName, xproto.AtomString)
	if err != nil || len(data) == 0 {
		data, _ = m.conn.GetProperty(c.win, xproto.AtomWmName, xproto.AtomString)
	}
	c.title = string(data)
	m.redrawBar()
}
```

`events.go`:

```go
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
```

## 4. `cmd/wm/main.go`

Replace lines 39–44 with:

```go
	logger.Info("config loaded",
		zap.String("terminal", cfg.Terminal),
		zap.String("font", cfg.Font),
		zap.Uint32("border_active_pixel", palette.BorderActive),
	)

	m, err := wm.New(conn, cfg, palette, logger)
	if err != nil {
		logger.Fatal("failed to init wm", zap.Error(err))
	}

	if err := m.Run(); err != nil {
		logger.Fatal("wm error", zap.Error(err))
	}
	logger.Info("wm exited")
```

(add `"github.com/joaosimsic/wm/internal/wm"` import)

## 5. Verify

`make build && make vet`, then `make dev` + `DISPLAY=:1 xterm &` — check borders, split, gaps, focus, workspace switching, bar.
