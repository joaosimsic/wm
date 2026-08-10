package wm

import (
	"github.com/BurntSushi/xgb"
	"github.com/BurntSushi/xgb/xproto"
)

func (wm *WM) handleEvent(ev xgb.Event) {
	switch e := ev.(type) {
	case xproto.KeyPressEvent:
		wm.handleKeyPress(e)
	case xproto.MapRequestEvent:
		wm.handleMapRequest(e)
	case xproto.ConfigureRequestEvent:
		wm.handleConfigureRequest(e)
	case xproto.DestroyNotifyEvent:
		wm.handleDestroyNotify(e)
	case xproto.UnmapNotifyEvent:
		wm.handleUnmapNotify(e)
	case xproto.ButtonPressEvent:
		wm.handleButtonPress(e)
	case xproto.ExposeEvent:
		if xproto.Window(e.Window) == wm.bar.Window {
			wm.bar.Redraw()
		}
	case xproto.ClientMessageEvent:
		wm.handleClientMessage(e)
	case xproto.PropertyNotifyEvent:
		wm.handlePropertyNotify(e)
	case xproto.EnterNotifyEvent:
		wm.handleEnterNotify(e)
	}
}

func (wm *WM) handleEnterNotify(ev xproto.EnterNotifyEvent) {
	mw, ok := wm.frames[ev.Event]
	if !ok {
		return
	}
	if mw == wm.focused {
		return
	}
	wm.setFocusByWindow(mw.Client)
}

func (wm *WM) handleMapRequest(ev xproto.MapRequestEvent) {
	if _, exists := wm.windows[ev.Window]; exists {
		xproto.MapWindow(wm.xu, ev.Window)
		return
	}
	wm.manageWindow(xproto.Window(ev.Window))
}

func (wm *WM) handleConfigureRequest(ev xproto.ConfigureRequestEvent) {
	vals := make([]uint32, 0)
	mask := uint16(0)

	appendIf := func(flag uint16, val uint32) {
		if ev.ValueMask&flag != 0 {
			mask |= flag
			vals = append(vals, val)
		}
	}

	appendIf(xproto.ConfigWindowX, uint32(ev.X))
	appendIf(xproto.ConfigWindowY, uint32(ev.Y))
	appendIf(xproto.ConfigWindowWidth, uint32(ev.Width))
	appendIf(xproto.ConfigWindowHeight, uint32(ev.Height))
	appendIf(xproto.ConfigWindowBorderWidth, uint32(ev.BorderWidth))
	appendIf(xproto.ConfigWindowSibling, uint32(ev.Sibling))
	appendIf(xproto.ConfigWindowStackMode, uint32(ev.StackMode))

	xproto.ConfigureWindow(wm.xu, ev.Window, mask, vals)
}

func (wm *WM) handleDestroyNotify(ev xproto.DestroyNotifyEvent) {
	wm.removeWindow(xproto.Window(ev.Window))
}

func (wm *WM) handleUnmapNotify(ev xproto.UnmapNotifyEvent) {
	wm.removeWindow(xproto.Window(ev.Window))
}

func (wm *WM) handleButtonPress(ev xproto.ButtonPressEvent) {
	if ev.Child != 0 {
		wm.setFocusByWindow(xproto.Window(ev.Child))
	}
}

func (wm *WM) handleClientMessage(ev xproto.ClientMessageEvent) {
	_ = ev
}

func (wm *WM) handlePropertyNotify(ev xproto.PropertyNotifyEvent) {
	reply, err := xproto.GetAtomName(wm.xu, ev.Atom).Reply()
	if err != nil {
		return
	}
	name := reply.Name
	if name == "WM_NAME" || name == "_NET_WM_NAME" {
		if mw, ok := wm.windows[ev.Window]; ok {
			mw.Title = wm.getWindowTitle(ev.Window)
		}
	}
}
