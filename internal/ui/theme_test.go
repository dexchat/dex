package ui

import (
	"fmt"
	"image/color"
	"strings"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	uv "github.com/charmbracelet/ultraviolet"
	"github.com/dexchat/dex/internal/config"
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

func newSystemThemeTestModel() *Model {
	cfg := newActivityTestModel().config
	cfg.UI.SystemTheme = &config.SystemTheme{Light: config.ThemeSolarizedLight, Dark: config.ThemeDracula}
	return New(cfg)
}

func TestSystemThemeStartsDarkWithoutTerminalBackground(t *testing.T) {
	m := newSystemThemeTestModel()

	if m.systemTheme != systemThemeDetecting {
		t.Fatal("expected a light/dark theme to start detecting")
	}
	if got, want := m.theme.Colors.Base.Background, styles.DraculaTheme().Colors.Base.Background; got != want {
		t.Fatalf("initial background = %v, want dark theme %v", got, want)
	}
	if got := m.View().BackgroundColor; got != nil {
		t.Fatalf("view background while detecting = %v, want unset", got)
	}
}

func TestSystemThemeAppliesFirstDetectedAppearance(t *testing.T) {
	light := styles.SolarizedLightTheme().Colors.Base.Background
	dark := styles.DraculaTheme().Colors.Base.Background
	tests := []struct {
		name string
		msgs []tea.Msg
		want color.Color
	}{
		{name: "light report", msgs: []tea.Msg{uv.LightColorSchemeEvent{}}, want: light},
		{name: "dark report", msgs: []tea.Msg{uv.DarkColorSchemeEvent{}}, want: dark},
		{name: "light background", msgs: []tea.Msg{tea.BackgroundColorMsg{Color: lipgloss.Color("#fdf6e3")}}, want: light},
		{name: "dark background", msgs: []tea.Msg{tea.BackgroundColorMsg{Color: lipgloss.Color("#1e1e2e")}}, want: dark},
		{
			name: "report before background",
			msgs: []tea.Msg{uv.LightColorSchemeEvent{}, tea.BackgroundColorMsg{Color: lipgloss.Color("#1e1e2e")}},
			want: light,
		},
		{name: "timeout keeps dark", msgs: []tea.Msg{themeDetectTimeoutMsg{}}, want: dark},
		{
			name: "background after timeout",
			msgs: []tea.Msg{themeDetectTimeoutMsg{}, tea.BackgroundColorMsg{Color: lipgloss.Color("#fdf6e3")}},
			want: dark,
		},
		{name: "report after timeout", msgs: []tea.Msg{themeDetectTimeoutMsg{}, uv.LightColorSchemeEvent{}}, want: light},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := newSystemThemeTestModel()
			m.Update(tea.WindowSizeMsg{Width: 100, Height: 20})

			for _, msg := range tt.msgs {
				m.Update(msg)
			}

			if m.systemTheme == systemThemeDetecting {
				t.Fatal("expected detection to finish")
			}
			if got := m.theme.Colors.Base.Background; got != tt.want {
				t.Fatalf("background = %v, want %v", got, tt.want)
			}
			if got := m.View().BackgroundColor; got != tt.want {
				t.Fatalf("view background = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSystemThemeFollowsSystemAppearanceChanges(t *testing.T) {
	m := newSystemThemeTestModel()
	m.Update(tea.WindowSizeMsg{Width: 100, Height: 20})
	m.Update(uv.DarkColorSchemeEvent{})
	light := styles.SolarizedLightTheme().Colors.Base.Background
	dark := styles.DraculaTheme().Colors.Base.Background

	steps := []struct {
		msg  tea.Msg
		want color.Color
	}{
		{msg: uv.LightColorSchemeEvent{}, want: light},
		{msg: tea.BackgroundColorMsg{Color: lipgloss.Color("#1e1e2e")}, want: light},
		{msg: uv.LightColorSchemeEvent{}, want: light},
		{msg: uv.DarkColorSchemeEvent{}, want: dark},
	}
	for i, step := range steps {
		m.Update(step.msg)
		if got := m.theme.Colors.Base.Background; got != step.want {
			t.Fatalf("step %d (%T): background = %v, want %v", i, step.msg, got, step.want)
		}
		if got := m.View().BackgroundColor; got != step.want {
			t.Fatalf("step %d (%T): view background = %v, want %v", i, step.msg, got, step.want)
		}
	}
}

func TestFixedThemeIgnoresDetectedAppearance(t *testing.T) {
	m := newActivityTestModel()
	before := m.theme.Colors.Base.Background

	if cmd := m.detectTheme(); cmd != nil {
		t.Fatal("expected no detection for a fixed theme")
	}
	m.Update(uv.LightColorSchemeEvent{})
	m.Update(tea.BackgroundColorMsg{Color: lipgloss.Color("#fdf6e3")})

	if got := m.theme.Colors.Base.Background; got != before {
		t.Fatalf("background = %v, want unchanged %v", got, before)
	}
}

// backgroundSequence returns the SGR parameters that set c as a truecolor
// background.
func backgroundSequence(c color.Color) string {
	r, g, b, _ := c.RGBA()
	return fmt.Sprintf("48;2;%d;%d;%d", r>>8, g>>8, b>>8)
}
