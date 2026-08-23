package palette

import (
	"fmt"
	"regexp"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/vaaleyard/dex/internal/ui/styles"
)

func TestGoToChannelActionOpensChannelPicker(t *testing.T) {
	m := New(styles.RosePineTheme())
	m.Toggle()

	_, cmd := m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("go to channel action did not return a command")
	}
	msg := cmd()
	if _, ok := msg.(OpenChannelPickerMsg); !ok {
		t.Fatalf("go to channel action returned %T, want OpenChannelPickerMsg", msg)
	}
}

func TestLastBufferActionRequestsLastBuffer(t *testing.T) {
	m := New(styles.RosePineTheme())
	m.Toggle()
	m.moveDown()

	_, cmd := m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("last channel action did not return a command")
	}
	if msg := cmd(); msg != (LastBufferMsg{}) {
		t.Fatalf("last channel action returned %T, want LastBufferMsg", msg)
	}
}

func TestEditInEditorActionRequestsEditor(t *testing.T) {
	m := New(styles.RosePineTheme())
	m.Toggle()
	m.moveDown()
	m.moveDown()

	_, cmd := m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("edit in editor action did not return a command")
	}
	if msg := cmd(); msg != (EditInEditorMsg{}) {
		t.Fatalf("edit in editor action returned %T, want EditInEditorMsg", msg)
	}
}

func TestHelpActionReturnsOpenHelpMsg(t *testing.T) {
	m := New(styles.RosePineTheme())
	m.Toggle()
	// Help is the 4th item (index 3)
	m.moveDown() // 1: LastBuffer
	m.moveDown() // 2: EditInEditor
	m.moveDown() // 3: help

	_, cmd := m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("help action did not return a command")
	}
	if msg := cmd(); msg != (OpenHelpMsg{}) {
		t.Fatalf("help action returned %T, want OpenHelpMsg", msg)
	}
}

func TestPaletteNavigationWithArrowKeys(t *testing.T) {
	m := New(styles.RosePineTheme())
	m.Toggle()

	if m.cursor != 0 {
		t.Fatalf("initial cursor = %d, want 0", m.cursor)
	}

	// Arrow down moves cursor to next item
	m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyDown})
	if m.cursor != 1 {
		t.Fatalf("cursor after KeyDown = %d, want 1", m.cursor)
	}

	// Arrow up moves cursor back to previous item
	m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyUp})
	if m.cursor != 0 {
		t.Fatalf("cursor after KeyUp = %d, want 0", m.cursor)
	}

	// Arrow up wraps around to last item
	m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyUp})
	wantLast := len(m.filteredCommands()) - 1
	if m.cursor != wantLast {
		t.Fatalf("cursor after KeyUp wrap = %d, want %d", m.cursor, wantLast)
	}
}

func TestPaletteWidthsDependOnMode(t *testing.T) {
	m := New(styles.RosePineTheme())
	m.SetSize(90, 12)
	m.Toggle()

	if got, want := lipgloss.Width(m.View()), 91; got != want {
		t.Fatalf("command palette width = %d, want %d", got, want)
	}

	m.ShowChannels([]Channel{{Server: "libera", Name: "#go"}})
	if got, want := lipgloss.Width(m.View()), 64; got != want {
		t.Fatalf("channel picker width = %d, want %d", got, want)
	}
}

func TestChannelPickerSelectsSortedChannel(t *testing.T) {
	m := New(styles.RosePineTheme())
	m.ShowChannels([]Channel{
		{Server: "oftc", Name: "#zeta"},
		{Server: "libera", Name: "#go"},
	})

	_, cmd := m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("channel selection did not return a command")
	}
	msg := cmd()
	got, ok := msg.(ChannelSelectionMsg)
	if !ok {
		t.Fatalf("channel selection returned %T, want ChannelSelectionMsg", msg)
	}
	want := ChannelSelectionMsg{Server: "libera", Channel: "#go"}
	if got != want {
		t.Fatalf("channel selection = %#v, want %#v", got, want)
	}
}

