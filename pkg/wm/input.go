package wm

import (
	"fmt"
	"os"

	"github.com/BurntSushi/xgb/xproto"

	"wm/pkg/layout"
	"wm/pkg/types"
)

var keysymNames = map[xproto.Keysym]string{
	0xff1b: "Escape",
	0xff0d: "Return",
	0xff08: "BackSpace",
	0xff09: "Tab",
	0xff50: "Home",
	0xff57: "End",
	0xff51: "Left",
	0xff53: "Right",
	0xff52: "Up",
	0xff54: "Down",
	0xffff: "Delete",
}

var actionDirs = map[types.Action]string{
	types.ActionFocusLeft: "left", types.ActionFocusDown: "down", types.ActionFocusUp: "up", types.ActionFocusRight: "right",
	types.ActionMoveLeft: "left", types.ActionMoveDown: "down", types.ActionMoveUp: "up", types.ActionMoveRight: "right",
}

var resizeDeltas = map[types.Action]float64{
	types.ActionResizeLeft: -0.02, types.ActionResizeDown: 0.02, types.ActionResizeUp: -0.02, types.ActionResizeRight: 0.02,
}

func (wm *WM) keycodeToSym(code xproto.Keycode, state uint16) string {
	si := xproto.Setup(wm.xu)
	minKc := int(si.MinKeycode)
	maxKc := int(si.MaxKeycode)

	if int(code) < minKc || int(code) > maxKc {
		return ""
	}

	reply, err := xproto.GetKeyboardMapping(wm.xu, code, 1).Reply()
	if err != nil || len(reply.Keysyms) == 0 {
		return ""
	}

	kps := len(reply.Keysyms)
	idx := 0
	if state&types.ShiftMask != 0 && kps >= 2 {
		idx = 1
	}
	if idx >= len(reply.Keysyms) {
		idx = 0
	}

	r := reply.Keysyms[idx]
	if r == 0 {
		return ""
	}

	if name, ok := keysymNames[r]; ok {
		return name
	}
	if r >= 32 && r <= 126 {
		return string(rune(r))
	}
	return fmt.Sprintf("0x%x", r)
}

func (wm *WM) handleKeyPress(ev xproto.KeyPressEvent) {
	if wm.mode == types.ModeCommand {
		wm.handleCommandKey(ev)
		return
	}

	action, ok := wm.matchKeyBind(ev)
	if !ok {
		return
	}

	if dir, ok := actionDirs[action]; ok {
		if isFocusAction(action) {
			wm.focusDirection(dir)
		} else {
			wm.moveWindow(dir)
		}
		return
	}

	if delta, ok := resizeDeltas[action]; ok {
		wm.resizeWindow(delta)
		return
	}

	switch action {
	case types.ActionCmdMode:
		wm.enterCommandMode()
	case types.ActionClose:
		wm.closeFocused()
	case types.ActionTerminal:
		if err := wm.launch(wm.conf.Terminal); err != nil {
			fmt.Fprintf(os.Stderr, "launch terminal: %v\n", err)
		}
	case types.ActionWS1, types.ActionWS2, types.ActionWS3, types.ActionWS4, types.ActionWS5,
		types.ActionWS6, types.ActionWS7, types.ActionWS8, types.ActionWS9, types.ActionWS10:
		for i, a := range wsActions {
			if a == action {
				wm.switchWorkspace(i)
			}
		}
	case types.ActionWSNext:
		wm.switchWorkspace((wm.currentWS + 1) % len(wm.workspaces))
	case types.ActionWSPrev:
		prev := wm.currentWS - 1
		if prev < 0 {
			prev = len(wm.workspaces) - 1
		}
		wm.switchWorkspace(prev)
	default:
		fmt.Fprintf(os.Stderr, "wm: unknown action %d\n", action)
	}
}

func isFocusAction(a types.Action) bool {
	return a >= types.ActionFocusLeft && a <= types.ActionFocusRight
}

func (wm *WM) handleCommandKey(ev xproto.KeyPressEvent) {
	sym := wm.keycodeToSym(ev.Detail, ev.State)

	switch sym {
	case "Escape":
		wm.mode = types.ModeNormal
		wm.cmdBuffer = ""
		wm.ungrabKeyboard()
	case "Return":
		wm.executeCommandString(wm.cmdBuffer)
		wm.mode = types.ModeNormal
		wm.cmdBuffer = ""
		wm.ungrabKeyboard()
	case "BackSpace":
		if len(wm.cmdBuffer) > 0 {
			wm.cmdBuffer = wm.cmdBuffer[:len(wm.cmdBuffer)-1]
		}
	default:
		if len(sym) == 1 && sym[0] >= 32 && sym[0] <= 126 {
			wm.cmdBuffer += sym
		}
	}

	wm.bar.Update(wm.mode, wm.currentWS, wm.cmdBuffer, wm.focusedTitle())
}

func (wm *WM) focusDirection(dir string) {
	ws := wm.workspaces[wm.currentWS]
	if ws.Root == nil {
		return
	}

	dirMap := map[string]layout.FocusDir{
		"left": layout.FocusDirLeft, "down": layout.FocusDirDown,
		"up": layout.FocusDirUp, "right": layout.FocusDirRight,
	}

	target := ws.Root.FocusInDir(dirMap[dir])
	if target != nil && target != wm.focused {
		wm.setFocus(target)
		wm.updateBorders()
		wm.bar.Update(wm.mode, wm.currentWS, wm.cmdBuffer, wm.focusedTitle())
	}
}

func (wm *WM) moveWindow(dir string) {
	_ = dir
}

func (wm *WM) resizeWindow(delta float64) {
	ws := wm.workspaces[wm.currentWS]
	if ws.Root == nil {
		return
	}
	if wm.focused != nil {
		ws.Root.SetFocusTo(wm.focused)
	}
	ws.Root.ResizeFocused(delta)
	wm.tileWorkspace()
}

func (wm *WM) enterCommandMode() {
	wm.mode = types.ModeCommand
	wm.cmdBuffer = ""
	wm.grabKeyboard()
	wm.bar.Update(wm.mode, wm.currentWS, wm.cmdBuffer, wm.focusedTitle())
}

func (wm *WM) grabKeyboard() {
	reply, err := xproto.GrabKeyboard(wm.xu, false, wm.root,
		xproto.TimeCurrentTime, xproto.GrabModeAsync, xproto.GrabModeAsync).Reply()
	if err != nil {
		fmt.Fprintf(os.Stderr, "grab keyboard: %v\n", err)
		return
	}
	if reply.Status != xproto.GrabStatusSuccess {
		fmt.Fprintf(os.Stderr, "grab keyboard failed: status %d\n", reply.Status)
		wm.mode = types.ModeNormal
	}
}

func (wm *WM) ungrabKeyboard() {
	xproto.UngrabKeyboard(wm.xu, xproto.TimeCurrentTime)
}

func (wm *WM) executeCommandString(input string) {
	cmd, err := types.ParseCommand(":" + input)
	if err != nil {
		return
	}
	if err := wm.executeCommand(cmd); err != nil {
		fmt.Fprintf(os.Stderr, "command error: %v\n", err)
	}
	wm.tileWorkspace()
}
