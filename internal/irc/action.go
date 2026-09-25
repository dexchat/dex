package irc

import "strings"

// ctcpDelim frames a CTCP message inside a PRIVMSG.
const ctcpDelim = "\x01"

// encodeAction frames text as a CTCP ACTION, the PRIVMSG sent by /me.
func encodeAction(text string) string {
	return ctcpDelim + "ACTION " + text + ctcpDelim
}

// decodeAction reports whether a PRIVMSG text is a CTCP ACTION and returns
// its text without the CTCP framing. Some clients omit the closing
// delimiter, so it is optional.
func decodeAction(text string) (string, bool) {
	body, ok := strings.CutPrefix(text, ctcpDelim+"ACTION")
	if !ok {
		return text, false
	}
	body = strings.TrimSuffix(body, ctcpDelim)
	if body == "" {
		return "", true
	}
	body, ok = strings.CutPrefix(body, " ")
	if !ok {
		// Another CTCP command that starts with ACTION, such as ACTIONS.
		return text, false
	}
	return body, true
}
