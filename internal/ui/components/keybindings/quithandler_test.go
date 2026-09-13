package keybindings

import (
	"testing"
	"time"
)

func TestQuitHandlerRequiresTwoPresses(t *testing.T) {
	h := quitHandler{timeout: time.Second}
	now := time.Unix(100, 0)

	if h.press(now) {
		t.Fatal("first press should not quit")
	}
	if !h.waiting(now) {
		t.Fatal("first press should wait for a second press")
	}
	if !h.press(now.Add(100 * time.Millisecond)) {
		t.Fatal("second press within timeout should quit")
	}
}

func TestQuitHandlerResetsAfterTimeout(t *testing.T) {
	h := quitHandler{timeout: time.Second}
	now := time.Unix(100, 0)

	if h.press(now) {
		t.Fatal("first press should not quit")
	}
	if h.waiting(now.Add(2 * time.Second)) {
		t.Fatal("expired press should not remain pending")
	}
	if h.press(now.Add(2 * time.Second)) {
		t.Fatal("press after timeout should start a new sequence")
	}
}
