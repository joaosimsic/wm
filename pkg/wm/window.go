package wm

import (
	"github.com/BurntSushi/xgb/xproto"

	"wm/pkg/layout"
)

func (wm *WM) manageWindow(client xproto.Window) {
	geom, err := xproto.GetGeometry(wm.xu, xproto.Drawable(client)).Reply()
	if err != nil {
		return
	}

	attrs, err := xproto.GetWindowAttributes(wm.xu, client).Reply()
	if err != nil {
		return
	}
	if attrs.OverrideRedirect {
		xproto.MapWindow(wm.xu, client)
		return
	}

	bw := wm.conf.BorderWidth
	frame := wm.createFrame(geom, bw)

	mw := &layout.ManagedWindow{
		Frame:  frame,
		Client: client,
		Title:  wm.getWindowTitle(client),
	}

	wm.windows[client] = mw
	wm.frames[frame] = mw

	xproto.ReparentWindow(wm.xu, client, frame, 0, 0)
	xproto.ChangeSaveSet(wm.xu, xproto.SetModeInsert, client)

	xproto.MapWindow(wm.xu, frame)
	xproto.MapWindow(wm.xu, client)

	ws := wm.workspaces[wm.currentWS]

	if wm.pendingSplit && ws.Root != nil {
		wm.insertWindowAsSplit(mw)
	} else {
		ws.insertWindow(mw)
	}

	wm.setFocus(mw)
	wm.updateBorders()
	wm.tileWorkspace()
	wm.bar.Update(wm.mode, wm.currentWS, wm.cmdBuffer, wm.focusedTitle())
}

func (wm *WM) createFrame(geom *xproto.GetGeometryReply, bw int) xproto.Window {
	frame, _ := xproto.NewWindowId(wm.xu)
	fw := uint16(geom.Width)
	fh := uint16(geom.Height)

	xproto.CreateWindow(wm.xu,
		wm.screen.RootDepth,
		frame,
		wm.root,
		int16(geom.X),
		int16(geom.Y),
		fw, fh,
		uint16(bw),
		xproto.WindowClassInputOutput,
		wm.screen.RootVisual,
		xproto.CwBackPixel|xproto.CwEventMask,
		[]uint32{
			wm.borderInactive,
			uint32(xproto.EventMaskSubstructureNotify | xproto.EventMaskSubstructureRedirect |
				xproto.EventMaskEnterWindow | xproto.EventMaskButtonPress),
		},
	)

	return frame
}

func (wm *WM) insertWindowAsSplit(mw *layout.ManagedWindow) {
	ws := wm.workspaces[wm.currentWS]
	wm.pendingSplit = false
	direction := wm.pendingDir
	newLeaf := &layout.Container{Type: layout.ContainerLeaf, Window: mw}

	parent := ws.Root.FindParentOf(wm.focused)
	if parent != nil && len(parent.Children) > 1 {
		parent.Children = append(parent.Children, newLeaf)
		newLeaf.Parent = parent
		parent.CurFocus = len(parent.Children) - 1
		return
	}

	focusedLeaf := ws.Root.FindFocused()
	if focusedLeaf != nil && focusedLeaf.Parent != nil {
		i := layout.FindChildIndex(focusedLeaf.Parent, focusedLeaf)
		newSplit := &layout.Container{Type: direction, Children: []*layout.Container{focusedLeaf.Parent.Children[i], newLeaf}}
		newSplit.FixChildren()
		focusedLeaf.Parent.Children[i] = newSplit
		newSplit.Parent = focusedLeaf.Parent
		newSplit.CurFocus = 1
		return
	}

	ws.Root = &layout.Container{Type: direction, Children: []*layout.Container{ws.Root, newLeaf}}
	ws.Root.FixChildren()
	ws.Root.CurFocus = 1
}

func (ws *Workspace) insertWindow(mw *layout.ManagedWindow) {
	newLeaf := &layout.Container{Type: layout.ContainerLeaf, Window: mw}

	switch {
	case ws.Root == nil:
		ws.Root = newLeaf
	case ws.Root.Type == layout.ContainerLeaf:
		oldLeaf := ws.Root
		ws.Root = &layout.Container{
			Type:     layout.ContainerVSplit,
			Children: []*layout.Container{oldLeaf, newLeaf},
		}
		ws.Root.FixChildren()
		ws.Root.CurFocus = 1
	default:
		if len(ws.Root.Children) > 0 && ws.Root.Children[ws.Root.CurFocus].Type == layout.ContainerLeaf {
			focused := ws.Root.FindFocused()
			if focused != nil && focused.Parent != nil {
				parent := focused.Parent
				parent.Children = append(parent.Children, newLeaf)
				newLeaf.Parent = parent
				parent.CurFocus = len(parent.Children) - 1
				return
			}
		}
		ws.Root.Children = append(ws.Root.Children, newLeaf)
		newLeaf.Parent = ws.Root
		ws.Root.CurFocus = len(ws.Root.Children) - 1
	}
}

func (wm *WM) removeWindow(client xproto.Window) {
	mw, ok := wm.windows[client]
	if !ok {
		return
	}

	for _, ws := range wm.workspaces {
		if ws.Root == nil {
			continue
		}
		if ws.Root.Type == layout.ContainerLeaf && ws.Root.Window == mw {
			ws.Root = nil
			continue
		}
		if ws.Root.RemoveWindow(mw) {
			switch {
			case len(ws.Root.Children) == 0:
				ws.Root = nil
			case len(ws.Root.Children) == 1:
				child := ws.Root.Children[0]
				child.Parent = nil
				ws.Root = child
			}
		}
	}

	if wm.focused == mw {
		wm.focused = nil
		ws := wm.workspaces[wm.currentWS]
		if ws.Root != nil {
			if newFocus := layout.FindNextLeaf(ws.Root); newFocus != nil {
				wm.setFocus(newFocus)
				ws.Root.SetFocusTo(newFocus)
			}
		}
	}

	xproto.DestroyWindow(wm.xu, mw.Frame)
	delete(wm.windows, client)
	delete(wm.frames, mw.Frame)

	wm.tileWorkspace()
	wm.updateBorders()
	wm.bar.Update(wm.mode, wm.currentWS, wm.cmdBuffer, wm.focusedTitle())
}

func (wm *WM) getWindowTitle(w xproto.Window) string {
	for _, prop := range []string{"_NET_WM_NAME", "WM_NAME"} {
		atom := wm.internAtom(prop)
		if atom == 0 {
			continue
		}
		reply, err := xproto.GetProperty(wm.xu, false, w, atom,
			xproto.GetPropertyTypeAny, 0, 256).Reply()
		if err != nil || reply.ValueLen == 0 {
			continue
		}
		return string(reply.Value)
	}
	return ""
}
