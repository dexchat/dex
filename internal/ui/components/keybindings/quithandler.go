package keybindings

import "time"

type quitHandler struct {
	pressCount int
	lastPress  time.Time
	timeout    time.Duration
	now        func() time.Time
}

var qh = &quitHandler{
	timeout: 1500 * time.Millisecond,
	now:     time.Now,
}

func QuitHandler() bool {
	return qh.press(qh.now())
}

func (h *quitHandler) press(now time.Time) bool {

	// Reset if the timeout has passed since the first press.
	if h.pressCount > 0 && now.Sub(h.lastPress) > h.timeout {
		h.pressCount = 0
	}

	h.pressCount++
	if h.pressCount == 1 {
		h.lastPress = now
	}

	return h.pressCount >= 2
}

// QuitHandlerIsWaitingForSecondPress reports whether a second press is pending.
func QuitHandlerIsWaitingForSecondPress() bool {
	return qh.waiting(qh.now())
}

func (h *quitHandler) waiting(now time.Time) bool {
	if h.pressCount == 0 {
		return false
	}
	// Check if it hasn't timed out yet.
	return now.Sub(h.lastPress) <= h.timeout
}
