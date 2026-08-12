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

type Colors struct {
	BarBg          string `toml:"bar_bg"`
	BarFg          string `toml:"bar_fg"`
	BarActiveBg    string `toml:"bar_active_bg"`
	BorderActive   string `toml:"border_active"`
	BorderInactive string `toml:"border_inactive"`
}

func Load() (*Config, error) {
	cfg := defaultConfig()

	path, err := defaultPath()
	if err != nil {
		return &cfg, fmt.Errorf("get config path: %w", err)
	}

	file, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return &cfg, nil
	}
	if err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}

	return &cfg, parse(file, &cfg)
}