func TestChannelPickerUsesChannelSpecificPresentation(t *testing.T) {
	m := New(styles.RosePineTheme())
	m.SetSize(90, 12)
	m.ShowChannels([]Channel{{Server: "libera", Name: "#go"}})

	view := palettePlainText(m.View())
	if !strings.Contains(view, "Search channels...") {
		t.Fatalf("channel picker is missing its search placeholder:\n%s", view)
	}

	line := palettePlainText(m.renderChannelLine(0, action{Name: "#go", Description: "libera"}, channelPickerWidth))
	channelPos := strings.Index(line, "#go")
	serverPos := strings.Index(line, "libera")
	if got, want := serverPos-channelPos, channelNameWidth+channelServerGapWidth; got != want {
		t.Fatalf("server starts %d columns after channel, want %d: %q", got, want, line)
	}

	m.input.SetValue("missing")
	view = palettePlainText(m.View())
	if !strings.Contains(view, "No channels found") {
		t.Fatalf("channel picker has the wrong empty state:\n%s", view)
	}
}

func TestChannelPickerIsHeightBoundedAndFollowsCursor(t *testing.T) {
	m := New(styles.RosePineTheme())
	m.SetSize(64, 12)
	channels := make([]Channel, 20)
	for i := range channels {
		channels[i] = Channel{Server: "libera", Name: fmt.Sprintf("#channel-%02d", i)}
	}
	m.ShowChannels(channels)

	for range 10 {
		m.moveDown()
	}
	filtered := m.filteredCommands()
	start := visibleStart(m.cursor, len(filtered), m.visibleRowCount())
	line := m.renderChannelLine(start, filtered[start], channelPickerWidth)
	if got := lipgloss.Height(line); got != 1 {
		t.Fatalf("command line height = %d, width = %d", got, lipgloss.Width(line))
	}

	view := m.View()
	if got := lipgloss.Height(view); got > 12 {
		t.Fatalf("palette height = %d, want at most 12", got)
	}
	if !strings.Contains(view, "#channel-10") {
		t.Fatalf("palette did not scroll to the selected channel:\n%s", view)
	}
	if strings.Contains(view, "#channel-00") {
		t.Fatalf("palette still shows the first channel after scrolling:\n%s", view)
	}
	if !strings.Contains(view, "┃") || !strings.Contains(view, "│") {
		t.Fatalf("palette does not show a scrollbar for overflowing channels:\n%s", view)
	}
}

func TestScrollbarThumbTracksVisibleWindow(t *testing.T) {
	if got := lipgloss.Width("┃"); got != 1 {
		t.Fatalf("scrollbar rune width = %d, want 1", got)
	}
	tests := []struct {
		name      string
		start     int
		wantTop   int
		wantShown bool
	}{
		{name: "no overflow", start: 0, wantShown: false},
		{name: "top", start: 0, wantTop: 0, wantShown: true},
		{name: "middle", start: 7, wantTop: 2, wantShown: true},
		{name: "bottom", start: 14, wantTop: 4, wantShown: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			total := 20
			visible := 6
			if !tt.wantShown {
				total = visible
			}
			top, height, shown := scrollbarThumb(tt.start, total, visible)
			if shown != tt.wantShown {
				t.Fatalf("shown = %v, want %v", shown, tt.wantShown)
			}
			if !shown {
				return
			}
			if top != tt.wantTop {
				t.Fatalf("top = %d, want %d", top, tt.wantTop)
			}
			if height != 2 {
				t.Fatalf("height = %d, want 2", height)
			}
		})
	}
}

var paletteANSIPattern = regexp.MustCompile(`\x1b\[[0-9;]*m`)

func palettePlainText(s string) string {
	return paletteANSIPattern.ReplaceAllString(s, "")
}
