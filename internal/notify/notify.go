// Package notify delivers desktop notifications. It does not decide when to
// notify or what to show; the UI owns that.
package notify

import (
	"context"
	"errors"
)

// ErrUnavailable wraps failures that make desktop notifications permanently
// unavailable for this process, such as a missing session bus.
var ErrUnavailable = errors.New("desktop notifications unavailable")

// Notification is a desktop notification. Summary and Body are plain text.
type Notification struct {
	Summary string
	Body    string
}

// Notifier delivers desktop notifications. Implementations must be safe for
// concurrent use.
type Notifier interface {
	Notify(ctx context.Context, n Notification) error
}
