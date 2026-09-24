package irc

import (
	"testing"
	"time"
)

func TestSentMessagesForgetsExpiredSends(t *testing.T) {
	var sent sentMessages
	start := time.Now()
	sent.add("libera:#go:hello", start)

	if sent.consume("libera:#go:hello", start.Add(sentMessageTTL+time.Second)) {
		t.Fatal("an expired send was still pending")
	}
	if len(sent.pending) != 0 {
		t.Fatalf("expired sends were kept: %v", sent.pending)
	}
}

func TestSentMessagesKeepsRecentSendsWhenPruning(t *testing.T) {
	var sent sentMessages
	start := time.Now()
	sent.add("libera:#go:old", start)
	sent.add("libera:#go:recent", start.Add(sentMessageTTL))

	now := start.Add(sentMessageTTL + time.Second)
	if sent.consume("libera:#go:old", now) {
		t.Fatal("an expired send was still pending")
	}
	if !sent.consume("libera:#go:recent", now) {
		t.Fatal("a recent send was forgotten while pruning an expired one")
	}
}
