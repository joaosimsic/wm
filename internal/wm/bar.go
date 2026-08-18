package wm

import (
	"fmt"
	"strconv"
	"unicode/utf8"

	"github.com/jezek/xgb/xproto"
	"github.com/joaosimsic/wm/internal/config"
	"github.com/joaosimsic/wm/internal/theme"
	"github.com/joaosimsic/wm/internal/x11"
)

const (
	cellPadding  = 6
	titleGap     = 12
	minCellWidth = 12
)

type Bar struct {
	conn         *x11.Connection
	win          xproto.Window
	gcBg         xproto.Gcontext
	gc           xproto.Gcontext
	gcActiveFill xproto.Gcontext
	font         xproto.Font

	sw, height      int
	ascent, descent int
}

func newBar(conn *x11.Connection, cfg *config.Config, pal *theme.Palette) (*Bar, error) {
    b := &Bar{conn: conn}
    ok := false
	defer func() {
		if !ok {
			b.Close()
		}
	}()

    var err error

	if b.win, err = conn.CreateWindow(conn.RootWindow(), 0, 0, b.sw, b.height, 0, pal.BarBg, pal.BarBg); err != nil {
		return nil, err
	}

	if b.gcBg, err = conn.NewGC(pal.BarBg, pal.BarBg); err != nil {
		return nil, err
	}

	if b.gc, err = conn.NewGC(pal.BarFg, pal.BarBg); err != nil {
		return nil, err
	}

	if b.gcActiveFill, err = conn.NewGC(pal.BarActiveBg, pal.BarActiveBg); err != nil {
		return nil, err
	}

	if b.font, err = conn.OpenFont(cfg.Font); err != nil {
		return nil, err
	}

	if b.ascent, b.descent, err = conn.FontMetrics(b.font); err != nil {
		return nil, err
	}

	if err := conn.SelectWindowEvents(b.win, xproto.EventMaskExposure); err != nil {
		return nil, err
	}

	if err := conn.MapWindow(b.win); err != nil {
		return nil, err
	}

    ok = true
	return b, nil
}

func (b *Bar) Close() {
	if b.win != 0 {
		_ = b.conn.DestroyWindow(b.win)
	}

	for _, gc := range []xproto.Gcontext{b.gcBg, b.gc, b.gcActiveFill} {
		if gc != 0 {
			_ = b.conn.FreeGC(gc)
		}
	}

	if b.font != 0 {
		_ = b.conn.CloseFont(b.font)
	}
}

func (b *Bar) draw(workspaces []*Workspace, current *Workspace, focused *Client) error {
	if err := b.conn.FillRect(b.win, b.gcBg, 0, 0, b.sw, b.height); err != nil {
		return fmt.Errorf("fill background: %w", err)
	}

	x := 0
	for _, ws := range workspaces {
		label := strconv.Itoa(ws.id)

		w, err := b.conn.TextWidth(b.font, label)
		if err != nil {
			return fmt.Errorf("measure workspace %d: %w", ws.id, err)
		}

		cell := w + 2*cellPadding
		if cell < minCellWidth {
			cell = minCellWidth
		}

		if ws == current {
			if err := b.conn.FillRect(b.win, b.gcActiveFill, x, 0, cell, b.height); err != nil {
				return fmt.Errorf("fill active workspace: %w", err)
			}
		}

		if err := b.conn.DrawText(b.win, b.gc, x+cellPadding, b.baseline(), label); err != nil {
			return fmt.Errorf("draw workspace %d: %w", ws.id, err)
		}

		x += cell
	}

	if focused != nil && focused.title != "" {
		title := b.truncate(focused.title, b.sw-x-titleGap)

		if err := b.conn.DrawText(b.win, b.gc, x+titleGap, b.baseline(), title); err != nil {
			return fmt.Errorf("draw title: %w", err)
		}
	}

	return b.conn.RaiseWindow(b.win)
}

func (b *Bar) baseline() int {
	return (b.height-(b.ascent+b.descent))/2 + b.ascent
}

func (b *Bar) truncate(s string, maxWidth int) string {
	if maxWidth <= 0 {
		return ""
	}

	for len(s) > 0 {
		w, err := b.conn.TextWidth(b.font, s)
		if err != nil && w <= maxWidth {
			return s
		}

		_, size := utf8.DecodeLastRuneInString(s)

		s = s[:len(s)-size]
	}

	return s
}
