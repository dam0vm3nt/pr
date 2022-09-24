package ui

import (
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/pterm/pterm"
	"github.com/vballestra/sv/sv"
	"strings"
)

type PrStatusView struct {
	pullRequests  []sv.PullRequestStatus
	currentSelect int
	w             int
	h             int

	ready bool
}

func (p PrStatusView) Init() tea.Cmd {
	return nil
}

func (p PrStatusView) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch m := msg.(type) {
	case tea.WindowSizeMsg:
		// save window info
		p.w = m.Width
		p.h = m.Height
		p.ready = true
		break
	case tea.KeyMsg:
		return p.handleKey(m)
	}

	return p, nil
}

func (p PrStatusView) handleKey(m tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch t := m.Type.String(); t {
	case "up":
		if p.currentSelect > 0 {
			p.currentSelect = p.currentSelect - 1
		}
	case "down":
		if p.currentSelect < len(p.pullRequests)-1 {
			p.currentSelect = p.currentSelect + 1
		}
	case "ctrl+c":
		return p, tea.Quit
	case "esc":
		return p, tea.Quit
	case "runes":
		switch r := m.Runes[0]; r {
		case 'q':
			return p, tea.Quit
		}
	}

	return p, nil
}

func (p PrStatusView) View() string {
	if !p.ready {
		return "... initializing ..."
	}

	data := pterm.TableData{{"ID", "Title", "Author", "Repository", "Branch", "State", "Reviews", "Checks", "Contexts"}}
	mineStyle := lipgloss.NewStyle().Bold(true).ColorWhitespace(true).Foreground(lipgloss.Color("#ff0000"))
	theirStyle := lipgloss.NewStyle().Bold(true).ColorWhitespace(true).Foreground(lipgloss.Color("#00ffff"))

	for i, pr := range p.pullRequests {
		selected := i == p.currentSelect
		checks := make([]string, 0)
		for s, k := range pr.GetChecksByStatus() {
			if k > 0 {
				checks = append(checks, fmt.Sprintf("%s: %d", s, k))
			}
		}

		contexts := make([]string, 0)
		for s, k := range pr.GetContextByStatus() {
			if k > 0 {
				contexts = append(contexts, fmt.Sprintf("%s: %d", s, k))
			}
		}

		reviewCount := make(map[string]map[string]int)
		for _, r := range pr.GetReviews() {
			if byStatus, ok := reviewCount[r.GetState()]; ok {
				if count, ok := byStatus[r.GetAuthor()]; ok {
					byStatus[r.GetAuthor()] = count + 1
				} else {
					byStatus[r.GetAuthor()] = 1
				}
				reviewCount[r.GetState()] = byStatus
			} else {
				byStatus = make(map[string]int)
				byStatus[r.GetAuthor()] = 1
				reviewCount[r.GetState()] = byStatus
			}
		}

		reviews := make([]string, 0)
		for s, byStatus := range reviewCount {
			var stats string
			stats = fmt.Sprintf("%d", len(byStatus))

			reviews = append(reviews, fmt.Sprintf("%s: %s", s, stats))
		}

		var idStr string

		if selected {
			idStr = fmt.Sprintf("* %5d", pr.GetId())
		} else {
			idStr = fmt.Sprintf("  %5d", pr.GetId())
		}

		if pr.IsMine() {
			idStr = mineStyle.Render(idStr)
		} else {
			idStr = theirStyle.Render(idStr)
		}

		data = append(data, []string{
			idStr, pr.GetTitle(), pr.GetAuthor(), pr.GetRepository(), pr.GetBranchName(), pr.GetStatus(), strings.Join(reviews, ", "), strings.Join(checks, ", "),
			strings.Join(contexts, ", "),
		})
	}

	tbl, rerr := pterm.DefaultTable.WithHasHeader().WithData(data).Srender()
	if rerr != nil {
		pterm.Fatal.Println(rerr)
	}

	return tbl
}

func RunPrStatusView(s sv.Sv) error {
	if ch, err := s.PullRequestStatus(); err != nil {
		return err
	} else {

		pullRequests := make([]sv.PullRequestStatus, 0)

		for p := range ch {
			pullRequests = append(pullRequests, p)
		}

		view := PrStatusView{
			pullRequests:  pullRequests,
			currentSelect: 0,
			w:             0,
			h:             0,
			ready:         false,
		}

		prg := tea.NewProgram(view)
		if err := prg.Start(); err != nil {
			return err
		}

		return nil
	}
}
