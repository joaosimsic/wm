package config

import "fmt"

type RGB struct {
	R uint8
	G uint8
	B uint8
}

type RGBColors struct {
	BarBg          RGB
	BarFg          RGB
	BarActiveBg    RGB
	BorderActive   RGB
	BorderInactive RGB
}

func (c Colors) RGB() (RGBColors, error) {
	var out RGBColors
	var err error

	if out.BarBg, err = hexToRGB(c.BarBg); err != nil {
		return RGBColors{}, fmt.Errorf("bar_bg: %w", err)
	}

	if out.BarFg, err = hexToRGB(c.BarFg); err != nil {
		return RGBColors{}, fmt.Errorf("bar_fg: %w", err)
	}

	if out.BarActiveBg, err = hexToRGB(c.BarActiveBg); err != nil {
		return RGBColors{}, fmt.Errorf("bar_active_bg: %w", err)
	}

	if out.BorderActive, err = hexToRGB(c.BorderActive); err != nil {
		return RGBColors{}, fmt.Errorf("border_active: %w", err)
	}

	if out.BorderInactive, err = hexToRGB(c.BorderInactive); err != nil {
		return RGBColors{}, fmt.Errorf("border_inactive: %w", err)
	}

	return out, nil
}

func hexToRGB(hex string) (RGB, error) {
	if len(hex) != 7 || hex[0] != '#' {
		return RGB{}, fmt.Errorf("invalid color: %q (want #RRGGBB)", hex)
	}

	r, err := hexByte(hex[1], hex[2])
	if err != nil {
		return RGB{}, fmt.Errorf("invalid color %q: %w", hex, err)
	}

	g, err := hexByte(hex[3], hex[4])
	if err != nil {
		return RGB{}, fmt.Errorf("invalid color %q: %w", hex, err)
	}

	b, err := hexByte(hex[5], hex[6])
	if err != nil {
		return RGB{}, fmt.Errorf("invalid color %q: %w", hex, err)
	}

	return RGB{
		R: r,
		G: g,
		B: b,
	}, nil
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
