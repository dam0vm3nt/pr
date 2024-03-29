package ui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/pterm/pterm"
	"time"
)

type messageMode int

const (
	normalMode messageMode = iota
	severeMode
)

type statusBar struct {
	width     int
	mode      messageMode
	message   string
	messageId int64
}

func newStatusBar() statusBar {
	return statusBar{-1, normalMode, "", 0}
}

func (s statusBar) Init() tea.Cmd {
	return nil
}

type showStatusMsg struct {
	mode    messageMode
	message string
	timeout time.Duration
}

type clearStatusMsg struct {
	messageId int64
}

func clearStatus(messageId int64) tea.Cmd {
	return func() tea.Msg {
		return clearStatusMsg{messageId: messageId}
	}
}

func showStatusCmd(mode messageMode, message string, timeout time.Duration) tea.Cmd {
	return func() tea.Msg {
		return showStatusMsg{mode, message, timeout}
	}
}

func showErrCmd(err error) tea.Cmd {
	return showStatusCmd(severeMode, err.Error(), 3*time.Second)
}

func (s statusBar) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		s.width = msg.Width
	case clearStatusMsg:
		if s.messageId == msg.messageId {
			s.message = ""
			s.mode = normalMode
			s.messageId++
		}
	case showStatusMsg:
		s.message = msg.message
		s.mode = msg.mode
		s.messageId++

		if msg.timeout > 0 {
			go func(id int64) {
				time.Sleep(msg.timeout)
				sendAsyncCmd(clearStatus(id))
			}(s.messageId)
		}
	}

	return s, nil
}

func (s statusBar) View() string {
	if s.width < 0 {
		return ""
	}
	m := pterm.RemoveColorFromString(s.message)
	if len(m) > s.width {
		m = m[0:s.width]
	}
	for len(m) < s.width {
		m += " "
	}
	style := func(mode messageMode) lipgloss.Style {
		switch mode {
		case severeMode:
			return lipgloss.NewStyle().
				Background(lipgloss.Color("#ff0000")).
				Foreground(lipgloss.Color("#ffffff")).
				Width(s.width).
				MaxHeight(1)
		default:
			return lipgloss.NewStyle().
				Background(lipgloss.Color("#000000")).
				Foreground(lipgloss.Color("#e0e0e0")).
				Width(s.width).
				MaxHeight(1)

		}
	}(s.mode)

	return style.Render(s.message)
}
