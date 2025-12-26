package chat

import (
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/vaaleyard/dex/internal/ui/components/input"
	"github.com/vaaleyard/dex/internal/ui/layout"
	"github.com/vaaleyard/dex/internal/ui/styles"
	"strings"
)

type Message struct {
	Nick string
	Text string
	Time string
}
type Model struct {
	messages []string
	viewport viewport.Model
	input    input.Model

	theme styles.Theme
}

func New(theme styles.Theme) Model {
	m := Model{
		messages: []string{
			"[16:11] connecting to server...",
			"[16:11] connected",
			"[16:11] *** Checking Ident",
			"[16:11] *** Looking up your hostname...",
			"[16:11] *** Couldn't look up your hostname",
			"[16:11] *** No Ident response",
			"[16:11] Your host is lithium.libera.chat[0000:0000::0000:0000:0000:0000/6697], running version dex-1.0-dev",
			"[16:11] This server was created Tue Jul 16 2024 at 04:09:28 UTC",
			"[16:11] lithium.libera.chat dex-1.0-dev 0000000000000000000 000000000000000000000000000000 0000000000",
			"[16:11] There are 62 users and 31057 invisible on 28 servers",
			"[16:11] 42 IRC Operators online",
			"[16:11] 58 unknown connection(s)",
			"[16:11] 22434 channels formed",
			"[16:11] I have 2032 clients and 1 servers",
			"[16:11] 2032 2704 Current local users 2032, max 2704",
			"[16:11] 31119 34153 Current global users 31119, max 34153",
			"[16:11] Highest connection count: 2705 (2704 clients) (426530 connections received)",
			"[16:11] - lithium.libera.chat Message of the Day -",
			"[16:11] - Welcome to Libera Chat, the IRC network for",
			"[16:11] - free & open-source software and peer directed projects.",
			"[16:11] -",
			"[16:11] - Use of Libera Chat is governed by our network policies.",
			"[16:11] - To reduce network abuses we perform open proxy checks",
			"[16:11] - on hosts at connection time.",
			"[16:11] - Please visit us in #libera for questions and support.",
			"[16:11] - Website and documentation:  https://libera.chat/",
			"[16:11] - Webchat:                    https://web.libera.chat/",
			"[16:11] - Network policies:           https://libera.chat/policies",
			"[16:11] - Email:                      support@libera.chat",
			"[16:11] End of /MOTD command.",
		},
		theme:    theme,
		input:    input.New(theme),
		viewport: viewport.New(0, 0),
	}
	m.viewport.SetContent(strings.Join(m.messages, "\n"))

	return m
}

func (m Model) Init() tea.Cmd {
	return m.input.Init()
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	cmd := m.input.Update(msg)
	return m, cmd
}

func (m Model) View(width, height int) string {
	m.viewport.Width = width
	m.viewport.Height = height - layout.InputBoxHeight

	inputView := m.input.View(width)

	combinedView := lipgloss.JoinVertical(
		lipgloss.Left,
		m.viewport.View(),
		inputView,
	)

	return m.theme.Styles.ChatArea.
		Render(combinedView)
}
