package config

import "fmt"

type Color struct {
	R uint8
	G uint8
	B uint8
}

type Colors struct {
	BarBg          Color `toml:"bar_bg"`
	BarFg          Color `toml:"bar_fg"`
	BarActiveBg    Color `toml:"bar_active_bg"`
	BorderActive   Color `toml:"border_active"`
	BorderInactive Color `toml:"border_inactive"`
}

func DefaultColors() Colors {
	return Colors{
		BarBg:          MustParseHex("#1a1a1a"),
		BarFg:          MustParseHex("#cccccc"),
		BarActiveBg:    MustParseHex("#4a4a4a"),
		BorderActive:   MustParseHex("#5f87af"),
		BorderInactive: MustParseHex("#333333"),
	}
}

func ParseHex(s string) (Color, error) {
	if len(s) != 7 || s[0] != '#' {
		return Color{}, fmt.Errorf("invalid color: %q (want #RRGGBB)", s)
	}

	r, err := hexByte(s[1], s[2])
	if err != nil {
		return Color{}, fmt.Errorf("invalid color %q: %w", s, err)
	}

	g, err := hexByte(s[3], s[4])
	if err != nil {
		return Color{}, fmt.Errorf("invalid color %q: %w", s, err)
	}

	b, err := hexByte(s[5], s[6])
	if err != nil {
		return Color{}, fmt.Errorf("invalid color %q: %w", s, err)
	}

	return Color{R: r, G: g, B: b}, nil
}

func MustParseHex(s string) Color {
	c, err := ParseHex(s)
	if err != nil {
		panic(err)
	}

	return c
}

func (c *Color) UnmarshalText(text []byte) error {
	parsed, err := ParseHex(string(text))
	if err != nil {
		return err
	}

	*c = parsed

	return nil
}

func hexByte(hi, lo byte) (uint8, error) {
	h, ok := hexValue(hi)
	if !ok {
		return 0, fmt.Errorf("invalid hex digit: %q", hi)
	}

	l, ok := hexValue(lo)
	if !ok {
		return 0, fmt.Errorf("invalid hex digit: %q", lo)
	}

	return h<<4 | l, nil
}

func hexValue(c byte) (uint8, bool) {
	switch {
	case c >= '0' && c <= '9':
		return c - '0', true
	case c >= 'a' && c <= 'f':
		return c - 'a' + 10, true
	case c >= 'A' && c <= 'F':
		return c - 'A' + 10, true
	default:
		return 0, false
	}
}
