package config

import (
	"fmt"
)

func validate(cfg *Config) error {
	if cfg.BorderWidth < 0 {
		return fmt.Errorf("border_width must be non-negative: %d", cfg.BorderWidth)
	}

	if cfg.Gap < 0 {
		return fmt.Errorf("gap must be non-negative: %d", cfg.Gap)
	}

	if cfg.BarHeight < 0 {
		return fmt.Errorf("bar_height must be non-negative: %d", cfg.BarHeight)
	}

    _, err := cfg.RGBColors()
    return err
}
