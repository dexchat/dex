package irc

import (
	"context"
	"sync"
	"time"
)

// eventBatchWindow is how long the oldest queued event waits for later events
// to join its batch. Bouncer playback can deliver thousands of lines in a
// burst; batching them keeps the UI from running one update per line.
const eventBatchWindow = 50 * time.Millisecond

// eventQueue is an unbounded FIFO of events for one server. Producers never
// block, so IRC handlers cannot stall the socket reader while the UI is busy.
// The zero value is ready to use.
type eventQueue struct {
	mu       sync.Mutex
	events   []Event
	oldestAt time.Time
	ready    chan struct{}
}

func (q *eventQueue) push(events ...Event) {
	if len(events) == 0 {
		return
	}

	q.mu.Lock()
	defer q.mu.Unlock()

	if len(q.events) == 0 {
		q.oldestAt = time.Now()
	}
	q.events = append(q.events, events...)
	select {
	case q.readyChan() <- struct{}{}:
	default:
	}
}

// next blocks until events are queued, waits until the oldest one has been
// queued for at least window, and then returns every queued event in order.
func (q *eventQueue) next(ctx context.Context, window time.Duration) ([]Event, error) {
	for {
		q.mu.Lock()
		ready := q.readyChan()
		if len(q.events) == 0 {
			q.mu.Unlock()
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-ready:
			}
			continue
		}
		wait := window - time.Since(q.oldestAt)
		q.mu.Unlock()

		if wait > 0 {
			timer := time.NewTimer(wait)
			select {
			case <-ctx.Done():
				timer.Stop()
				return nil, ctx.Err()
			case <-timer.C:
			}
		}

		q.mu.Lock()
		events := q.events
		q.events = nil
		q.mu.Unlock()
		if len(events) > 0 {
			return events, nil
		}
	}
}

// readyChan returns the wake-up channel, creating it on first use. The caller
// must hold q.mu.
func (q *eventQueue) readyChan() chan struct{} {
	if q.ready == nil {
		q.ready = make(chan struct{}, 1)
	}
	return q.ready
}
