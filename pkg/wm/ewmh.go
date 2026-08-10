package wm

import (
	"github.com/BurntSushi/xgb/xproto"

	"wm/pkg/xutil"
)

func (wm *WM) internAtom(name string) xproto.Atom {
	reply, err := xproto.InternAtom(wm.xu, false, uint16(len(name)), name).Reply()
	if err != nil {
		return 0
	}

	wm.ewmhAtoms[name] = reply.Atom
	return reply.Atom
}

func (wm *WM) atom(name string) xproto.Atom {
	if a, ok := wm.ewmhAtoms[name]; ok {
		return a
	}
	return wm.internAtom(name)
}

func (wm *WM) initEWMH() error {
	supportedNames := []string{
		"_NET_SUPPORTED",
		"_NET_WM_NAME",
		"_NET_WM_PID",
		"_NET_WM_WINDOW_TYPE",
		"_NET_ACTIVE_WINDOW",
		"_NET_CLOSE_WINDOW",
		"_NET_WM_STATE",
		"_NET_WM_STATE_FULLSCREEN",
		"_NET_NUMBER_OF_DESKTOPS",
		"_NET_CURRENT_DESKTOP",
		"_NET_DESKTOP_VIEWPORT",
		"_NET_WORKAREA",
		"_NET_WM_DESKTOP",
	}

	supportedAtoms := make([]xproto.Atom, 0, len(supportedNames))
	for _, name := range supportedNames {
		a := wm.internAtom(name)
		supportedAtoms = append(supportedAtoms, a)
	}

	sup := wm.atom("_NET_SUPPORTED")
	data := make([]byte, len(supportedAtoms)*4)
	for i, a := range supportedAtoms {
		data[i*4] = byte(a)
		data[i*4+1] = byte(a >> 8)
		data[i*4+2] = byte(a >> 16)
		data[i*4+3] = byte(a >> 24)
	}
	xproto.ChangeProperty(wm.xu, xproto.PropModeReplace, wm.root,
		sup, xproto.AtomAtom, 32, uint32(len(supportedAtoms)), data)

	nd := wm.atom("_NET_NUMBER_OF_DESKTOPS")
	n := uint32(len(wm.workspaces))
	xproto.ChangeProperty(wm.xu, xproto.PropModeReplace, wm.root,
		nd, xproto.AtomCardinal, 32, 1, []byte{
			byte(n), byte(n >> 8), byte(n >> 16), byte(n >> 24),
		})

	cd := wm.atom("_NET_CURRENT_DESKTOP")
	xproto.ChangeProperty(wm.xu, xproto.PropModeReplace, wm.root,
		cd, xproto.AtomCardinal, 32, 1, []byte{0, 0, 0, 0})

	wa := wm.atom("_NET_WORKAREA")
	sh := wm.screen.HeightInPixels - uint16(wm.conf.BarHeight)
	workarea := []uint32{0, 0, uint32(wm.screen.WidthInPixels), uint32(sh)}
	wab := make([]byte, 16)
	for i, v := range workarea {
		wab[i*4] = byte(v)
		wab[i*4+1] = byte(v >> 8)
		wab[i*4+2] = byte(v >> 16)
		wab[i*4+3] = byte(v >> 24)
	}
	xproto.ChangeProperty(wm.xu, xproto.PropModeReplace, wm.root,
		wa, xproto.AtomCardinal, 32, 4, wab)

	return nil
}

func (wm *WM) ewmhSendClientMessage(win xproto.Window, msgType string, data [5]uint32) {
	ev := xproto.ClientMessageEvent{
		Format: 32,
		Window: win,
		Type:   wm.atom(msgType),
		Data:   xproto.ClientMessageDataUnionData32New(data[:]),
	}

	xproto.SendEvent(wm.xu, false, wm.root,
		uint32(xproto.EventMaskSubstructureRedirect),
		string(ev.Bytes()))
}

func (wm *WM) ewmhCloseWindow(client xproto.Window) {
	ev := xproto.ClientMessageEvent{
		Format: 32,
		Window: client,
		Type:   wm.atom("_NET_CLOSE_WINDOW"),
		Data:   xproto.ClientMessageDataUnionData32New([]uint32{0, 0, 0, 0, 0}),
	}
	xproto.SendEvent(wm.xu, false, client,
		uint32(xproto.EventMaskNoEvent),
		string(ev.Bytes()))
}

func (wm *WM) ewmhActiveWindowSet(win xproto.Window) {
	ev := xproto.ClientMessageEvent{
		Format: 32,
		Window: win,
		Type:   wm.atom("_NET_ACTIVE_WINDOW"),
		Data: xproto.ClientMessageDataUnionData32New([]uint32{
			1, xproto.TimeCurrentTime, 0, 0, 0,
		}),
	}
	xproto.SendEvent(wm.xu, false, wm.root,
		uint32(xproto.EventMaskSubstructureRedirect|xproto.EventMaskSubstructureNotify),
		string(ev.Bytes()))
}

func (wm *WM) ewmhCurrentDesktopSet(ws uint) {
	v := uint32(ws)
	xproto.ChangeProperty(wm.xu, xproto.PropModeReplace, wm.root,
		wm.atom("_NET_CURRENT_DESKTOP"), xproto.AtomCardinal, 32, 1,
		[]byte{byte(v), byte(v >> 8), byte(v >> 16), byte(v >> 24)})
}

func (wm *WM) initRoot() error {
	if err := xutil.CheckOtherWM(wm.xu, wm.root); err != nil {
		return err
	}

	evMask := xproto.EventMaskSubstructureRedirect |
		xproto.EventMaskSubstructureNotify |
		xproto.EventMaskStructureNotify |
		xproto.EventMaskButtonPress |
		xproto.EventMaskButtonRelease |
		xproto.EventMaskPropertyChange

	attrMask := uint32(xproto.CwEventMask)
	attrVals := []uint32{uint32(evMask)}

	if wm.barBg != 0 {
		attrMask |= uint32(xproto.CwBackPixel)
		attrVals = append(attrVals, wm.barBg)
	}

	xproto.ChangeWindowAttributes(wm.xu, wm.root, attrMask, attrVals)

	if cursor, err := xutil.CreateRootCursor(wm.xu); err == nil {
		xproto.ChangeWindowAttributes(wm.xu, wm.root,
			uint32(xproto.CwCursor), []uint32{uint32(cursor)})
	}

	xproto.SetInputFocus(wm.xu,
		xproto.InputFocusPointerRoot,
		xproto.InputFocusNone,
		xproto.TimeCurrentTime)

	return nil
}
