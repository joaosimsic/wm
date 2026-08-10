package wm

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/BurntSushi/xgb/xproto"

	"wm/pkg/layout"
)

func (wm *WM) setFocus(mw *layout.ManagedWindow) {
	if wm.focused != nil {
		wm.focused.Focused = false
	}
	wm.focused = mw
	if mw != nil {
		mw.Focused = true
		xproto.SetInputFocus(wm.xu,
			xproto.InputFocusPointerRoot,
			mw.Client,
			xproto.TimeCurrentTime)
		wm.ewmhActiveWindowSet(mw.Client)
	}
}

func (wm *WM) setFocusByWindow(client xproto.Window) {
	mw, ok := wm.windows[client]
	if !ok {
		return
	}
	wm.setFocus(mw)
	ws := wm.workspaces[wm.currentWS]
	if ws.Root != nil {
		ws.Root.SetFocusTo(mw)
	}
	wm.updateBorders()
	wm.bar.Update(wm.mode, wm.currentWS, wm.cmdBuffer, wm.focusedTitle())
}

func (wm *WM) updateBorders() {
	for _, mw := range wm.windows {
		color := wm.borderInactive
		if mw.Focused {
			color = wm.borderActive
		}
		xproto.ChangeWindowAttributes(wm.xu, mw.Frame,
			xproto.CwBorderPixel, []uint32{color})
	}
}

func (wm *WM) applyLayout(c *layout.Container) {
	if c == nil {
		return
	}
	if c.Type == layout.ContainerLeaf && c.Window != nil {
		w := c.Window
		bw := wm.conf.BorderWidth
		r := w.Geom
		bw32 := uint32(bw)

		xproto.ConfigureWindow(wm.xu, w.Frame,
			xproto.ConfigWindowX|xproto.ConfigWindowY|
				xproto.ConfigWindowWidth|xproto.ConfigWindowHeight|
				xproto.ConfigWindowBorderWidth,
			[]uint32{
				uint32(r.X - bw), uint32(r.Y - bw),
				uint32(r.W + 2*bw), uint32(r.H + 2*bw),
				bw32,
			})

		xproto.ConfigureWindow(wm.xu, w.Client,
			xproto.ConfigWindowX|xproto.ConfigWindowY|
				xproto.ConfigWindowWidth|xproto.ConfigWindowHeight,
			[]uint32{0, 0, uint32(r.W), uint32(r.H)})
		return
	}
	for _, ch := range c.Children {
		wm.applyLayout(ch)
	}
}

func (wm *WM) launch(cmd string) error {
	parts := strings.Fields(cmd)
	if len(parts) == 0 {
		return fmt.Errorf("empty command")
	}

	c := exec.Command(parts[0], parts[1:]...)
	c.Env = append(os.Environ(), "DISPLAY="+os.Getenv("DISPLAY"))
	c.Stdin = nil
	c.Stdout = nil
	c.Stderr = nil

	if err := c.Start(); err != nil {
		return fmt.Errorf("start %s: %w", cmd, err)
	}

	go c.Wait()
	return nil
}

func (wm *WM) killWindow(client xproto.Window) error {
	return xproto.KillClientChecked(wm.xu, uint32(client)).Check()
}

func (wm *WM) closeFocused() {
	if wm.focused == nil {
		return
	}
	wm.ewmhCloseWindow(wm.focused.Client)
}

func (wm *WM) focusedTitle() string {
	if wm.focused == nil {
		return ""
	}
	return wm.focused.Title
}
