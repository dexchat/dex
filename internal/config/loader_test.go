package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestFilePathUsesXDGConfigHome(t *testing.T) {
	configHome := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", configHome)

	got, err := filePath()
	if err != nil {
		t.Fatalf("filePath() error = %v", err)
	}

	want := filepath.Join(configHome, "dex", "config.toml")
	if got != want {
		t.Fatalf("filePath() = %q, want %q", got, want)
	}
}

func TestFilePathDefaultsToUserConfigDirectory(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", "")

	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("os.UserHomeDir() error = %v", err)
	}

	got, err := filePath()
	if err != nil {
		t.Fatalf("filePath() error = %v", err)
	}

	want := filepath.Join(home, ".config", "dex", "config.toml")
	if got != want {
		t.Fatalf("filePath() = %q, want %q", got, want)
	}
}

func TestLoadUsesDefaultConfigWhenFileIsMissing(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if len(cfg.Servers) != 1 {
		t.Fatalf("len(cfg.Servers) = %d, want 1", len(cfg.Servers))
	}
	server := cfg.Servers[0]
	if server.Name != "libera" || server.Address != "irc.libera.chat" || server.Port != 6697 {
		t.Fatalf("default server = %+v", server)
	}
	if server.Nickname == "" || !strings.HasPrefix(server.Nickname, "dex_") {
		t.Fatalf("default nickname = %q, want dex_ prefix", server.Nickname)
	}
	if len(server.Channels) != 1 || server.Channels[0] != "#dexchat" {
		t.Fatalf("default channels = %v, want [#dexchat]", server.Channels)
	}
}

