package ui

import (
	"context"
	"errors"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/dexchat/dex/internal/config"
	"github.com/dexchat/dex/internal/irc"
	"github.com/dexchat/dex/internal/notify"
	"github.com/dexchat/dex/internal/ui/components/chat"
)

const desktopNotificationTimeout = 2 * time.Second

// desktopNotificationResultMsg reports a finished Notify call. It carries the
// originating server so errors reach that server's buffer.
type desktopNotificationResultMsg struct {
	server string
	err    error
}

// SetNotifier enables desktop notifications through n.
func (m *Model) SetNotifier(n notify.Notifier) {
	m.notifier = n
}

// desktopCooldownElapsed reports whether buffer key may send a desktop
// notification now and, if so, starts its cooldown.
func (m *Model) desktopCooldownElapsed(key BufferKey) bool {
	now := m.now()
	if last, ok := m.lastDesktopNotifiedAt[key]; ok &&
		now.Sub(last) < m.config.Notifications.Cooldown {
		return false
	}
	m.lastDesktopNotifiedAt[key] = now
	return true
}

// desktopNotificationCmd builds the notification in Update and sends it from
// a command, so the D-Bus call never blocks the update loop.
func (m *Model) desktopNotificationCmd(buf *Buffer, msg irc.BufferNewMessageMsg) tea.Cmd {
	notifier := m.notifier
	server := buf.Server
	n := desktopNotification(msg, m.config.Notifications.ShowBody)
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), desktopNotificationTimeout)
		defer cancel()
		return desktopNotificationResultMsg{server: server, err: notifier.Notify(ctx, n)}
	}
}

// desktopNotification never puts message text in the summary, so disabling
// the body hides it completely.
func desktopNotification(msg irc.BufferNewMessageMsg, showBody bool) notify.Notification {
	n := notify.Notification{Summary: msg.From}
	if !msg.DirectMessage {
		n.Summary += " in " + msg.Buffer
	}
	if showBody {
		n.Body = chat.PlainText(msg.Text)
		if msg.Action {
			n.Body = "* " + msg.From + " " + n.Body
		}
	}
	return n
}

// applyDesktopNotificationResult disables desktop notifications when the
// session bus is unavailable. Errors are shown only when desktop
// notifications were enabled explicitly: the default is best effort, because
// many sessions, such as SSH, have no bus or no notification daemon. Other
// failures are reported once until a notification succeeds again.
func (m *Model) applyDesktopNotificationResult(msg desktopNotificationResultMsg) {
	// Results of commands started before notifications were disabled.
	if m.notifier == nil {
		return
	}
	if msg.err == nil {
		m.desktopErrorReported = false
		return
	}

	unavailable := errors.Is(msg.err, notify.ErrUnavailable)
	if unavailable {
		m.notifier = nil
	}
	if m.config.Notifications.Desktop != config.DesktopOn {
		return
	}

	text := msg.err.Error()
	if unavailable {
		text += "; disabled until dex restarts"
	} else if m.desktopErrorReported {
		return
	} else {
		m.desktopErrorReported = true
	}

	if buf := m.buffers[makeBufferKey(msg.server, "")]; buf != nil {
		m.addCommandError(buf, text)
	}
}
