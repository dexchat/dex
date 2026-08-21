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
	m.SetSize(90, 12)
	channels := make([]Channel, 20)
	for i := range channels {
		channels[i] = Channel{Server: "libera", Name: fmt.Sprintf("#channel-%02d", i)}
	}
	m.ShowChannels(channels)

	for range 10 {
		m.moveDown()
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
}
