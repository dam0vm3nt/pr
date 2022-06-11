package ui

import "C"
import (
	"github.com/bluekeyes/go-gitdiff/gitdiff"
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
	codeLines    map[int]codeLine
}

func (cv *contentView) printf(line string, args ...any) {
	cv.content.printf(line, args...)
}

func (cv *contentView) clear() {
	c := make(lines, 0)
	cv.content = &c
	cv.codeLines = make(map[int]codeLine)
}

func (c contentView) Init() tea.Cmd {
	return nil
}

type lineCmd int

const (
	newComment lineCmd = iota
	editComment
	deleteComment
)

type lineCommandMsg struct {
	cmd  lineCmd
	line int
	code *codeLine
}

func lineCommand(cmd lineCmd, line int, code *codeLine) func() tea.Msg {
	return func() tea.Msg {
		return lineCommandMsg{cmd, line, code}
	}
}

func (content contentView) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		content.viewport.Width = msg.Width
		content.viewport.Height = msg.Height
		return content, renderPrCmd
	case tea.KeyMsg:
		if !content.isLineSelected() {
			if msg.String() == " " {
				content.selectLine(content.viewport.YOffset)
			} else {
				newViewport, cmd := content.viewport.Update(msg)
				content.viewport = newViewport
				return content, cmd
			}
		} else {
			switch msg.String() {
			case "up":
				content.selectUp()
			case "down":
				content.selectDown()
			case " ":
				content.selectLine(-1)
			case "+":
				if ln, ok := content.codeLines[content.selectedLine]; ok {
					cmd := lineCommand(newComment, content.selectedLine, &ln)
					content.selectLine(-1)
					return content, cmd
				}

			}
		}
	}
	return content, nil
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

type codeLine struct {
	old      int64
	new      int64
	position int
	file     *gitdiff.File
	code     gitdiff.Line
	commitId string
}

func (c *contentView) saveLine(commitId string, old int64, new int64, pos int, path *gitdiff.File, ln gitdiff.Line) {
	c.codeLines[len(*c.content)] = codeLine{old, new, pos, path, ln, commitId}
}
