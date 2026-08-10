package wm

import (
	"fmt"

	"github.com/BurntSushi/xgb/xproto"

	"wm/pkg/layout"
)

func (wm *WM) switchWorkspace(ws int) {
	if ws == wm.currentWS || ws < 0 || ws >= len(wm.workspaces) {
		return
	}

	for _, mw := range wm.windows {
		xproto.UnmapWindow(wm.xu, mw.Frame)
	}

	wm.currentWS = ws

	newWS := wm.workspaces[ws]
	if newWS.Root != nil {
		for _, mw := range newWS.Root.LeafWindows() {
			xproto.MapWindow(wm.xu, mw.Frame)
			xproto.MapWindow(wm.xu, mw.Client)
		}
	}

	wm.focused = nil
	if newWS.Root != nil {
		if f := newWS.Root.FindFocused(); f != nil && f.Window != nil {
			wm.setFocus(f.Window)
		}
	}

	wm.tileWorkspace()
	wm.updateBorders()
	wm.ewmhCurrentDesktopSet(uint(ws))
	wm.bar.Update(wm.mode, wm.currentWS, wm.cmdBuffer, wm.focusedTitle())
}

func (wm *WM) tileWorkspace() {
	ws := wm.workspaces[wm.currentWS]
	if ws.Root == nil {
		return
	}

	sw := int(wm.screen.WidthInPixels)
	sh := int(wm.screen.HeightInPixels)
	bh := wm.conf.BarHeight

	area := layout.Rect{X: 0, Y: 0, W: sw, H: sh - bh}
	bw := wm.conf.BorderWidth
	gap := wm.conf.Gap

	ws.Root.Tile(area, bw, gap)
	wm.applyLayout(ws.Root)
}

func (wm *WM) scanExistingWindows() error {
	tree, err := xproto.QueryTree(wm.xu, wm.root).Reply()
	if err != nil {
		return fmt.Errorf("query tree: %w", err)
	}

	for _, child := range tree.Children {
		attrs, err := xproto.GetWindowAttributes(wm.xu, child).Reply()
		if err != nil {
			continue
		}
		if attrs.MapState == xproto.MapStateViewable && !attrs.OverrideRedirect {
			wm.manageWindow(child)
		}
	}

	return nil
}
