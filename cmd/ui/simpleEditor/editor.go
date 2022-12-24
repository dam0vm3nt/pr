package simpleEditor

import (
	"errors"
	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"strings"
)

type TaModel struct {
	title    string
	textArea textarea.Model
}

func (t TaModel) Init() tea.Cmd {
	return textarea.Blink
}

func (t TaModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	cmds := make([]tea.Cmd, 0)
	taModel, taCmd := t.textArea.Update(msg)

	t.textArea = taModel
	cmds = append(cmds, taCmd)

	switch m := msg.(type) {
	case tea.KeyMsg:
		switch k := m.String(); k {
		case "esc":
			cmds = append(cmds, tea.Quit)
		}
	}

	return t, tea.Batch(cmds...)
}

func (t TaModel) View() string {
	titleStle := lipgloss.NewStyle().Background(lipgloss.Color("#ffffff")).Foreground(lipgloss.Color("#000000")).Bold(true).Width(t.textArea.Width()).ColorWhitespace(true)
	lines := make([]string, 0)
	lines = append(lines, titleStle.Render(t.title))
	lines = append(lines, t.textArea.View())
	return strings.Join(lines, "\n")
}

type Opts interface {
	apply(model TaModel) TaModel
}

type WithWidth struct {
	Width int
}

func (o WithWidth) apply(model TaModel) TaModel {
	model.textArea.SetWidth(o.Width)
	return model
}

type WithTitle struct {
	Title string
}

func (o WithTitle) apply(model TaModel) TaModel {
	model.title = o.Title
	return model
}

type WithValue struct {
	Value string
}

func (w WithValue) apply(model TaModel) TaModel {
	model.textArea.SetValue(w.Value)
	return model
}

type WithPlaceholder struct {
	Value string
}

func (w WithPlaceholder) apply(model TaModel) TaModel {
	model.textArea.Placeholder = w.Value
	return model
}

func RunEditor(opts ...Opts) (string, error) {
	model := TaModel{title: "", textArea: textarea.New()}
	model.textArea.Prompt = "| "
	model.textArea.Focus()

	for _, o := range opts {
		model = o.apply(model)
	}

	prg := tea.NewProgram(model)
	if m, err := prg.StartReturningModel(); err != nil {
		return "", err
	} else if model, ok := m.(TaModel); ok {
		return model.textArea.Value(), nil
	} else {
		return "", errors.New("that's unusual")
	}
}
