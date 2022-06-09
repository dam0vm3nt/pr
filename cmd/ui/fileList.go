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
	active          bool
	selectedLine    int
	firstLine       int
}

func (f fileList) Init() tea.Cmd {
	return nil
}

func (f fileList) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch m := msg.(type) {
	case tea.WindowSizeMsg:
		f.w = m.Width
		f.h = m.Height
	case tea.KeyMsg:
		switch m.String() {
		case "down":
			if f.selectedLine < len(f.pullRequestData.files)-1 {
				f.selectedLine += 1
			}
			for (f.selectedLine - f.firstLine) >= f.h {
				f.firstLine++
			}
		case "up":
			if f.selectedLine > 0 {
				f.selectedLine -= 1
			}
			for f.selectedLine < f.firstLine {
				f.firstLine--
			}
		}
	case focusChangedMsg:
		f.active = m.newFocus == FILEVIEW_ADDRESS
	}
	return f, nil
}

func (f fileList) View() string {
	s := lipgloss.NewStyle().Width(f.w).Inline(true)
	var sel lipgloss.Style
	if f.active {
		sel = lipgloss.NewStyle().Width(f.w).Background(lipgloss.Color("#ffffff")).
			Foreground(lipgloss.Color("#000000")).Inline(true)
	} else {
		sel = lipgloss.NewStyle().Width(f.w).Background(lipgloss.Color("#e0e0e0")).
			Foreground(lipgloss.Color("#000000")).Inline(true)
	}
	l := make([]string, 0)
	for i, fn := range f.pullRequestData.files[f.firstLine:] {
		var ss lipgloss.Style
		if i+f.firstLine == f.selectedLine {
			ss = sel
		} else {
			ss = s
		}
		l = append(l, ss.Render(fillLine(fn.NewName, f.w)))
		if len(l) >= f.h {
			break
		}
	}
	return strings.Join(l, "\n")
}
