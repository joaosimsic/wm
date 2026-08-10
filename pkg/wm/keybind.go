package wm

import (
	"github.com/BurntSushi/xgb/xproto"

	"wm/pkg/types"
	"wm/pkg/xutil"
)

var wsActions = []types.Action{
	types.ActionWS1, types.ActionWS2, types.ActionWS3, types.ActionWS4, types.ActionWS5,
	types.ActionWS6, types.ActionWS7, types.ActionWS8, types.ActionWS9, types.ActionWS10,
}

type bindDef struct {
	keysym xproto.Keysym
	mods   uint16
	action types.Action
}

func (wm *WM) setupKeybinds() error {
	var binds []bindDef

	dirKeysyms := []xproto.Keysym{0x068, 0x06a, 0x06b, 0x06c} // h, j, k, l
	modActions := map[uint16]types.Action{
		0:                 types.ActionFocusLeft,
		types.ShiftMask:   types.ActionMoveLeft,
		types.ControlMask: types.ActionResizeLeft,
	}
	for i, ks := range dirKeysyms {
		for mod, base := range modActions {
			binds = append(binds, bindDef{ks, types.Mod4Mask | mod, types.Action(int(base) + i)})
		}
	}

	binds = append(binds,
		bindDef{0x03b, types.Mod4Mask, types.ActionCmdMode},
		bindDef{0x063, types.Mod4Mask, types.ActionClose},
		bindDef{0xFF0D, types.Mod4Mask, types.ActionTerminal},
		bindDef{0x031, types.Mod4Mask, types.ActionWS1},
		bindDef{0x032, types.Mod4Mask, types.ActionWS2},
		bindDef{0x033, types.Mod4Mask, types.ActionWS3},
		bindDef{0x034, types.Mod4Mask, types.ActionWS4},
		bindDef{0x035, types.Mod4Mask, types.ActionWS5},
		bindDef{0x036, types.Mod4Mask, types.ActionWS6},
		bindDef{0x037, types.Mod4Mask, types.ActionWS7},
		bindDef{0x038, types.Mod4Mask, types.ActionWS8},
		bindDef{0x039, types.Mod4Mask, types.ActionWS9},
		bindDef{0x030, types.Mod4Mask, types.ActionWS10},
		bindDef{0xFF09, types.Mod4Mask, types.ActionWSNext},
		bindDef{0xFF09, types.Mod4Mask | types.ShiftMask, types.ActionWSPrev},
	)

	for _, b := range binds {
		xutil.GrabCombo(wm.xu, wm.root, b.keysym, b.mods)
		kc, ok := xutil.FindKeycode(wm.xu, b.keysym)
		if !ok {
			continue
		}
		wm.keyBinds = append(wm.keyBinds, types.KeyBind{
			Mods:    b.mods,
			Keycode: kc,
			Action:  b.action,
		})
	}

	return nil
}

func (wm *WM) matchKeyBind(ev xproto.KeyPressEvent) (types.Action, bool) {
	state := ev.State & ^uint16(types.IgnoredMods)
	for _, kb := range wm.keyBinds {
		targetMods := kb.Mods & ^uint16(types.IgnoredMods)
		if targetMods == state && kb.Keycode == ev.Detail {
			return kb.Action, true
		}
	}
	return 0, false
}
