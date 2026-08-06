package chat

import (
	"strings"
	"unicode"
)

type nicknameCompletion struct {
	active     bool
	start      int
	end        int
	candidates []string
	index      int
}

func (m *Model) completeNickname() {
	if m.channelMembers == nil {
		return
	}

	value := []rune(m.input.Value())
	if m.completion.active && m.input.Position() == m.completion.end {
		m.completion.index = (m.completion.index + 1) % len(m.completion.candidates)
		m.applyNicknameCompletion(value)
		return
	}

	end := m.input.Position()
	start := end
	for start > 0 && !unicode.IsSpace(value[start-1]) {
		start--
	}
	if start == end {
		return
	}

	prefix := strings.ToLower(string(value[start:end]))
	candidates := make([]string, 0)
	for _, nick := range m.channelMembers.Nicknames() {
		if strings.HasPrefix(strings.ToLower(nick), prefix) {
			candidates = append(candidates, nick)
		}
	}
	if len(candidates) == 0 {
		return
	}

	m.completion = nicknameCompletion{
		active:     true,
		start:      start,
		end:        end,
		candidates: candidates,
	}
	m.applyNicknameCompletion(value)
}

func (m *Model) applyNicknameCompletion(value []rune) {
	replacement := m.completion.candidates[m.completion.index]
	if m.completion.start == 0 && m.completion.end == len(value) {
		replacement += ": "
	}

	replacementRunes := []rune(replacement)
	completed := make([]rune, 0, len(value)-(m.completion.end-m.completion.start)+len(replacementRunes))
	completed = append(completed, value[:m.completion.start]...)
	completed = append(completed, replacementRunes...)
	completed = append(completed, value[m.completion.end:]...)

	m.completion.end = m.completion.start + len(replacementRunes)
	m.input.SetValue(string(completed))
	m.input.SetCursor(m.completion.end)
}
