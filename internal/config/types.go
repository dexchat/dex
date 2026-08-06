package config

import "time"

const (
	NotificationMention       = "mention"
	NotificationDirectMessage = "direct_message"
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
	UnreadBadges  *bool `toml:"unread_badges"`
	MentionBadges *bool `toml:"mention_badges"`
}

type UI struct {
	UnreadBadges  bool
	MentionBadges bool
}

type BadgeSettings struct {
	Unread  bool
	Mention bool
}

type Server struct {
	Name          string
	Address       string   `toml:"address"`
	Port          int      `toml:"port"`
	SSL           *bool    `toml:"ssl"`
	Password      string   `toml:"password"`
	Channels      []string `toml:"channels"`
	Nickname      string   `toml:"nickname"`
	Username      string   `toml:"username"`
	Realname      string   `toml:"realname"`
	UnreadBadges  *bool    `toml:"unread_badges"`
	MentionBadges *bool    `toml:"mention_badges"`
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
