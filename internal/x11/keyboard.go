package x11

import (
	"errors"
	"fmt"

	"github.com/jezek/xgb/xproto"
)

const modifierSlots = 8

type Keyboard struct {
	Keymap  *KeyMapping
	NumLock uint16
}

func (c *Connection) KeyMapping() (*KeyMapping, error) {
	setup := xproto.Setup(c.conn)

	codeRange := byte(setup.MaxKeycode - setup.MinKeycode + 1)

	reply, err := xproto.GetKeyboardMapping(
		c.conn,
		setup.MinKeycode,
		codeRange,
	).Reply()
	if err != nil {
		return nil, err
	}

	km := newKeyMapping(
		setup.MinKeycode,
		setup.MaxKeycode,
		reply.KeysymsPerKeycode,
		reply.Keysyms,
	)

	return km, nil
}

func (c *Connection) Keyboard() (*Keyboard, error) {
	km, err := c.KeyMapping()
	if err != nil {
		return nil, err
	}

	numLock, err := c.numLockMask(km)
	if err != nil {
		return nil, err
	}

	return &Keyboard{
		Keymap:  km,
		NumLock: numLock,
	}, nil
}

func (c *Connection) numLockMask(km *KeyMapping) (uint16, error) {
	reply, err := xproto.GetModifierMapping(c.conn).Reply()
	if err != nil {
		return 0, fmt.Errorf("get modifier mapping: %w", err)
	}

	mask, ok := numLockMaskFromMapping(km, reply)
	if !ok {
		return 0, errors.New("NumLock keycode not found in modifier mapping")
	}

	return mask, nil
}

func numLockMaskFromMapping(
	km *KeyMapping,
	reply *xproto.GetModifierMappingReply,
) (uint16, bool) {
	numLockKeycodes, ok := km.KeycodesForKeysym(KeysymNumLock)
	if !ok {
		return 0, false
	}

	keycodePerMod := int(reply.KeycodesPerModifier)
	if keycodePerMod <= 0 {
		return 0, false
	}

	numLockKeycodesMap := make(map[xproto.Keycode]struct{}, len(numLockKeycodes))
	for _, keycode := range numLockKeycodes {
		numLockKeycodesMap[keycode] = struct{}{}
	}

	for mod := range modifierSlots {
		start := mod * keycodePerMod
		end := min(start+keycodePerMod, len(reply.Keycodes))

		for _, keycode := range reply.Keycodes[start:end] {
			if _, ok := numLockKeycodesMap[keycode]; ok {
				return uint16(1 << mod), true
			}
		}
	}

	return 0, false
}
