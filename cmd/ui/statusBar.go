package ui

import (
	"fmt"
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
	width   int
	mode    messageMode
	message string
}

func newStatusBar() statusBar {
	return statusBar{-1, normalMode, ""}
}

func (s statusBar) Init() tea.Cmd {
	//TODO implement me
	panic("implement me")
}

type showStatusMsg struct {
	mode    messageMode
	message string
}

func (s statusBar) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		s.width = msg.Width
	case showErrMsg:
		s.message = fmt.Sprintf("%e", msg.err)
		s.mode = severeMode
		go func() {
			time.Sleep(2 * time.Second)
			sendAsyncMsg(showStatusMsg{normalMode, ""})
		}()
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
				Background(lipgloss.Color("red")).
				Foreground(lipgloss.Color("white")).
				MaxHeight(1)
		default:
			return lipgloss.NewStyle().
				Background(lipgloss.Color("black")).
				Foreground(lipgloss.Color("lightgray")).
				MaxHeight(1)

		}
	}(s.mode)

	return style.Render(s.message)
}
