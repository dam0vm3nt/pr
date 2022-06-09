package ui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"strings"
)

type fileList struct {
	pullRequestData *pullRequestData
	w               int
	h               int
}

func (f fileList) Init() tea.Cmd {
	return nil
}

func (f fileList) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch m := msg.(type) {
	case tea.WindowSizeMsg:
		f.w = m.Width
		f.h = m.Height
	}
	return f, nil
}

func (f fileList) View() string {
	s := lipgloss.NewStyle().Width(f.w).Inline(true)
	l := make([]string, 0)
	for _, fn := range f.pullRequestData.files {
		l = append(l, s.Render(fillLine(fn.NewName, f.w)))
		if len(l) >= f.h {
			break
		}
	}
	return strings.Join(l, "\n")
}