func TestLoadParsesUIDefaultsAndServerOverrides(t *testing.T) {
	cfg, err := loadFromBytes([]byte(`
[ui]
theme = "ayu-dark"
unread_badges = false
mention_badges = true
unread_on_user_events = false

[servers.libera]
address = "irc.libera.chat"
port = 6697
nickname = "dexuser"
channels = ["#go"]
unread_badges = true
ignore_direct_messages_from = ["AlertBot"]
notify_channels = ["#Alerts"]
`))
	if err != nil {
		t.Fatalf("loadFromBytes() error = %v", err)
	}

	if got := cfg.UI.UnreadBadges; got != false {
		t.Fatalf("global unread_badges = %v, want false", got)
	}
	if got := cfg.UI.Theme; got != ThemeAyuDark {
		t.Fatalf("UI theme = %q, want %q", got, ThemeAyuDark)
	}
	if got := cfg.UI.MentionBadges; got != true {
		t.Fatalf("global mention_badges = %v, want true", got)
	}
	if got := cfg.UI.UnreadOnUserEvents; got != false {
		t.Fatalf("global unread_on_user_events = %v, want false", got)
	}

	server := cfg.Servers[0]
	if server.UnreadBadges == nil || *server.UnreadBadges != true {
		t.Fatalf("server unread_badges override = %v, want true", server.UnreadBadges)
	}
	if server.MentionBadges != nil {
		t.Fatalf("server mention_badges override = %v, want nil", *server.MentionBadges)
	}
	if len(server.IgnoreDirectMessagesFrom) != 1 || server.IgnoreDirectMessagesFrom[0] != "AlertBot" {
		t.Fatalf("server ignore_direct_messages_from = %v, want [AlertBot]", server.IgnoreDirectMessagesFrom)
	}
	if !cfg.IgnoresDirectMessageFrom("LIBERA", "alertbot") {
		t.Fatal("expected ignored direct message sender to match case-insensitively")
	}
	if cfg.IgnoresDirectMessageFrom("libera", "other-user") {
		t.Fatal("unexpectedly ignored non-configured direct message sender")
	}
	if !cfg.NotifiesChannel("LIBERA", "#alerts") {
		t.Fatal("expected notify channel to match case-insensitively")
	}
	if cfg.NotifiesChannel("libera", "#other") {
		t.Fatal("unexpectedly configured notification for another channel")
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
	if got := cfg.UI.Theme; got != ThemeRosePine {
		t.Fatalf("UI theme default = %q, want %q", got, ThemeRosePine)
	}
	if got := cfg.UI.MentionBadges; got != true {
		t.Fatalf("global mention_badges default = %v, want true", got)
	}
	if got := cfg.UI.UnreadOnUserEvents; got != true {
		t.Fatalf("global unread_on_user_events default = %v, want true", got)
	}
}

func TestLoadRejectsUnknownUITheme(t *testing.T) {
	_, err := loadFromBytes([]byte(`
[ui]
theme = "rose-pink"

[servers.libera]
address = "irc.libera.chat"
port = 6697
nickname = "dexuser"
channels = ["#go"]
`))
	if err == nil || !strings.Contains(err.Error(), `unknown UI theme "rose-pink"`) {
		t.Fatalf("expected unknown UI theme error, got %v", err)
	}
}

func TestLoadAcceptsGruvboxTheme(t *testing.T) {
	cfg, err := loadFromBytes([]byte(`
[ui]
theme = "gruvbox-dark"

[servers.libera]
address = "irc.libera.chat"
port = 6697
nickname = "dexuser"
channels = ["#go"]
`))
	if err != nil {
		t.Fatalf("loadFromBytes() error = %v", err)
	}
	if got := cfg.UI.Theme; got != ThemeGruvboxDark {
		t.Fatalf("UI theme = %q, want %q", got, ThemeGruvboxDark)
	}
}

func TestLoadDefaultsNotifications(t *testing.T) {
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

	if !cfg.Notifications.Sound {
		t.Fatal("Notifications.Sound = false, want true")
	}
	if !cfg.Notifications.Events[NotificationMention] {
		t.Fatal("mention notifications disabled by default, want enabled")
	}
	if !cfg.Notifications.Events[NotificationDirectMessage] {
		t.Fatal("direct message notifications disabled by default, want enabled")
	}
	if got, want := cfg.Notifications.Cooldown, 2*time.Second; got != want {
		t.Fatalf("Notifications.Cooldown = %v, want %v", got, want)
	}
}

func TestLoadParsesNotifications(t *testing.T) {
	cfg, err := loadFromBytes([]byte(`
[notifications]
sound = false
events = ["mention"]
cooldown = "5s"

[servers.libera]
address = "irc.libera.chat"
port = 6697
nickname = "dexuser"
channels = ["#go"]
`))
	if err != nil {
		t.Fatalf("loadFromBytes() error = %v", err)
	}

	if cfg.Notifications.Sound {
		t.Fatal("Notifications.Sound = true, want false")
	}
	if !cfg.Notifications.Events[NotificationMention] {
		t.Fatal("mention notifications disabled, want enabled")
	}
	if cfg.Notifications.Events[NotificationDirectMessage] {
		t.Fatal("direct message notifications enabled, want disabled")
	}
	if got, want := cfg.Notifications.Cooldown, 5*time.Second; got != want {
		t.Fatalf("Notifications.Cooldown = %v, want %v", got, want)
	}
}

func TestLoadAllowsEmptyNotificationEvents(t *testing.T) {
	cfg, err := loadFromBytes([]byte(`
[notifications]
events = []

[servers.libera]
address = "irc.libera.chat"
port = 6697
nickname = "dexuser"
channels = ["#go"]
`))
	if err != nil {
		t.Fatalf("loadFromBytes() error = %v", err)
	}
	if got := len(cfg.Notifications.Events); got != 0 {
		t.Fatalf("len(Notifications.Events) = %d, want 0", got)
	}
}

func TestLoadRejectsUnknownNotificationEvent(t *testing.T) {
	_, err := loadFromBytes([]byte(`
[notifications]
events = ["mentoin"]

[servers.libera]
address = "irc.libera.chat"
port = 6697
nickname = "dexuser"
channels = ["#go"]
`))
	if err == nil || !strings.Contains(err.Error(), `unknown notification event "mentoin"`) {
		t.Fatalf("expected unknown notification event error, got %v", err)
	}
}

func TestLoadRejectsInvalidNotificationCooldown(t *testing.T) {
	_, err := loadFromBytes([]byte(`
[notifications]
cooldown = "soon"

[servers.libera]
address = "irc.libera.chat"
port = 6697
nickname = "dexuser"
channels = ["#go"]
`))
	if err == nil || !strings.Contains(err.Error(), `invalid notifications cooldown "soon"`) {
		t.Fatalf("expected invalid notification cooldown error, got %v", err)
	}
}

func TestLoadRejectsNegativeNotificationCooldown(t *testing.T) {
	_, err := loadFromBytes([]byte(`
[notifications]
cooldown = "-2s"

[servers.libera]
address = "irc.libera.chat"
port = 6697
nickname = "dexuser"
channels = ["#go"]
`))
	if err == nil || !strings.Contains(err.Error(), "notifications cooldown must not be negative") {
		t.Fatalf("expected negative notification cooldown error, got %v", err)
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
