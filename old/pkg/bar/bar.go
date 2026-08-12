// Package bar renders the status bar.
package bar

import (
	"fmt"

	"github.com/BurntSushi/xgb"
	"github.com/BurntSushi/xgb/xproto"

	"wm/pkg/types"
)

type Bar struct {
	Window   xproto.Window
	Gc       xproto.Gcontext
	Font     xproto.Font
	BgColor  uint32
	FgColor  uint32
	ActiveBg uint32

	activeWS    int
	mode        types.Mode
	cmdText     string
	screenWidth uint16
	barHeight   uint16

	conn *xgb.Conn
}

func New(conn *xgb.Conn, sw, sh uint16, bh uint16, barBg, barFg, barActiveBg uint32, root xproto.Window, activeWS int, mode types.Mode) (*Bar, error) {
	barWin, err := newWindow(conn, sw, sh, bh, barBg, root)
	if err != nil {
		return nil, err
	}

	font, fontErr := FindUsableFont(conn)

	gc, err := newGC(conn, barWin, font, fontErr, barFg, barBg)
	if err != nil {
		return nil, err
	}

	err = xproto.MapWindowChecked(conn, barWin).Check()
	if err != nil {
		return nil, fmt.Errorf("map bar window: %w", err)
	}

	return &Bar{
		Window:      barWin,
		Gc:          gc,
		Font:        font,
		BgColor:     barBg,
		FgColor:     barFg,
		ActiveBg:    barActiveBg,
		conn:        conn,
		screenWidth: sw,
		barHeight:   bh,
		activeWS:    activeWS,
		mode:        mode,
	}, nil
}

func newWindow(conn *xgb.Conn, sw, sh uint16, bh uint16, barBg uint32, root xproto.Window) (xproto.Window, error) {
	barWin, err := xproto.NewWindowId(conn)
	if err != nil {
		return 0, fmt.Errorf("new window id: %w", err)
	}

	xproto.CreateWindow(conn, 0, barWin, root,
		int16(0), int16(sh)-int16(bh), sw, bh, 0,
		xproto.WindowClassInputOutput,
		xproto.Setup(conn).DefaultScreen(conn).RootVisual,
		xproto.CwBackPixel|xproto.CwOverrideRedirect|xproto.CwEventMask,
		[]uint32{
			barBg,
			1,
			uint32(xproto.EventMaskExposure | xproto.EventMaskButtonPress),
		},
	)

	return barWin, nil
}

func newGC(conn *xgb.Conn, barWin xproto.Window, font xproto.Font, fontErr error, barFg, barBg uint32) (xproto.Gcontext, error) {
	gc, err := xproto.NewGcontextId(conn)
	if err != nil {
		return 0, fmt.Errorf("new gc id: %w", err)
	}
	gcMask := uint32(xproto.GcForeground | xproto.GcBackground)
	gcVals := []uint32{barFg, barBg}
	if fontErr == nil {
		gcMask |= xproto.GcFont
		gcVals = append(gcVals, uint32(font))
	}
	xproto.CreateGC(conn, gc, xproto.Drawable(barWin), gcMask, gcVals)
	return gc, nil
}

func FindUsableFont(conn *xgb.Conn) (xproto.Font, error) {
	candidates := []string{"fixed", "6x13", "9x15"}

	for _, name := range candidates {
		font, ferr := xproto.NewFontId(conn)
		if ferr != nil {
			continue
		}
		cerr := xproto.OpenFontChecked(conn, font, uint16(len(name)), name).Check()
		if cerr != nil {
			xproto.CloseFont(conn, font)
			continue
		}
		return font, nil
	}

	return 0, fmt.Errorf("no usable font found")
}

func (b *Bar) Redraw() {
	conn := b.conn

	xproto.ClearArea(conn, false, b.Window, 0, 0, b.screenWidth, b.barHeight)

	if b.Font == 0 {
		conn.Sync()
		return
	}

	text := b.barText()
	if text == "" {
		conn.Sync()
		return
	}

	_ = xproto.ImageText8(conn, byte(len(text)), xproto.Drawable(b.Window), b.Gc, 4, 15, text)
	conn.Sync()
}

func (b *Bar) Update(mode types.Mode, ws int, cmdText string, _ string) {
	b.mode = mode
	b.activeWS = ws
	b.cmdText = cmdText
	b.Redraw()
}

func (b *Bar) barText() string {
	if b.mode == types.ModeCommand {
		return ":" + b.cmdText
	}
	return fmt.Sprintf("[NORMAL]  WS:%d", b.activeWS+1)
}
