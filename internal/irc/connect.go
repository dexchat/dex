package irc

import (
	"context"
	"fmt"
	"time"
)

// reconnectDelays are the waits before each retry after a connection ends.
// The last delay repeats until a connection registers with the server again.
var reconnectDelays = []time.Duration{
	10 * time.Second,
	20 * time.Second,
	40 * time.Second,
	80 * time.Second,
	160 * time.Second,
	300 * time.Second,
}

// backoff chooses the wait before each reconnect attempt. The last delay
// repeats until a connection registers with the server again.
type backoff struct {
	delays  []time.Duration
	attempt int
}

// next returns the wait before the next attempt. A connection that registered
// resets the sequence, so a long-lived connection that drops is retried after
// the shortest delay.
func (b *backoff) next(registered bool) time.Duration {
	if registered {
		b.attempt = 0
	}
	delay := b.delays[min(b.attempt, len(b.delays)-1)]
	b.attempt++
	return delay
}

// connectLoop keeps the client connected until ctx is canceled.
func (c *Client) connectLoop(ctx context.Context, delays []time.Duration) {
	retry := backoff{delays: delays}
	for {
		c.registered.Store(false)
		err := c.Connect()
		// girc returns nil only after Close, which dex calls when shutting
		// down; ctx is canceled first.
		if ctx.Err() != nil || err == nil {
			return
		}

		delay := retry.next(c.registered.Load())
		c.events.push(BufferNewMessageMsg{
			Server:    c.serverName,
			Buffer:    "",
			Timestamp: time.Now(),
			From:      "--",
			Text:      fmt.Sprintf("irc: %v, reconnecting in %d seconds...", err, int(delay/time.Second)),
			Type:      MessageTypeDisconnected,
		})
		if !waitOrDone(ctx, delay) {
			return
		}
	}
}

// waitOrDone waits for d and reports false if ctx ends first.
func waitOrDone(ctx context.Context, d time.Duration) bool {
	timer := time.NewTimer(d)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return false
	case <-timer.C:
		return true
	}
}
