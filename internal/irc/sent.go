package irc

import (
	"strings"
	"sync"
	"time"

	"github.com/lrstanley/girc"
)

// sentMessageTTL bounds how long a sent message waits for its echo. Echoes
// usually arrive within a second, but a busy ZNC queue can delay our own
// messages by minutes. A send whose echo never matches, for example because
// the server changed the text, is forgotten after this long.
const sentMessageTTL = 10 * time.Minute

// sentMessages tracks messages sent from this client whose echo has not
// arrived yet, so the echo is stored but not shown a second time. Each send is
// counted: sending the same text twice expects two echoes.
type sentMessages struct {
	mu      sync.Mutex
	pending map[string][]time.Time
}

func (s *sentMessages) add(key string, now time.Time) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.prune(now)
	if s.pending == nil {
		s.pending = make(map[string][]time.Time)
	}
	s.pending[key] = append(s.pending[key], now)
}

// consume reports whether a send is pending for key and, if so, forgets the
// oldest one.
func (s *sentMessages) consume(key string, now time.Time) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.prune(now)
	sends := s.pending[key]
	switch len(sends) {
	case 0:
		return false
	case 1:
		delete(s.pending, key)
	default:
		s.pending[key] = sends[1:]
	}
	return true
}

// prune forgets sends older than sentMessageTTL. The caller must hold s.mu.
func (s *sentMessages) prune(now time.Time) {
	for key, sends := range s.pending {
		expired := 0
		for expired < len(sends) && now.Sub(sends[expired]) > sentMessageTTL {
			expired++
		}
		switch {
		case expired == len(sends):
			delete(s.pending, key)
		case expired > 0:
			s.pending[key] = sends[expired:]
		}
	}
}

// trackSent prepares for the echo of a message sent to target. message is the
// PRIVMSG text as sent, including any CTCP ACTION framing, because the echo
// carries it too. Without echo-message the server sends no echo, so the
// client queues one itself and the UI stores the message in history as it
// does for a server echo.
func (c *Client) trackSent(target, message string, echoEnabled bool) {
	if echoEnabled {
		c.sent.add(pendingMessageKey(c.serverName, target, message), time.Now())
		return
	}
	text, action := decodeAction(message)
	c.events.push(BufferNewMessageMsg{
		Server:        c.serverName,
		Buffer:        target,
		DirectMessage: !girc.IsValidChannel(target),
		Timestamp:     time.Now(),
		From:          c.GetNick(),
		Text:          text,
		Action:        action,
		OwnEcho:       true,
	})
}

// pendingMessageKey identifies a sent message and its echo. TrimSpace
// normalizes whitespace because IRC servers may strip or add leading and
// trailing spaces, and the echo would then no longer match.
func pendingMessageKey(server, target, message string) string {
	return strings.ToLower(server) + ":" + strings.ToLower(target) + ":" + strings.TrimSpace(message)
}
