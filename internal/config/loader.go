package config

import (
	"fmt"
	"os"
	"sort"
	"time"

	"github.com/BurntSushi/toml"
)

func Load() (*Config, error) {
	data, err := os.ReadFile("config.toml")
	if err != nil {
		return nil, fmt.Errorf("failed to read config.toml: %w", err)
	}

	cfg, err := loadFromBytes(data)
	if err != nil {
		return nil, err
	}

	return cfg, nil
}

func loadFromBytes(data []byte) (*Config, error) {
	var raw rawConfig
	md, err := toml.Decode(string(data), &raw)
	if err != nil {
		return nil, fmt.Errorf("failed to parse config.toml: %w", err)
	}
	if undecoded := md.Undecoded(); len(undecoded) > 0 {
		return nil, fmt.Errorf("failed to parse config.toml: unknown fields: %v", undecoded)
	}

	if err := validate(raw.Servers); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	cfg := &Config{
		UI: UI{
			UnreadBadges:       true,
			MentionBadges:      true,
			UnreadOnUserEvents: true,
		},
		Notifications: Notifications{
			Sound: true,
			Events: map[string]bool{
				NotificationMention:       true,
				NotificationDirectMessage: true,
			},
			Cooldown: 2 * time.Second,
		},
	}
	if raw.UI.UnreadBadges != nil {
		cfg.UI.UnreadBadges = *raw.UI.UnreadBadges
	}
	if raw.UI.MentionBadges != nil {
		cfg.UI.MentionBadges = *raw.UI.MentionBadges
	}
	if raw.UI.UnreadOnUserEvents != nil {
		cfg.UI.UnreadOnUserEvents = *raw.UI.UnreadOnUserEvents
	}

	if raw.Notifications.Sound != nil {
		cfg.Notifications.Sound = *raw.Notifications.Sound
	}
	if raw.Notifications.Events != nil {
		cfg.Notifications.Events = make(map[string]bool, len(raw.Notifications.Events))
		for _, event := range raw.Notifications.Events {
			switch event {
			case NotificationMention, NotificationDirectMessage:
				cfg.Notifications.Events[event] = true
			default:
				return nil, fmt.Errorf(
					"config validation failed: unknown notification event %q",
					event)
			}
		}
	}
	if raw.Notifications.Cooldown != "" {
		cooldown, err := time.ParseDuration(raw.Notifications.Cooldown)
		if err != nil {
			return nil, fmt.Errorf(
				"invalid notifications cooldown %q: %w",
				raw.Notifications.Cooldown,
				err,
			)
		}

		if cooldown < 0 {
			return nil, fmt.Errorf(
				"config validation failed: notifications cooldown must not be negative",
			)
		}
		cfg.Notifications.Cooldown = cooldown
	}

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
