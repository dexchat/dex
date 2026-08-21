package palette

import (
	"fmt"
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
	line := m.renderCommandLine(start, filtered[start], channelPickerWidth, scrollbarWidth)
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
