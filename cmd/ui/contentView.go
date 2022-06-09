package ui

import (
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"strings"
)

type contentView struct {
	data     *pullRequestData
	viewport viewport.Model
	content  *lines
}

func (cv *contentView) printf(line string, args ...any) {
	cv.content.printf(line, args...)
}

func (cv *contentView) clear() {
	c := make(lines, 0)
	cv.content = &c
}

func (c contentView) Init() tea.Cmd {
	return nil
}

func (c contentView) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		c.viewport.Width = msg.Width
		c.viewport.Height = msg.Height
		return c, renderPrCmd
	}
	return c, nil
}

func (c contentView) View() string {
	return c.viewport.View()
}

func (c *contentView) updateViewportWithContent() {
	c.viewport.SetContent(strings.Join(*c.content, "\n"))
}
