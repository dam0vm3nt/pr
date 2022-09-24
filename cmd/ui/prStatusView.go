package ui

import (
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/evertras/bubble-table/table"
	"github.com/vballestra/sv/sv"
	"strings"
)

type PrStatusView struct {
	pullRequests []sv.PullRequestStatus
	w            int
	h            int

	ready bool

	statusTable table.Model
}

func (p PrStatusView) Init() tea.Cmd {
	return nil
}

type setupTableMsg struct {
}

const (
	colId         = "id"
	colTitle      = "title"
	colAuthor     = "author"
	colRepository = "repo"
	colBranch     = "branch"
	colState      = "state"
	colReviews    = "reviews"
	colChecks     = "checks"
	colContexts   = "contexts"
)

var mineStyle = lipgloss.NewStyle().Bold(true).ColorWhitespace(true).Foreground(lipgloss.Color("#ff0000"))
var theirStyle = lipgloss.NewStyle().Bold(true).ColorWhitespace(true).Foreground(lipgloss.Color("#00ffff"))

func (m setupTableMsg) Update(p PrStatusView) (PrStatusView, tea.Cmd) {
	rows := make([]table.Row, 0)
	for _, pi := range p.pullRequests {

		checks := make([]string, 0)
		for s, k := range pi.GetChecksByStatus() {
			if k > 0 {
				checks = append(checks, fmt.Sprintf("%s: %d", s, k))
			}
		}

		contexts := make([]string, 0)
		for s, k := range pi.GetContextByStatus() {
			if k > 0 {
				contexts = append(contexts, fmt.Sprintf("%s: %d", s, k))
			}
		}

		reviewCount := make(map[string]map[string]int)
		for _, r := range pi.GetReviews() {
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

		if pi.IsMine() {
			idStr = mineStyle.Render(idStr)
		} else {
			idStr = theirStyle.Render(idStr)
		}

		row := table.NewRow(table.RowData{
			colId:         pi.GetId(),
			colTitle:      pi.GetTitle(),
			colAuthor:     pi.GetAuthor(),
			colRepository: pi.GetRepository(),
			colBranch:     pi.GetBranchName(),
			colState:      pi.GetStatus(),
			colReviews:    strings.Join(reviews, ", "),
			colChecks:     strings.Join(checks, ", "),
			colContexts:   strings.Join(contexts, ", "),
		})
		rows = append(rows, row)
	}
	p.statusTable = table.New([]table.Column{
		table.NewColumn(colId, "ID", 5).
			WithFormatString("%05d"),
		table.NewFlexColumn(colTitle, "Title", 3),
		table.NewColumn(colAuthor, "Author", 10),
		table.NewFlexColumn(colBranch, "Branch", 1),
		table.NewFlexColumn(colRepository, "Repository", 1),
		table.NewColumn(colState, "State", 10),
		table.NewFlexColumn(colReviews, "Reviews", 2),
		table.NewFlexColumn(colChecks, "Checks", 2),
		table.NewFlexColumn(colContexts, "Contexts", 2),
	}).WithRows(rows).
		WithTargetWidth(p.w).
		BorderRounded().
		WithKeyMap(table.DefaultKeyMap()).
		Focused(true).
		WithHighlightedRow(0)
	p.ready = true

	return p, tea.ClearScrollArea
}

func setupTable() tea.Msg {
	return setupTableMsg{}
}

func (p PrStatusView) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	cmds := make([]tea.Cmd, 0)
	var pp tea.Model = p
	switch m := msg.(type) {
	case tea.WindowSizeMsg:
		// save window info
		p.w = m.Width
		p.h = m.Height
		pp = p
		cmds = append(cmds, setupTable)
		break
	case tea.KeyMsg:
		p_, cmd := p.handleKey(m)
		pp = p_
		cmds = append(cmds, cmd)
	case setupTableMsg:
		p_, cmd := m.Update(p)
		pp = p_
		cmds = append(cmds, cmd)
	}

	return pp, tea.Batch(cmds...)
}

func (p PrStatusView) handleKey(m tea.KeyMsg) (tea.Model, tea.Cmd) {
	pp := p
	cmds := make([]tea.Cmd, 0)

	switch t := m.Type.String(); t {
	case "ctrl+c":
		cmds = append(cmds, tea.Quit)
	case "esc":
		cmds = append(cmds, tea.Quit)
	case "runes":
		switch r := m.Runes[0]; r {
		case 'q':
			cmds = append(cmds, tea.Quit)
		}
	}

	// delegate table
	statusTable_, cmd := pp.statusTable.Update(m)
	cmds = append(cmds, cmd)
	pp.statusTable = statusTable_

	return pp, tea.Batch(cmds...)
}

func (p PrStatusView) View() string {
	if !p.ready {
		return "... initializing ..."
	}

	return p.statusTable.View()
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
			pullRequests: pullRequests,
			w:            0,
			h:            0,
			ready:        false,
		}

		prg := tea.NewProgram(view)
		if err := prg.Start(); err != nil {
			return err
		}

		return nil
	}
}
