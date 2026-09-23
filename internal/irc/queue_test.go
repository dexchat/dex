package irc

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"
)

func TestEventQueueReturnsEventsInPushOrder(t *testing.T) {
	var queue eventQueue
	want := []Event{
		BufferNewMessageMsg{Server: "libera", Buffer: "#go", Text: "first"},
		ChannelPartedMsg{Server: "libera", Channel: "#go"},
		ChannelJoinedMsg{Server: "libera", Channel: "#rust"},
	}
	queue.push(want[0])
	queue.push(want[1:]...)

	got, err := queue.next(context.Background(), 0)
	if err != nil {
		t.Fatalf("next() error = %v", err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("next() = %#v, want %#v", got, want)
	}
}

func TestEventQueueCollectsEventsPushedDuringBatchWindow(t *testing.T) {
	var queue eventQueue
	queue.push(ChannelJoinedMsg{Server: "libera", Channel: "#go"})

	go func() {
		time.Sleep(10 * time.Millisecond)
		queue.push(ChannelJoinedMsg{Server: "libera", Channel: "#rust"})
	}()

	got, err := queue.next(context.Background(), 200*time.Millisecond)
	if err != nil {
		t.Fatalf("next() error = %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("next() returned %d events, want both events from the window: %#v", len(got), got)
	}
}

func TestEventQueueWaitsForFirstEvent(t *testing.T) {
	var queue eventQueue
	result := make(chan []Event, 1)
	go func() {
		events, _ := queue.next(context.Background(), 0)
		result <- events
	}()

	select {
	case events := <-result:
		t.Fatalf("next() returned %#v before any event was pushed", events)
	case <-time.After(20 * time.Millisecond):
	}

	queue.push(ChannelJoinedMsg{Server: "libera", Channel: "#go"})
	select {
	case events := <-result:
		if len(events) != 1 {
			t.Fatalf("next() returned %d events, want 1", len(events))
		}
	case <-time.After(time.Second):
		t.Fatal("next() did not return after an event was pushed")
	}
}

func TestEventQueueNextStopsWhenContextIsCanceled(t *testing.T) {
	var queue eventQueue
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	if _, err := queue.next(ctx, 0); !errors.Is(err, context.Canceled) {
		t.Fatalf("next() error = %v, want context.Canceled", err)
	}
}

func TestClientManagerNextStopsAfterDisconnectAll(t *testing.T) {
	manager := NewClientManager(nil)
	manager.clients["libera"] = &Client{serverName: "libera"}
	manager.cancel()

	if _, err := manager.Next("libera"); !errors.Is(err, context.Canceled) {
		t.Fatalf("Next() error = %v, want context.Canceled", err)
	}
	if _, err := manager.Next("missing"); err == nil {
		t.Fatal("Next() for an unknown server should fail")
	}
}
