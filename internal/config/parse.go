package config

import (
	"fmt"

	"github.com/BurntSushi/toml"
)

func parse(file []byte, cfg *Config) error {
    if err := toml.Unmarshal(file, cfg); err != nil {
        return fmt.Errorf("unmarshal config: %w", err)
    }

    return validate(cfg)
}
