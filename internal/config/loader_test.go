package config

import (
	"strings"
	"testing"
)

func TestLoadParsesUIDefaultsAndServerOverrides(t *testing.T) {
	cfg, err := loadFromBytes([]byte(`
[ui]
unread_badges = false
mention_badges = true

[servers.libera]
address = "irc.libera.chat"
port = 6697
nickname = "dexuser"
channels = ["#go"]
unread_badges = true
`))
	if err != nil {
		t.Fatalf("loadFromBytes() error = %v", err)
	}

	if got := cfg.UI.UnreadBadges; got != false {
		t.Fatalf("global unread_badges = %v, want false", got)
	}
	if got := cfg.UI.MentionBadges; got != true {
		t.Fatalf("global mention_badges = %v, want true", got)
	}

	server := cfg.Servers[0]
	if server.UnreadBadges == nil || *server.UnreadBadges != true {
		t.Fatalf("server unread_badges override = %v, want true", server.UnreadBadges)
	}
	if server.MentionBadges != nil {
		t.Fatalf("server mention_badges override = %v, want nil", *server.MentionBadges)
	}
}

func TestLoadDefaultsUIBadgesToEnabled(t *testing.T) {
	cfg, err := loadFromBytes([]byte(`
[servers.libera]
address = "irc.libera.chat"
port = 6697
nickname = "dexuser"
channels = ["#go"]
`))
	if err != nil {
		t.Fatalf("loadFromBytes() error = %v", err)
	}

	if got := cfg.UI.UnreadBadges; got != true {
		t.Fatalf("global unread_badges default = %v, want true", got)
	}
	if got := cfg.UI.MentionBadges; got != true {
		t.Fatalf("global mention_badges default = %v, want true", got)
	}
}

func TestServerBadgeSettingsFallsBackToGlobal(t *testing.T) {
	cfg := &Config{
		UI: UI{
			UnreadBadges:  false,
			MentionBadges: true,
		},
	}
	server := &Server{Name: "libera"}

	if got := cfg.ServerBadgeSettings(server); got != (BadgeSettings{Unread: false, Mention: true}) {
		t.Fatalf("ServerBadgeSettings() = %+v", got)
	}
}

func TestServerBadgeSettingsUsesOverrides(t *testing.T) {
	cfg := &Config{
		UI: UI{
			UnreadBadges:  true,
			MentionBadges: true,
		},
	}
	server := &Server{
		Name:          "libera",
		UnreadBadges:  boolPtr(false),
		MentionBadges: boolPtr(false),
	}

	if got := cfg.ServerBadgeSettings(server); got != (BadgeSettings{Unread: false, Mention: false}) {
		t.Fatalf("ServerBadgeSettings() = %+v", got)
	}
}

func TestLoadRejectsUnknownTopLevelSections(t *testing.T) {
	_, err := loadFromBytes([]byte(`
[global]
unread_badges = false

[servers.libera]
address = "irc.libera.chat"
port = 6697
nickname = "dexuser"
channels = ["#go"]
`))
	if err == nil || !strings.Contains(err.Error(), "global") {
		t.Fatalf("expected unknown top-level section error, got %v", err)
	}
}

func boolPtr(v bool) *bool {
	return &v
}
