// Package xutil provides low-level X protocol utility functions.
package xutil

import (
	"fmt"

	"github.com/BurntSushi/xgb"
	"github.com/BurntSushi/xgb/xproto"

	"wm/pkg/types"
)

func FindKeycode(conn *xgb.Conn, keysym xproto.Keysym) (xproto.Keycode, bool) {
	si := xproto.Setup(conn)
	minKc := si.MinKeycode
	maxKc := si.MaxKeycode

	reply, err := xproto.GetKeyboardMapping(conn, minKc, byte(maxKc-minKc+1)).Reply()
	if err != nil {
		return 0, false
	}

	kps := len(reply.Keysyms) / int(maxKc-minKc+1)
	for i := minKc; i <= maxKc; i++ {
		idx := int(i-minKc) * kps
		for j := 0; j < kps && idx+j < len(reply.Keysyms); j++ {
			if reply.Keysyms[idx+j] == keysym {
				return i, true
			}
		}
	}
	return 0, false
}

func GrabCombo(conn *xgb.Conn, root xproto.Window, keysym xproto.Keysym, mods uint16) {
	kc, ok := FindKeycode(conn, keysym)
	if !ok {
		return
	}
	combos := []uint16{0, types.LockMask, types.Mod2Mask, types.Mod5Mask,
		types.LockMask | types.Mod2Mask, types.LockMask | types.Mod5Mask,
		types.Mod2Mask | types.Mod5Mask, types.LockMask | types.Mod2Mask | types.Mod5Mask}
	for _, m := range combos {
		xproto.GrabKey(conn, true, root, mods|m, kc,
			xproto.GrabModeAsync, xproto.GrabModeAsync)
	}
}

func CheckOtherWM(conn *xgb.Conn, root xproto.Window) error {
	err := xproto.ChangeWindowAttributesChecked(conn, root,
		xproto.CwEventMask,
		[]uint32{uint32(xproto.EventMaskSubstructureRedirect)}).Check()
	if err != nil {
		return fmt.Errorf("another window manager is already running: %w", err)
	}
	return nil
}

func CreateRootCursor(conn *xgb.Conn) (xproto.Cursor, error) {
	cursor, err := xproto.NewCursorId(conn)
	if err != nil {
		return 0, err
	}

	cursorFont, err := xproto.NewFontId(conn)
	if err != nil {
		return cursor, err
	}

	name := "cursor"
	if cerr := xproto.OpenFontChecked(conn, cursorFont, uint16(len(name)), name).Check(); cerr != nil {
		return cursor, cerr
	}

	xproto.CreateGlyphCursor(conn, cursor,
		cursorFont, cursorFont,
		68, 69,
		65535, 65535, 65535,
		0, 0, 0)

	return cursor, nil
}
