package x11

import (
	"errors"
	"fmt"

	"github.com/jezek/xgb/xproto"
)

func (c *Connection) GrabKey(win xproto.Window, keycode xproto.Keycode, mods uint16) error {
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

func (c *Connection) GrabAllCombos(
	win xproto.Window,
	keycode xproto.Keycode,
	mods uint16,
	numlock uint16,
) error {
	combos := []uint16{
		mods,
		mods | xproto.KeyButMaskLock,
		mods | numlock,
		mods | xproto.KeyButMaskLock | numlock,
	}

	grabbed := make(map[uint16]bool)

	var failure []error

	for _, m := range combos {
		if grabbed[m] {
			continue
		}

		grabbed[m] = true

		if err := c.GrabKey(win, keycode, m); err != nil {
			failure = append(failure, fmt.Errorf("keycode %d mods 0x%x: %w", keycode, m, err))
		}
	}

	return errors.Join(failure...)
}

func (c *Connection) UngrabAllCombos(
	win xproto.Window,
	keycode xproto.Keycode,
	mods uint16,
	numlock uint16,
) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	combos := []uint16{
		mods,
		mods | xproto.KeyButMaskLock,
		mods | numlock,
		mods | xproto.KeyButMaskLock | numlock,
	}

	var failures []error

	for _, m := range combos {
		if err := c.UngrabKey(
			win,
			keycode,
			m,
		); err != nil {
			failures = append(
				failures,
				fmt.Errorf(
					"keycode %d mods 0x%x: %w",
					keycode,
					m,
					err,
				),
			)
		}
	}

	return errors.Join(failures...)
}
