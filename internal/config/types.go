package config

import (
	"strings"
	"time"
)

const (
	NotificationMention       = "mention"
	NotificationDirectMessage = "direct_message"
	ThemeRosePine             = "rose-pine"
	ThemeAyuDark              = "ayu-dark"
	ThemeDracula              = "dracula"
	ThemeGruvboxDark          = "gruvbox-dark"
	ThemeSolarizedLight       = "solarized-light"
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
	Events   []string `toml:"events"`
	Cooldown string   `toml:"cooldown"`
}

type Notifications struct {
	Sound    bool
	Events   map[string]bool
	Cooldown time.Duration
}

type rawUI struct {
	Theme              string `toml:"theme"`
	UnreadBadges       *bool  `toml:"unread_badges"`
	MentionBadges      *bool  `toml:"mention_badges"`
	UnreadOnUserEvents *bool  `toml:"unread_on_user_events"`
}

type UI struct {
	Theme              string
	UnreadBadges       bool
	MentionBadges      bool
	UnreadOnUserEvents bool
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
