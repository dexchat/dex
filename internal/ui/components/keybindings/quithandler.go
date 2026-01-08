package keybindings

import "time"

type quitHandler struct {
	pressCount int
	lastPress  time.Time
	timeout    time.Duration
}

var qh = &quitHandler{
	timeout: 1500 * time.Millisecond,
}

func QuitHandler() bool {
	now := time.Now()

	// Reset if timeout has passed since the first press
	if qh.pressCount > 0 && now.Sub(qh.lastPress) > qh.timeout {
		qh.pressCount = 0
	}

	qh.pressCount++
	if qh.pressCount == 1 {
		qh.lastPress = now
	}

	return qh.pressCount >= 2
}

// QuitHandlerIsWaitingForSecondPress returns true if the user pressed Ctrl+C and is waiting for the second press
func QuitHandlerIsWaitingForSecondPress() bool {
	if qh.pressCount == 0 {
		return false
	}
	// Check if it hasn't timed out yet
	return time.Now().Sub(qh.lastPress) <= qh.timeout
}
