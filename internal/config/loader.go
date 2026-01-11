package config

import (
	"fmt"
	"os"
	"sort"

	"github.com/BurntSushi/toml"
)

func Load() (*Config, error) {
	data, err := os.ReadFile("config.toml")
	if err != nil {
		return nil, fmt.Errorf("failed to read config.toml: %w", err)
	}

	var raw rawConfig
	if err := toml.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("failed to parse config.toml: %w", err)
	}

	if err := validate(raw.Servers); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	cfg := &Config{}
	for name, server := range raw.Servers {
		server.Name = name
		sort.Strings(server.Channels)
		cfg.Servers = append(cfg.Servers, server)
	}
	sort.Slice(cfg.Servers, func(i, j int) bool {
		return cfg.Servers[i].Name < cfg.Servers[j].Name
	})

	return cfg, nil
}

func validate(servers map[string]*Server) error {
	if len(servers) == 0 {
		return fmt.Errorf("at least one server must be configured")
	}

	for name, server := range servers {
		if server.Address == "" {
			return fmt.Errorf("server %q: address is required", name)
		}
		if server.Port <= 0 || server.Port > 65535 {
			return fmt.Errorf("server %q: port must be between 1 and 65535", name)
		}
	}

	return nil
}
