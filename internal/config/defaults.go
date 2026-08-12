package config

func Default() Config {
	return Config{
		Terminal:    "xterm",
		Font:        "fixed",
		BorderWidth: 2,
		Gap:         0,
		BarHeight:   22,
		Colors: Colors{
			BarBg:          "#1a1a1a",
			BarFg:          "#cccccc",
			BarActiveBg:    "#4a4a4a",
			BorderActive:   "#5f87af",
			BorderInactive: "#333333",
		},
	}
}
