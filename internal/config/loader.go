package config

import (
	"fmt"
	"os"

	"github.com/BurntSushi/toml"
)

func Load() (*Config, error) {
	data, err := os.ReadFile("config.toml")
	if err != nil {
		return nil, fmt.Errorf("failed to read config.toml: %w", err)
	}

	var cfg Config
	if err := toml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config.toml: %w", err)
	}

	if err := cfg.validate(); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return &cfg, nil
}

func (c *Config) validate() error {
	if len(c.Servers) == 0 {
		return fmt.Errorf("at least one server must be configured")
	}

	for name, server := range c.Servers {
		if server.Address == "" {
			return fmt.Errorf("server %q: address is required", name)
		}
		if server.Port <= 0 || server.Port > 65535 {
			return fmt.Errorf("server %q: port must be between 1 and 65535", name)
		}
	}

	return nil
}
