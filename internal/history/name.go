package history

import "strings"

// NameKey returns the form of an IRC server, channel, or nick name used for
// keys and comparisons. IRC names are case-insensitive; the result is not
// meant for display.
func NameKey(name string) string {
	return strings.ToLower(name)
}
