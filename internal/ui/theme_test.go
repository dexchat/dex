package ui

import (
	"fmt"
	"image/color"
	"strings"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/dexchat/dex/internal/ui/components/chat"
	"github.com/dexchat/dex/internal/ui/styles"
)

func TestSetThemeRestylesActiveAndInactiveBuffers(t *testing.T) {
	m := newActivityTestModel()
	var cmds []tea.Cmd
	m.selectBuffer("libera", "#go", &cmds)
	m.Update(tea.WindowSizeMsg{Width: 100, Height: 20})

	active := m.buffers[makeBufferKey("libera", "#go")]
	inactive := m.buffers[makeBufferKey("libera", "#random")]
	active.Chat.AddMessage(chat.Message{Timestamp: time.Now(), Username: "alice", Text: "visible"})
	inactive.Chat.AddMessage(chat.Message{Timestamp: time.Now(), Username: "bob", Text: "hidden"})

	oldBackground := backgroundSequence(m.theme.Colors.Base.Background)
	light := styles.SolarizedLightTheme()
	newBackground := backgroundSequence(light.Colors.Base.Background)

	m.setTheme(light)

	if view := active.Chat.View(); strings.Contains(view, oldBackground) || !strings.Contains(view, newBackground) {
		t.Fatalf("active chat was not re-rendered with the new theme:\n%q", view)
	}
	if view := m.channels.View(channelsPanelMaxWidth, 20); strings.Contains(view, oldBackground) {
		t.Fatalf("channels sidebar kept the old background:\n%q", view)
	}

	m.selectBuffer("libera", "#random", &cmds)

	if view := inactive.Chat.View(); strings.Contains(view, oldBackground) || !strings.Contains(view, newBackground) {
		t.Fatalf("inactive chat was not re-rendered on selection:\n%q", view)
	}
}

func TestSetThemeBeforeWindowSizeAppliesOnFirstLayout(t *testing.T) {
	m := New(newActivityTestModel().config)
	light := styles.SolarizedLightTheme()

	m.setTheme(light)
	m.Update(tea.WindowSizeMsg{Width: 100, Height: 20})

	newBackground := backgroundSequence(light.Colors.Base.Background)
	if view := m.channels.View(channelsPanelMaxWidth, 20); !strings.Contains(view, newBackground) {
		t.Fatalf("channels sidebar did not use the theme set before layout:\n%q", view)
	}
}

// backgroundSequence returns the SGR parameters that set c as a truecolor
// background.
func backgroundSequence(c color.Color) string {
	r, g, b, _ := c.RGBA()
	return fmt.Sprintf("48;2;%d;%d;%d", r>>8, g>>8, b>>8)
}
