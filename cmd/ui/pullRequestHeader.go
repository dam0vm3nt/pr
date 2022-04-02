package ui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"strings"
)

type pullRequestHeader struct {
	data           *pullRequestData
	header         *lines
	currentHeading []int
	headings       [][]Heading
	width          int
	height         int
	maxReviews     int
	maxChecks      int
}

func (p pullRequestHeader) measureHeight() int {
	return min1(p.maxChecks, len(p.data.checks)) + min1(len(p.data.reviews), p.maxReviews) + 4
}

func (p pullRequestHeader) Init() tea.Cmd {
	return nil
}

func (p pullRequestHeader) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		p.width = msg.Width
		p.height = msg.Height
	}

	return p, nil
}

func (p pullRequestHeader) View() string {

	linesCh := make(chan string)
	go func() {
		for _, l := range *p.header {
			linesCh <- l
		}

		for l := 0; l < len(p.currentHeading); l++ {
			if p.currentHeading[l] >= 0 {
				linesCh <- p.headings[l][p.currentHeading[l]].text
			} else {
				linesCh <- ""
			}
		}

		close(linesCh)
	}()

	style := lipgloss.
		NewStyle().
		Width(p.width).Height(1).
		Background(lipgloss.Color("#fefefe")).Foreground(lipgloss.Color("#000000"))

	lines := make([]string, 0)
	for l := range linesCh {
		l = style.Render(l)
		l := strings.Split(l, "\n")[0]
		lines = append(lines, l)
	}

	return strings.Join(lines, "\n")
}
