package config

import "fmt"

func validate(cfg *Config) []error {
	def := Default()

	var errs []error

	if cfg.BorderWidth < 0 {
		errs = append(errs, fmt.Errorf("border_width must be non-negative: %d, using default %d", cfg.BorderWidth, def.BorderWidth))
		cfg.BorderWidth = def.BorderWidth
	}

	if cfg.Gap < 0 {
		errs = append(errs, fmt.Errorf("gap must be non-negative: %d, using default %d", cfg.Gap, def.Gap))
		cfg.Gap = def.Gap
	}

	if cfg.BarHeight < 0 {
		errs = append(errs, fmt.Errorf("bar_height must be non-negative: %d, using default %d", cfg.BarHeight, def.BarHeight))
		cfg.BarHeight = def.BarHeight
	}

	if cfg.Terminal == "" {
		errs = append(errs, fmt.Errorf("terminal missing, using default %q", def.Terminal))
		cfg.Terminal = def.Terminal
	}

	if cfg.Font == "" {
		errs = append(errs, fmt.Errorf("font missing, using default %q", def.Font))
		cfg.Font = def.Font
	}

	if cfg.Mod == "" {
		errs = append(errs, fmt.Errorf("mod missing, using default %q", def.Mod))
		cfg.Mod = def.Mod
	}

	if cfg.SplitRatio <= 0 || cfg.SplitRatio >= 1 {
		errs = append(errs, fmt.Errorf("split_ratio must be between 0 and 1: %f, using default %f", cfg.SplitRatio, def.SplitRatio))
		cfg.SplitRatio = def.SplitRatio
	}

	if cfg.Workspaces < 1 {
		errs = append(errs, fmt.Errorf("workspaces must be at least 1: %d, using default %d", cfg.Workspaces, def.Workspaces))
		cfg.Workspaces = def.Workspaces
	}

	if len(cfg.Keys) == 0 {
		errs = append(errs, fmt.Errorf("keys missing, using defaults"))
		cfg.Keys = def.Keys
	}

	return errs
}
