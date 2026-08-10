// Package config provides configuration loading and color parsing for the window manager.
package config

import (
	"fmt"
	"strings"
)

type Config struct {
	Terminal    string `toml:"terminal"`
	Font        string `toml:"font"`
	BorderWidth int    `toml:"border_width"`
	Gap         int    `toml:"gap"`
	BarHeight   int    `toml:"bar_height"`
	Colors      Colors `toml:"colors"`
}

type Colors struct {
	BarBg          string `toml:"bar_bg"`
	BarFg          string `toml:"bar_fg"`
	BarActiveBg    string `toml:"bar_active_bg"`
	BorderActive   string `toml:"border_active"`
	BorderInactive string `toml:"border_inactive"`
}

func Default() Config {
	return Config{
		Terminal:    "xterm",
		Font:        "fixed",
		BorderWidth: 2,
		Gap:         0,
		BarHeight:   22,
	}
}

func HexToRGB(hex string) (uint32, error) {
	h := strings.TrimPrefix(hex, "#")
	if len(h) != 6 {
		return 0, fmt.Errorf("invalid color: %s", hex)
	}
	r := hval(h[0])*16 + hval(h[1])
	g := hval(h[2])*16 + hval(h[3])
	b := hval(h[4])*16 + hval(h[5])
	return uint32(r)*65536 + uint32(g)*256 + uint32(b), nil
}

func hval(c byte) uint16 {
	switch {
	case c >= '0' && c <= '9':
		return uint16(c - '0')
	case c >= 'a' && c <= 'f':
		return uint16(c-'a') + 10
	case c >= 'A' && c <= 'F':
		return uint16(c-'A') + 10
	}
	return 0
}
