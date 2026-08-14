package config

func Default() Config {
	return Config{
		Terminal:    "xterm",
		Font:        "fixed",
		BorderWidth: 2,
		Gap:         0,
		BarHeight:   22,
		Colors:      DefaultColors(),
	}
}
