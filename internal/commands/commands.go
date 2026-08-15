package commands

import "strings"

type Command struct {
	Name string
	Args []string
}

// Parse distinguishes slash commands from regular chat messages.
// Unknown slash commands are still returned as commands so callers can report
// them instead of accidentally sending them to IRC.
func Parse(input string) (Command, bool) {
	input = strings.TrimSpace(input)
	if !strings.HasPrefix(input, "/") {
		return Command{}, false
	}

	fields := strings.Fields(input)
	if len(fields) == 0 {
		return Command{}, true
	}

	return Command{
		Name: strings.ToLower(strings.TrimPrefix(fields[0], "/")),
		Args: fields[1:],
	}, true
}
