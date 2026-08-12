package config

import "fmt"

type RGB struct {
	R uint8
	G uint8
	B uint8
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

func (c Colors) RGB() (map[string]RGB, error) {
	colors := map[string]string{
		"bar_bg":          c.BarBg,
		"bar_fg":          c.BarFg,
		"bar_active_bg":   c.BarActiveBg,
		"border_active":   c.BorderActive,
		"border_inactive": c.BorderInactive,
	}

	out := make(map[string]RGB, len(colors))

	for name, hex := range colors {
		rgb, err := hexToRGB(hex)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", name, err)
		}

		out[name] = rgb
	}

	return out, nil
}
