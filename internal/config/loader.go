package config

import (
	"crypto/rand"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/BurntSushi/toml"
)

func Load() (*Config, error) {
	path, err := filePath()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return defaultConfig()
		}
		return nil, fmt.Errorf("failed to read %s: %w", path, err)
	}

	cfg, err := loadFromBytes(data)
	if err != nil {
		return nil, err
	}

	return cfg, nil
}

func defaultConfig() (*Config, error) {
	nickname, err := randomNickname()
	if err != nil {
		return nil, fmt.Errorf("failed to generate default nickname: %w", err)
	}

	cfg := defaultValues()
	cfg.Servers = []*Server{{
		Name:     "libera",
		Address:  "irc.libera.chat",
		Port:     6697,
		Nickname: nickname,
		Channels: []string{"#dexchat"},
	}}
	return cfg, nil
}

func randomNickname() (string, error) {
	var suffix [3]byte
	if _, err := rand.Read(suffix[:]); err != nil {
		return "", err
	}
	return fmt.Sprintf("dex_%02x%02x%02x", suffix[0], suffix[1], suffix[2]), nil
}

func filePath() (string, error) {
	configDir := os.Getenv("XDG_CONFIG_HOME")
	if configDir == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("failed to determine config directory: %w", err)
		}
		configDir = filepath.Join(home, ".config")
	}

	return filepath.Join(configDir, "dex", "config.toml"), nil
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

	cfg := defaultValues()
	if raw.UI.Theme != "" {
		switch raw.UI.Theme {
		case ThemeRosePine, ThemeAyuDark, ThemeDracula, ThemeGruvboxDark, ThemeSolarizedLight:
			cfg.UI.Theme = raw.UI.Theme
		default:
			return nil, fmt.Errorf("config validation failed: unknown UI theme %q", raw.UI.Theme)
		}
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
				return nil, fmt.Errorf("config validation failed: unknown notification event %q", event)
			}
		}
	}
	if raw.Notifications.Cooldown != "" {
		cooldown, err := time.ParseDuration(raw.Notifications.Cooldown)
		if err != nil {
			return nil, fmt.Errorf("invalid notifications cooldown %q: %w", raw.Notifications.Cooldown, err)
		}
		if cooldown < 0 {
			return nil, fmt.Errorf("config validation failed: notifications cooldown must not be negative")
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

func defaultValues() *Config {
	return &Config{
		UI:            UI{Theme: ThemeRosePine, UnreadBadges: true, MentionBadges: true, UnreadOnUserEvents: true},
		Notifications: Notifications{Sound: true, Events: map[string]bool{NotificationMention: true, NotificationDirectMessage: true}, Cooldown: 2 * time.Second},
	}
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
		if server.SSLSkipVerify && !server.UseSSL() {
			return fmt.Errorf("server %q: ssl_skip_verify requires ssl", name)
		}
	}

	return nil
}
