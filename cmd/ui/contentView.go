package ui

import "C"
import (
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/pterm/pterm"
	"strings"
)

type contentView struct {
	data         *pullRequestData
	viewport     viewport.Model
	content      *lines
	selectedLine int
	oldLine      string
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

func (c *contentView) isLineSelected() bool {
	return c.selectedLine != -1
}

func (c *contentView) selectUp() {
	if c.selectedLine > 0 {
		c.selectLine(c.selectedLine - 1)

		for c.selectedLine < c.viewport.YOffset {
			c.viewport.YOffset -= 1
		}
	}
}

func (c *contentView) selectDown() {
	if c.selectedLine != -1 && c.selectedLine+1 < len(*c.content) {
		c.selectLine(c.selectedLine + 1)
		for c.selectedLine-c.viewport.YOffset >= c.viewport.Height {
			c.viewport.YOffset += 1
		}
	}
}

func (c *contentView) selectLine(l int) {

	if c.selectedLine != -1 {
		(*c.content)[c.selectedLine] = c.oldLine
		c.oldLine = ""
		c.selectedLine = -1
	}

	if l != -1 {
		c.oldLine = (*c.content)[l]
		newLine := pterm.RemoveColorFromString(c.oldLine)
		newLine = lipgloss.NewStyle().
			Width(c.viewport.Width).
			MaxHeight(1).
			Background(lipgloss.Color("#a00000")).
			Foreground(lipgloss.Color("#ffffff")).
			Bold(true).
			Italic(true).
			Render("➡    " + newLine[5:])
		(*c.content)[l] = newLine
		c.selectedLine = l
	}

	c.updateViewportWithContent()
}

func (c contentView) View() string {
	return c.viewport.View()
}

func (c *contentView) updateViewportWithContent() {
	c.viewport.SetContent(strings.Join(*c.content, "\n"))
}
