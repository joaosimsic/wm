package theme

import (
	"fmt"

	"github.com/joaosimsic/wm/internal/config"
	"github.com/joaosimsic/wm/internal/x11"
)

type Palette struct {
	BarBg          uint32
	BarFg          uint32
	BarActiveBg    uint32
	BorderActive   uint32
	BorderInactive uint32
}

func NewPalette(conn *x11.Connection, colors config.Colors) (*Palette, error) {
	p := &Palette{}

	alloc := func(name string, c config.Color, dst *uint32) error {
		pixel, err := conn.AllocColorRGB(x11.RGB(c))
		if err != nil {
			return fmt.Errorf("%s: %w", name, err)
		}

		*dst = pixel

		return nil
	}

	for _, f := range []struct {
		name string
		c    config.Color
		dst  *uint32
	}{
		{"bar_bg", colors.BarBg, &p.BarBg},
		{"bar_fg", colors.BarFg, &p.BarFg},
		{"bar_active_bg", colors.BarActiveBg, &p.BarActiveBg},
		{"border_active", colors.BorderActive, &p.BorderActive},
		{"border_inactive", colors.BorderInactive, &p.BorderInactive},
	} {
		if err := alloc(f.name, f.c, f.dst); err != nil {
			return nil, err
		}
	}

	return p, nil
}
