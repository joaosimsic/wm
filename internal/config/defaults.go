package config

import "fmt"

const defaultWorkspaces = 9

func Default() Config {
	return Config{
		Terminal:    "xterm",
		Font:        "fixed",
		BorderWidth: 2,
		Gap:         0,
		BarHeight:   22,
		Mod:         "Mod4",
		SplitRatio:  0.5,
		Workspaces:  defaultWorkspaces,
		Keys:        DefaultKeys(),
		Colors:      DefaultColors(),
	}
}

func DefaultKeys() map[string]string {
	keys := map[string]string{
		"spawn":         "Return",
		"close":         "q",
		"quit":          "Shift+q",
		"focus_next":    "j",
		"focus_prev":    "k",
		"swap_next":     "Shift+j",
		"swap_prev":     "Shift+k",
		"focus_master":  "m",
		"grow_master":   "l",
		"shrink_master": "h",
	}

	for i := range defaultWorkspaces {
		keys[fmt.Sprintf("workspace_%d", i)] = fmt.Sprintf("%d", i)
		keys[fmt.Sprintf("move_to_workspace_%d", i)] = fmt.Sprintf("Shift+%d", i)
	}

	return keys
}
