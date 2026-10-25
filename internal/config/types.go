package config

import (
	"fmt"
	"strings"
	"time"
)

const (
	NotificationMention       = "mention"
	NotificationDirectMessage = "direct_message"
	ThemeRosePine             = "rose-pine"
	ThemeRosePineDawn         = "rose-pine-dawn"
	ThemeAyuDark              = "ayu-dark"
	ThemeDracula              = "dracula"
	ThemeGruvboxDark          = "gruvbox-dark"
	ThemeSolarizedLight       = "solarized-light"
	ThemeNone                 = "none"
)

// DesktopMode controls desktop notifications.
type DesktopMode int

const (
	// DesktopOff never touches D-Bus.
	DesktopOff DesktopMode = iota
	// DesktopAuto is the default: notify when a session bus exists and turn
	// off silently when it does not, such as on macOS or over SSH.
	DesktopAuto
	// DesktopOn is set explicitly and reports a missing session bus.
	DesktopOn
)

// rawConfig stores the config file as it is
type rawConfig struct {
	UI            rawUI              `toml:"ui"`
	Notifications rawNotifications   `toml:"notifications"`
	Servers       map[string]*Server `toml:"servers"`
}

// Config stores the config file sorted and ready to be used
type Config struct {
	UI UI

	Notifications Notifications

	// Servers is used to create the buffers on startup only
	// Later, the source of truth is the irc client channels
	// as the user can join channels while using the app
	Servers []*Server
}

type rawNotifications struct {
	Sound    *bool    `toml:"sound"`
	Desktop  *bool    `toml:"desktop"`
	ShowBody *bool    `toml:"show_body"`
	Events   []string `toml:"events"`
	Cooldown string   `toml:"cooldown"`
}

type Notifications struct {
	Sound    bool
	Desktop  DesktopMode
	ShowBody bool
	Events   map[string]bool
	Cooldown time.Duration
}

type rawUI struct {
	Theme              rawTheme `toml:"theme"`
	UnreadBadges       *bool    `toml:"unread_badges"`
	MentionBadges      *bool    `toml:"mention_badges"`
	UnreadOnUserEvents *bool    `toml:"unread_on_user_events"`
}

// rawTheme stores ui.theme, which is either a theme name or a table with
// light and dark theme names.
type rawTheme struct {
	Name  string
	Light string
	Dark  string
}

type UI struct {
	Theme string
	// SystemTheme is set when ui.theme is a light/dark table. It replaces
	// Theme with Light or Dark to follow the system appearance.
	SystemTheme        *SystemTheme
	UnreadBadges       bool
	MentionBadges      bool
	UnreadOnUserEvents bool
}

type SystemTheme struct {
	Light string
	Dark  string
}

// UnmarshalTOML decodes a theme name or a table with light and dark theme
// names. Names are validated by the loader.
func (t *rawTheme) UnmarshalTOML(value any) error {
	switch value := value.(type) {
	case string:
		t.Name = value
	case map[string]any:
		for key, field := range value {
			name, ok := field.(string)
			if !ok {
				return fmt.Errorf("theme.%s must be a theme name", key)
			}
			switch key {
			case "light":
				t.Light = name
			case "dark":
				t.Dark = name
			default:
				return fmt.Errorf("unknown theme field %q, want light or dark", key)
			}
		}
		if t.Light == "" || t.Dark == "" {
			return fmt.Errorf("theme table needs both light and dark")
		}
	default:
		return fmt.Errorf("theme must be a theme name or a table with light and dark")
	}
	return nil
}

type BadgeSettings struct {
	Unread  bool
	Mention bool
}

type Server struct {
	Name                     string
	Address                  string   `toml:"address"`
	Port                     int      `toml:"port"`
	SSL                      *bool    `toml:"ssl"`
	SSLSkipVerify            bool     `toml:"ssl_skip_verify"`
	Password                 string   `toml:"password"`
	Channels                 []string `toml:"channels"`
	Nickname                 string   `toml:"nickname"`
	Username                 string   `toml:"username"`
	Realname                 string   `toml:"realname"`
	UnreadBadges             *bool    `toml:"unread_badges"`
	MentionBadges            *bool    `toml:"mention_badges"`
	IgnoreDirectMessagesFrom []string `toml:"ignore_direct_messages_from"`
	NotifyChannels           []string `toml:"notify_channels"`
}

// IgnoresDirectMessageFrom reports whether a private message from nick should
// be excluded from notifications on the named server.
func (c *Config) IgnoresDirectMessageFrom(serverName, nick string) bool {
	if c == nil {
		return false
	}
	for _, server := range c.Servers {
		if !strings.EqualFold(server.Name, serverName) {
			continue
		}
		for _, ignoredNick := range server.IgnoreDirectMessagesFrom {
			if strings.EqualFold(ignoredNick, nick) {
				return true
			}
		}
	}
	return false
}

// NotifiesChannel reports whether every new message in the channel should
// trigger a notification on the named server.
func (c *Config) NotifiesChannel(serverName, channelName string) bool {
	if c == nil {
		return false
	}
	for _, server := range c.Servers {
		if !strings.EqualFold(server.Name, serverName) {
			continue
		}
		for _, channel := range server.NotifyChannels {
			if strings.EqualFold(channel, channelName) {
				return true
			}
		}
	}
	return false
}

func (c *Config) ServerBadgeSettings(server *Server) BadgeSettings {
	settings := BadgeSettings{
		Unread:  c.UI.UnreadBadges,
		Mention: c.UI.MentionBadges,
	}
	if server == nil {
		return settings
	}
	if server.UnreadBadges != nil {
		settings.Unread = *server.UnreadBadges
	}
	if server.MentionBadges != nil {
		settings.Mention = *server.MentionBadges
	}
	return settings
}
