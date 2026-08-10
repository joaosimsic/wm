// Package wm implements the core window manager.
package wm

import (
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/BurntSushi/xgb"
	"github.com/BurntSushi/xgb/xproto"

	"wm/pkg/bar"
	"wm/pkg/config"
	"wm/pkg/layout"
	"wm/pkg/types"
)

type WM struct {
	xu     *xgb.Conn
	screen *xproto.ScreenInfo
	root   xproto.Window

	mode       types.Mode
	workspaces []*Workspace
	currentWS  int
	focused    *layout.ManagedWindow

	bar      *bar.Bar
	keyBinds []types.KeyBind

	conf           config.Config
	barBg          uint32
	barFg          uint32
	barActiveBg    uint32
	borderActive   uint32
	borderInactive uint32

	windows   map[xproto.Window]*layout.ManagedWindow
	frames    map[xproto.Window]*layout.ManagedWindow
	cmdBuffer string

	running   bool
	connAlive bool
	closeOnce sync.Once

	ewmhAtoms map[string]xproto.Atom

	pendingSplit bool
	pendingDir   layout.ContainerType
}

type Workspace struct {
	Root *layout.Container
}

func New() (*WM, error) {
	conf := config.Default()

	if _, err := toml.DecodeFile("config.toml", &conf); err != nil && !os.IsNotExist(err) {
		return nil, fmt.Errorf("parse config: %w", err)
	}

	display := os.Getenv("DISPLAY")
	if display == "" {
		display = ":0"
	}

	conn, err := xgb.NewConnDisplay(display)
	if err != nil {
		return nil, fmt.Errorf("connect to X: %w", err)
	}

	screen := xproto.Setup(conn).DefaultScreen(conn)
	root := screen.Root

	wm := &WM{
		xu:         conn,
		screen:     screen,
		root:       root,
		mode:       types.ModeNormal,
		currentWS:  0,
		conf:       conf,
		windows:    make(map[xproto.Window]*layout.ManagedWindow),
		frames:     make(map[xproto.Window]*layout.ManagedWindow),
		workspaces: make([]*Workspace, 10),
		running:    true,
		connAlive:  true,
		ewmhAtoms:  make(map[string]xproto.Atom),
	}

	for i := range wm.workspaces {
		wm.workspaces[i] = &Workspace{}
	}

	if err := wm.initEWMH(); err != nil {
		return nil, fmt.Errorf("init ewmh: %w", err)
	}

	if err := wm.allocColors(); err != nil {
		return nil, fmt.Errorf("alloc colors: %w", err)
	}

	if err := wm.initRoot(); err != nil {
		return nil, fmt.Errorf("init root: %w", err)
	}

	if err := wm.setupKeybinds(); err != nil {
		return nil, fmt.Errorf("setup keybinds: %w", err)
	}

	var errBar error
	wm.bar, errBar = bar.New(
		wm.xu,
		wm.screen.WidthInPixels,
		wm.screen.HeightInPixels,
		uint16(wm.conf.BarHeight),
		wm.barBg, wm.barFg, wm.barActiveBg,
		wm.root,
		wm.currentWS,
		wm.mode,
	)
	if errBar != nil {
		return nil, fmt.Errorf("create bar: %w", errBar)
	}

	if err := wm.scanExistingWindows(); err != nil {
		return nil, fmt.Errorf("scan windows: %w", err)
	}

	return wm, nil
}

func (wm *WM) Run() error {
	fmt.Fprintf(os.Stderr, "wm: starting on display %s (%dx%d)\n",
		os.Getenv("DISPLAY"), wm.screen.WidthInPixels, wm.screen.HeightInPixels)
	wm.bar.Redraw()
	wm.tileWorkspace()
	fmt.Fprintf(os.Stderr, "wm: entering event loop\n")

	events := make(chan xgb.Event, 16)
	errors := make(chan xgb.Error, 16)
	go wm.readEvents(events, errors)

	for wm.running {
		select {
		case ev, ok := <-events:
			if !ok {
				wm.running = false
				wm.connAlive = false
				continue
			}
			wm.handleEvent(ev)
		case err := <-errors:
			if wm.running {
				fmt.Fprintf(os.Stderr, "protocol error: %v\n", err)
			}
		case <-time.After(5 * time.Millisecond):
		}
	}

	wm.shutdown()
	return nil
}

func (wm *WM) readEvents(events chan<- xgb.Event, errors chan<- xgb.Error) {
	for {
		ev, err := wm.xu.WaitForEvent()
		if ev == nil && err == nil {
			close(events)
			return
		}
		if err != nil {
			errors <- err
			continue
		}
		events <- ev
	}
}

func (wm *WM) Stop() {
	wm.running = false
}

func (wm *WM) shutdown() {
	wm.closeOnce.Do(func() {
		wm.running = false
		if !wm.connAlive {
			return
		}
		for _, mw := range wm.windows {
			xproto.ReparentWindow(wm.xu, mw.Client, wm.root, 0, 0)
			xproto.DestroyWindow(wm.xu, mw.Frame)
		}
		xproto.DestroyWindow(wm.xu, wm.bar.Window)
		wm.xu.Close()
	})
}
