package notify

import (
	"context"
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/godbus/dbus/v5"
)

const (
	appName      = "dex"
	maxBodyRunes = 300
	busName      = "org.freedesktop.Notifications"
	objectPath   = dbus.ObjectPath("/org/freedesktop/Notifications")
	notifyMethod = busName + ".Notify"
)

// The notification daemon interprets a subset of markup in the body only, so
// the summary is passed through unchanged.
var bodyEscaper = strings.NewReplacer("&", "&amp;", "<", "&lt;", ">", "&gt;")

// DBus sends notifications through the org.freedesktop.Notifications
// interface on the session bus. It connects once; the connection and the
// connection error never change afterwards.
type DBus struct {
	conn *dbus.Conn
	err  error
}

// NewDBus connects to the session bus. A connection failure is not returned
// here; every Notify call reports it wrapped in ErrUnavailable.
func NewDBus() *DBus {
	conn, err := connectSessionBus()
	return &DBus{conn: conn, err: err}
}

// connectSessionBus connects without autolaunching a bus. Autolaunch runs
// dbus-launch, which would start an orphan bus with no notification daemon
// in sessions without one, such as over SSH.
func connectSessionBus() (*dbus.Conn, error) {
	conn, err := dbus.SessionBusPrivateNoAutoStartup()
	if err != nil {
		return nil, err
	}
	if err := conn.Auth(nil); err != nil {
		_ = conn.Close()
		return nil, err
	}
	if err := conn.Hello(); err != nil {
		_ = conn.Close()
		return nil, err
	}
	return conn, nil
}

// Notify sends n to the notification daemon.
func (d *DBus) Notify(ctx context.Context, n Notification) error {
	if d.err != nil {
		return fmt.Errorf("%w: connect to session bus: %w", ErrUnavailable, d.err)
	}
	obj := d.conn.Object(busName, objectPath)
	if err := obj.CallWithContext(ctx, notifyMethod, 0, notifyArgs(n)...).Err; err != nil {
		return fmt.Errorf("send desktop notification: %w", err)
	}
	return nil
}

// Close closes the session bus connection, if any.
func (d *DBus) Close() error {
	if d.conn == nil {
		return nil
	}
	return d.conn.Close()
}

// notifyArgs builds the arguments of org.freedesktop.Notifications.Notify:
// app_name, replaces_id, app_icon, summary, body, actions, hints and
// expire_timeout.
func notifyArgs(n Notification) []any {
	return []any{
		appName,
		uint32(0),
		"",
		n.Summary,
		formatBody(n.Body),
		[]string{},
		map[string]dbus.Variant{"category": dbus.MakeVariant("im.received")},
		int32(-1),
	}
}

// formatBody truncates before escaping so the cut never splits an entity.
func formatBody(body string) string {
	if utf8.RuneCountInString(body) > maxBodyRunes {
		runes := []rune(body)
		body = string(runes[:maxBodyRunes-1]) + "…"
	}
	return bodyEscaper.Replace(body)
}
