package chat

import (
	"image/color"
	"regexp"
	"strconv"
	"strings"
	"unicode"

	"charm.land/lipgloss/v2"
)

const (
	ircBold      = '\x02'
	ircColor     = '\x03'
	ircHexColor  = '\x04'
	ircReset     = '\x0f'
	ircReverse   = '\x16'
	ircItalic    = '\x1d'
	ircUnderline = '\x1f'
)

// https://modern.ircdocs.horse/formatting.html#colors
var ircColors = []color.Color{
	lipgloss.Color("#FFFFFF"),
	lipgloss.Color("#000000"),
	lipgloss.Color("#74A9C4"),
	lipgloss.Color("#009300"),
	lipgloss.Color("#FF0000"),
	lipgloss.Color("#7F0000"),
	lipgloss.Color("#9C009C"),
	lipgloss.Color("#FC7F00"),
	lipgloss.Color("#FFFF00"),
	lipgloss.Color("#00FC00"),
	lipgloss.Color("#009393"),
	lipgloss.Color("#00FFFF"),
	lipgloss.Color("#0000FC"),
	lipgloss.Color("#FF00FF"),
	lipgloss.Color("#7F7F7F"),
	lipgloss.Color("#D2D2D2"),
}

var colorRegex = regexp.MustCompile(`^(\d{1,2})(?:,(\d{1,2}))?`)
var hexColorRegex = regexp.MustCompile(`^#?([[:xdigit:]]{6})`)

type formatState struct {
	bold      bool
	italic    bool
	underline bool
	reverse   bool
	fg        color.Color
	bg        color.Color
	hasFg     bool
	hasBg     bool
}

type styledSegment struct {
	text  string
	style lipgloss.Style
}

type runeRange struct {
	start int
	end   int
}

func parseIRCFormat(text string, baseStyle lipgloss.Style) []styledSegment {
	var segments []styledSegment
	var state formatState
	var currentText strings.Builder

	flushSegment := func() {
		if currentText.Len() > 0 {
			style := buildStyle(baseStyle, state)
			segments = append(segments, styledSegment{
				text:  currentText.String(),
				style: style,
			})
			currentText.Reset()
		}
	}

	i := 0
	runes := []rune(text)
	for i < len(runes) {
		r := runes[i]

		switch r {
		case ircBold:
			flushSegment()
			state.bold = !state.bold
			i++
		case ircItalic:
			flushSegment()
			state.italic = !state.italic
			i++
		case ircUnderline:
			flushSegment()
			state.underline = !state.underline
			i++
		case ircReverse:
			flushSegment()
			state.reverse = !state.reverse
			i++
		case ircHexColor:
			flushSegment()
			i++
			remaining := string(runes[i:])
			if match := hexColorRegex.FindStringSubmatch(remaining); match != nil {
				state.fg = lipgloss.Color("#" + match[1])
				state.hasFg = true
				i += len(match[0])
			} else {
				state.hasFg = false
				state.hasBg = false
			}
		case ircReset:
			flushSegment()
			state = formatState{}
			i++
		case ircColor:
			flushSegment()
			i++
			remaining := string(runes[i:])
			if match := hexColorRegex.FindStringSubmatch(remaining); match != nil {
				state.fg = lipgloss.Color("#" + match[1])
				state.hasFg = true
				i += len(match[0])
			} else if match := colorRegex.FindStringSubmatch(remaining); match != nil {
				if fg, err := strconv.Atoi(match[1]); err == nil && fg < len(ircColors) {
					state.fg = ircColors[fg]
					state.hasFg = true
				}
				if match[2] != "" {
					if bg, err := strconv.Atoi(match[2]); err == nil && bg < len(ircColors) {
						state.bg = ircColors[bg]
						state.hasBg = true
					}
				}
				i += len(match[0])
			} else {
				state.hasFg = false
				state.hasBg = false
			}
		default:
			currentText.WriteRune(r)
			i++
		}
	}

	flushSegment()
	return segments
}

