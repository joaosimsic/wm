package x11

import "github.com/jezek/xgb/xproto"

const KeysymNumLock = 0xff7f

var modifierMap = map[string]uint16{
	"shift":   xproto.KeyButMaskShift,
	"lock":    xproto.KeyButMaskLock,
	"caps":    xproto.KeyButMaskLock,
	"control": xproto.KeyButMaskControl,
	"ctrl":    xproto.KeyButMaskControl,
	"alt":     xproto.KeyButMaskMod1,
	"mod1":    xproto.KeyButMaskMod1,
	"mod2":    xproto.KeyButMaskMod2,
	"mod3":    xproto.KeyButMaskMod3,
	"super":   xproto.KeyButMaskMod4,
	"win":     xproto.KeyButMaskMod4,
	"mod4":    xproto.KeyButMaskMod4,
	"mod5":    xproto.KeyButMaskMod5,
}

type KeyMapping struct {
	minCode           xproto.Keycode
	maxCode           xproto.Keycode
	keysymsPerKeycode byte
	keysyms           []xproto.Keysym
	keycodesByKeysym  map[xproto.Keysym][]xproto.Keycode
}

func newKeyMapping(
	minCode xproto.Keycode,
	maxCode xproto.Keycode,
	keysymsPerKeycode byte,
	keysyms []xproto.Keysym,
) *KeyMapping {
	return &KeyMapping{
		minCode:           minCode,
		maxCode:           maxCode,
		keysymsPerKeycode: keysymsPerKeycode,
		keysyms:           keysyms,
		keycodesByKeysym:  make(map[xproto.Keysym][]xproto.Keycode),
	}
}

func (m *KeyMapping) buildInverseMapping() {
	if m.keysymsPerKeycode == 0 {
		return
	}

	for code := m.minCode; code <= m.maxCode; code++ {
		syms, ok := m.KeysymsForKeycode(code)
		if !ok {
			continue
		}

		for _, sym := range syms {
			if sym == 0 {
				continue
			}

			m.keycodesByKeysym[sym] = append(m.keycodesByKeysym[sym], code)
		}
	}
}

func (m *KeyMapping) KeysymsForKeycode(code xproto.Keycode) ([]xproto.Keysym, bool) {
	if m == nil || code < m.minCode || code > m.maxCode {
		return nil, false
	}

	n := int(m.keysymsPerKeycode)
	if n == 0 {
		return nil, false
	}

	offset := int(code-m.minCode) * n
	if offset+n > len(m.keysyms) {
		return nil, false
	}

	return m.keysyms[offset : offset+n], true
}

func (m *KeyMapping) KeycodesForKeysym(keysym xproto.Keysym) ([]xproto.Keycode, bool) {
	if m == nil {
		return nil, false
	}

	codes, ok := m.keycodesByKeysym[keysym]

	return codes, ok
}
