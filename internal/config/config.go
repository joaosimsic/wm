package config

import (
	"errors"
	"fmt"
	"os"
)

type Config struct {
	Terminal    string `toml:"terminal"`
	Font        string `toml:"font"`
	BorderWidth int    `toml:"border_width"`
	Gap         int    `toml:"gap"`
	BarHeight   int    `toml:"bar_height"`
	Colors      Colors `toml:"colors"`
}

func Load() (*Config, error) {
	cfg := Default()

	path, err := defaultPath()
	if err != nil {
		return &cfg, fmt.Errorf("config path: %w", err)
	}

	file, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return &cfg, fmt.Errorf("config file %s not found, using defaults", path)
	}
	if err != nil {
		return &cfg, fmt.Errorf("read config %s: %w, using defaults", path, err)
	}

	if err := parse(file, &cfg); err != nil {
		cfg = Default()
		return &cfg, fmt.Errorf("parse config %s: %w, using defaults", path, err)
	}

	return &cfg, errors.Join(validate(&cfg)...)
}
