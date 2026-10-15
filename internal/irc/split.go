package irc

import (
	"strings"
	"unicode/utf8"

	"github.com/lrstanley/girc"
)

// maxMessageBytes returns how many bytes of text fit in one PRIVMSG to
// target. girc splits longer events itself, and the echoes of those parts
// would no longer match the tracked message, so dex sends each part on its
// own. The limit mirrors girc's check, so girc never splits a part again.
func (c *Client) maxMessageBytes(target string, action bool) int {
	// An empty trailing parameter counts its colon, which a non-empty part
	// may not need, so the overhead is never underestimated.
	overhead := (&girc.Event{Command: girc.PRIVMSG, Params: []string{target, ""}}).LenOpts(false)
	if action {
		overhead += len(encodeAction(""))
	}
	// girc splits unless the event is strictly shorter than its limit.
	return c.MaxEventLength() - overhead - 1
}

// splitMessage splits text into parts of at most limit bytes. It breaks at
// the last space that fits, dropping that space, or else at a UTF-8 boundary.
// Each part holds at least one rune, so a limit smaller than a rune still
// makes progress.
func splitMessage(text string, limit int) []string {
	var parts []string
	for len(text) > limit {
		cut := max(limit, 0)
		for cut > 0 && !utf8.RuneStart(text[cut]) {
			cut--
		}
		next := cut
		// text[cut] is a space when the part fills the limit exactly.
		if space := strings.LastIndexByte(text[:cut+1], ' '); space > 0 {
			cut, next = space, space+1
		} else if cut == 0 {
			_, cut = utf8.DecodeRuneInString(text)
			next = cut
		}
		parts = append(parts, text[:cut])
		text = text[next:]
	}
	if len(parts) == 0 || strings.TrimSpace(text) != "" {
		parts = append(parts, text)
	}
	return parts
}
