package ui

import (
	"fmt"
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

type fileSelectedMsg struct {
	ordinal int
}

func fileSelected(file int) tea.Cmd {
	return func() tea.Msg {
		return fileSelectedMsg{file}
	}
}

func (f fileList) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd = nil
	switch m := msg.(type) {
	case tea.WindowSizeMsg:
		f.w = m.Width
		f.h = m.Height
	case tea.KeyMsg:
		switch m.String() {
		case "down":
			if f.selectedLine < len(f.pullRequestData.files)-1 {
				f.selectedLine += 1
				cmd = fileSelected(f.selectedLine)
			}
			for (f.selectedLine - f.firstLine) >= f.h {
				f.firstLine++
			}
		case "up":
			if f.selectedLine > 0 {
				f.selectedLine -= 1
				cmd = fileSelected(f.selectedLine)
			}
			for f.selectedLine < f.firstLine {
				f.firstLine--
			}
		}
	case fileSelectedMsg:
		if f.selectedLine != m.ordinal {
			f.selectedLine = m.ordinal
			f.firstLine = min1(-min1(-(f.firstLine-f.h), -f.firstLine), f.selectedLine)
		}
	case focusChangedMsg:
		f.active = m.newFocus == FILEVIEW_ADDRESS
	}
	return f, cmd
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
	for i, file := range f.pullRequestData.files[f.firstLine:] {
		var ss lipgloss.Style
		if i+f.firstLine == f.selectedLine {
			ss = sel
		} else {
			ss = s
		}
		fn := getFileName(file)
		fn = crop(fn, f.w)
		l = append(l, ss.Render(fillLine(fn, f.w)))
		if len(l) >= f.h {
			break
		}
	}
	return strings.Join(l, "\n")
}

func crop(fn string, w int) string {
	if len(fn) > w {
		x := strings.Index(fn, "/") + 1
		if x <= 0 {
			x = w / 4
		} else {
			x = min1(w/4, x)
		}
		s := len(fn) - (w - x - 3)
		return fmt.Sprintf("%s...%s", fn[0:x], fn[s:])
	}
	return fn
}