func buildStyle(base lipgloss.Style, state formatState) lipgloss.Style {
	style := base

	if state.bold {
		style = style.Bold(true)
	}
	if state.italic {
		style = style.Italic(true)
	}
	if state.underline {
		style = style.Underline(true)
	}

	fg := state.fg
	bg := state.bg

	if state.reverse {
		fg, bg = bg, fg
		state.hasFg, state.hasBg = state.hasBg, state.hasFg
	}

	if state.hasFg {
		style = style.Foreground(fg)
	}
	if state.hasBg {
		style = style.Background(bg)
	}

	return style
}

func renderIRCFormattedMessage(text string, baseStyle lipgloss.Style) string {
	return renderStyledSegments(parseIRCFormat(text, baseStyle))
}

func renderIRCFormattedMessageWithMentions(text, nickname string, baseStyle, mentionStyle lipgloss.Style) (string, bool) {
	segments := parseIRCFormat(text, baseStyle)
	ranges := findNickMentions(visibleIRCText(segments), nickname)
	if len(ranges) == 0 {
		return renderStyledSegments(segments), false
	}

	var result strings.Builder
	position := 0
	rangeIndex := 0
	for _, seg := range segments {
		runes := []rune(seg.text)
		runStart := 0
		runHighlighted := false
		for i := range runes {
			for rangeIndex < len(ranges) && position >= ranges[rangeIndex].end {
				rangeIndex++
			}
			highlighted := rangeIndex < len(ranges) && position >= ranges[rangeIndex].start
			if i > 0 && highlighted != runHighlighted {
				result.WriteString(renderMentionRun(runes[runStart:i], seg.style, mentionStyle, runHighlighted))
				runStart = i
			}
			runHighlighted = highlighted
			position++
		}
		result.WriteString(renderMentionRun(runes[runStart:], seg.style, mentionStyle, runHighlighted))
	}
	return result.String(), true
}

// MessageMentionsNick reports whether the visible IRC message text contains a
// complete nickname, ignoring formatting control sequences.
func MessageMentionsNick(text, nickname string) bool {
	return len(findNickMentions(visibleIRCText(parseIRCFormat(text, lipgloss.NewStyle())), nickname)) > 0
}

func visibleIRCText(segments []styledSegment) string {
	var text strings.Builder
	for _, seg := range segments {
		text.WriteString(seg.text)
	}
	return text.String()
}

func renderMentionRun(runes []rune, style, mentionStyle lipgloss.Style, highlighted bool) string {
	if len(runes) == 0 {
		return ""
	}
	if highlighted {
		style = style.
			Foreground(mentionStyle.GetForeground()).
			Bold(true)
	}
	return style.Render(string(runes))
}

func renderStyledSegments(segments []styledSegment) string {
	var result strings.Builder
	for _, seg := range segments {
		result.WriteString(seg.style.Render(seg.text))
	}
	return result.String()
}

func findNickMentions(text, nickname string) []runeRange {
	textRunes := []rune(text)
	nickRunes := []rune(nickname)
	if len(nickRunes) == 0 || len(nickRunes) > len(textRunes) {
		return nil
	}

	var ranges []runeRange
	for start := 0; start+len(nickRunes) <= len(textRunes); {
		end := start + len(nickRunes)
		if strings.EqualFold(string(textRunes[start:end]), nickname) &&
			(start == 0 || !isNickRune(textRunes[start-1])) &&
			(end == len(textRunes) || !isNickRune(textRunes[end])) {
			ranges = append(ranges, runeRange{start: start, end: end})
			start = end
			continue
		}
		start++
	}
	return ranges
}

func isNickRune(r rune) bool {
	if unicode.IsLetter(r) || unicode.IsDigit(r) {
		return true
	}
	return strings.ContainsRune("_-[]\\`^{}|", r)
}
