package x11

import (
	"fmt"
	"strings"

	"github.com/jezek/xgb/xproto"
)

const (
	KeysymsPerGroup = 2
	LevelUnshifted  = 0
	LevelShift      = 1
)

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
	MinCode           xproto.Keycode
	MaxCode           xproto.Keycode
	KeysymsPerKeycode byte
	Keysyms           []xproto.Keysym
}

func (c *Connection) KeyMapping() (*KeyMapping, error) {
	setup := xproto.Setup(c.conn)

	count := byte(setup.MaxKeycode - setup.MinKeycode + 1)

	reply, err := xproto.GetKeyboardMapping(
		c.conn,
		setup.MinKeycode,
		count,
	).Reply()
	if err != nil {
		return nil, err
	}

	return &KeyMapping{
		MinCode:           setup.MinKeycode,
		MaxCode:           setup.MaxKeycode,
		KeysymsPerKeycode: reply.KeysymsPerKeycode,
		Keysyms:           reply.Keysyms,
	}, nil
}

func (m *KeyMapping) KeysymsForCode(
	code xproto.Keycode,
	group int,
) xproto.Keysym {
	return m.KeysymForCodeAndLevel(code, group, LevelUnshifted)
}

func (m *KeyMapping) KeysymForCodeAndLevel(
	code xproto.Keycode,
	group int,
	level int,
) xproto.Keysym {
	idx, ok := m.keysymIndex(code, group, level)
	if !ok {
		return 0
	}

	return m.Keysyms[idx]
}

func (c *Connection) GrabKey(
	win xproto.Window,
	keycode xproto.Keycode,
	mods uint16,
) error {
	return xproto.GrabKeyChecked(
		c.conn,
		true,
		win,
		mods,
		keycode,
		xproto.GrabModeAsync,
		xproto.GrabModeAsync,
	).Check()
}

func (c *Connection) UngrabKey(
	win xproto.Window,
	keycode xproto.Keycode,
	mods uint16,
) error {
	return xproto.UngrabKeyChecked(
		c.conn,
		keycode,
		win,
		mods,
	).Check()
}

func ParseModifiers(names ...string) (uint16, error) {
	var mask uint16

	for _, name := range names {
		key := strings.ToLower(strings.TrimSpace(name))
		if key == "" {
			continue
		}

		modifier, ok := modifierMap[key]
		if !ok {
			return 0, fmt.Errorf("unknown modifier name: %q", name)
		}

		mask |= modifier
	}

	return mask, nil
}

func ParseModifierString(str string) (uint16, error) {
	if strings.TrimSpace(str) == "" {
		return 0, nil
	}

	return ParseModifiers(strings.Split(str, "+")...)
}

func (m *KeyMapping) keysymIndex(
	code xproto.Keycode,
	group int,
	level int,
) (int, bool) {
	if code < m.MinCode || code > m.MaxCode {
		return 0, false
	}

	keysPerCode := int(m.KeysymsPerKeycode)

	if keysPerCode == 0 {
		return 0, false
	}

	if level < 0 || level >= KeysymsPerGroup {
		return 0, false
	}

	if level >= keysPerCode {
		return 0, false
	}

	if keysPerCode%KeysymsPerGroup != 0 {
		return 0, false
	}

	groupsPerKey := keysPerCode / KeysymsPerGroup
	if group < 0 || group >= groupsPerKey {
		return 0, false
	}

	keyOffset := int(code - m.MinCode)

	flatIndex := keyOffset*keysPerCode +
		group*KeysymsPerGroup +
		level

	if flatIndex < 0 || flatIndex >= len(m.Keysyms) {
		return 0, false
	}

	return flatIndex, true
}
