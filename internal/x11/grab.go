package x11

import (
	"errors"
	"fmt"

	"github.com/jezek/xgb/xproto"
)

func (c *Connection) GrabKey(win xproto.Window, keycode xproto.Keycode, mods uint16) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	return c.grabKey(win, keycode, mods)
}

func (c *Connection) UngrabKey(
	win xproto.Window,
	keycode xproto.Keycode,
	mods uint16,
) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	return c.ungrabKey(win, keycode, mods)
}

func (c *Connection) GrabAllCombos(
	win xproto.Window,
	keycode xproto.Keycode,
	mods uint16,
	numlock uint16,
) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	var failures []error

	for _, m := range combos(mods, numlock) {
		if err := c.grabKey(win, keycode, m); err != nil {
			failures = append(failures, fmt.Errorf("keycode %d mods 0x%x: %w", keycode, m, err))
		}
	}

	return errors.Join(failures...)
}

func (c *Connection) UngrabAllCombos(
	win xproto.Window,
	keycode xproto.Keycode,
	mods uint16,
	numlock uint16,
) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	var failures []error

	for _, m := range combos(mods, numlock) {
		if err := c.ungrabKey(win, keycode, m); err != nil {
			failures = append(failures, fmt.Errorf("keycode %d mods 0x%x: %w", keycode, m, err))
		}
	}

	return errors.Join(failures...)
}

func combos(mods, numlock uint16) []uint16 {
	all := []uint16{
		mods,
		mods | xproto.KeyButMaskLock,
		mods | numlock,
		mods | xproto.KeyButMaskLock | numlock,
	}

	seen := make(map[uint16]struct{}, len(all))
	result := make([]uint16, 0, len(all))

	for _, m := range all {
		if _, ok := seen[m]; ok {
			continue
		}

		seen[m] = struct{}{}
		result = append(result, m)
	}

	return result
}

func (c *Connection) grabKey(
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

func (c *Connection) ungrabKey(
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
