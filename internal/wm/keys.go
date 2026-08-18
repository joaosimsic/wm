package wm

import (
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/jezek/xgb/xproto"
	"github.com/joaosimsic/wm/internal/x11"
)

var namedKeysyms = map[string]xproto.Keysym{
	"Return":    0xff0d,
	"Tab":       0xff09,
	"space":     0x20,
	"Escape":    0xff1b,
	"BackSpace": 0xff08,
	"Delete":    0xffff,
	"Home":      0xff50,
	"End":       0xff57,
	"Left":      0xff51,
	"Right":     0xff53,
	"Up":        0xff52,
	"Down":      0xff54,
	"Page_Up":   0xff55,
	"Page_Down": 0xff56,
}

type combo struct {
	mods uint16
	sym  xproto.Keysym
}

func keysymFromName(name string) (xproto.Keysym, bool) {
	r, size := utf8.DecodeRuneInString(name)
	if size == len(name) && r != utf8.RuneError {
		return xproto.Keysym(r), true
	}

	s, ok := namedKeysyms[name]

	return s, ok
}

func parseCombo(s string) (combo, error) {
	var c combo

	parts := strings.Split(s, "+")

	mods := parts[:len(parts)-1]
	parsedCombo := parts[len(parts)-1]

	for _, p := range mods {
		m, err := x11.ParseModifiers(p)
		if err != nil {
			return c, fmt.Errorf("combo %q: %w", s, err)
		}

		c.mods |= m
	}

	sym, ok := keysymFromName(parsedCombo)
	if !ok {
		return c, fmt.Errorf("combo %q: unknown keysym %q", s, parsedCombo)
	}

	c.sym = sym

	return c, nil
}
