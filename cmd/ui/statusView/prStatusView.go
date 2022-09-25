package statusView

import (
	"fmt"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/evertras/bubble-table/table"
	"github.com/pterm/pterm"
	"github.com/vballestra/sv/cmd/ui"
	"github.com/vballestra/sv/sv"
	"strings"
)

type PrStatusView struct {
	sv           sv.Sv
	pullRequests []sv.PullRequestStatus
	w            int
	h            int

	ready  bool
	loaded bool

	statusTable   table.Model
	asyncMsg      chan tea.Cmd
	loadingStatus spinner.Model
}

func (p PrStatusView) Init() tea.Cmd {
	return loadPrStatusCmd
}

type setupTableMsg struct {
}

type showStatusError struct {
	err string
}

func (m showStatusError) Update(view PrStatusView) (PrStatusView, tea.Cmd) {
	view.statusTable = view.statusTable.WithStaticFooter(m.err)
	return view, nil
}

func showStatusErrorCmd(err string) tea.Cmd {
	return func() tea.Msg {
		return showStatusError{err}
	}
}

type loadPrStatusMsg struct {
}

func loadPrStatusCmd() tea.Msg {
	return loadPrStatusMsg{}
}

type finishedLoadingMsg struct {
	pullRequests []sv.PullRequestStatus
}

func (m finishedLoadingMsg) Update(p PrStatusView) (tea.Model, tea.Cmd) {
	p.loaded = true
	p.pullRequests = m.pullRequests
	return p, setupTable
}

func finishedLoadingCmd(pullRequests []sv.PullRequestStatus) tea.Cmd {
	return func() tea.Msg { return finishedLoadingMsg{pullRequests} }
}

func (m loadPrStatusMsg) Update(view PrStatusView) (tea.Model, tea.Cmd) {
	view.loadingStatus = spinner.New(spinner.WithSpinner(spinner.Points))
	view.loaded = false
	cmd := view.loadingStatus.Tick
	go func(tick tea.Cmd) {
		if ch, err := view.sv.PullRequestStatus(); err == nil {
			pullRequests := make([]sv.PullRequestStatus, 0)

			for p := range ch {
				pullRequests = append(pullRequests, p)
				view.asyncMsg <- tick
			}

			view.asyncMsg <- finishedLoadingCmd(pullRequests)
		} else {
			pterm.Fatal.Println(err)
		}

	}(cmd)

	return view, cmd
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

		checks := renderChecks(pi)

		contexts := renderContexts(pi)

		reviews := renderReviews(pi)

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
			colReviews:    strings.Join(reviews, " "),
			colChecks:     strings.Join(checks, ", "),
			colContexts:   strings.Join(contexts, " "),
		})
		rows = append(rows, row)
	}
	p.statusTable = table.New([]table.Column{
		table.NewColumn(colId, "ID", 5).
			WithFormatString("%05d"),
		table.NewFlexColumn(colTitle, "Title", 2).
			WithStyle(lipgloss.NewStyle().
				Align(lipgloss.Left)),
		table.NewColumn(colAuthor, "Author", 10).
			WithStyle(lipgloss.NewStyle().
				Align(lipgloss.Center)),
		table.NewFlexColumn(colBranch, "Branch", 2).
			WithStyle(lipgloss.NewStyle().
				Align(lipgloss.Left)),
		table.NewFlexColumn(colRepository, "Repository", 2).
			WithStyle(lipgloss.NewStyle().
				Align(lipgloss.Left)),
		table.NewColumn(colState, "State", 10).
			WithStyle(lipgloss.NewStyle().
				Align(lipgloss.Center)),
		table.NewFlexColumn(colReviews, "Reviews", 1).
			WithStyle(lipgloss.NewStyle().
				Align(lipgloss.Left)),
		table.NewFlexColumn(colChecks, "Checks", 1).
			WithStyle(lipgloss.NewStyle().
				Align(lipgloss.Center)),
		table.NewFlexColumn(colContexts, "Contexts", 1).
			WithStyle(lipgloss.NewStyle().
				Align(lipgloss.Center)),
	}).WithRows(rows).
		WithTargetWidth(p.w).
		BorderRounded().
		WithKeyMap(table.DefaultKeyMap()).
		Focused(true).
		WithHighlightedRow(0).
		WithPageSize(10).
		WithFooterVisibility(true)
	p.ready = true

	return p, tea.ClearScrollArea
}

func renderChecks(pi sv.PullRequestStatus) []string {
	changesMap := map[string]string{
		"SUCCESS": "👌",
		"FAILED":  "👎",
	}

	stylesMap := map[string]lipgloss.Style{
		"SUCCESS": lipgloss.NewStyle().Foreground(lipgloss.Color("#00ff00")),
		"FAILED":  lipgloss.NewStyle().Foreground(lipgloss.Color("#ff0000")),
	}

	checks := make([]string, 0)
	for s, k := range pi.GetChecksByStatus() {
		if k > 0 {
			ss, ok := changesMap[s]
			if !ok {
				ss = s
			}
			st, ok := stylesMap[s]
			if !ok {
				st = lipgloss.NewStyle()
			}
			checks = append(checks, st.Render(fmt.Sprintf("%d%s", k, ss)))
		}
	}
	return checks
}

func renderContexts(pi sv.PullRequestStatus) []string {
	changesMap := map[string]string{
		"SUCCESS": "👌",
		"FAILED":  "👎",
	}

	stylesMap := map[string]lipgloss.Style{
		"SUCCESS": lipgloss.NewStyle().Foreground(lipgloss.Color("#00ff00")),
		"FAILED":  lipgloss.NewStyle().Foreground(lipgloss.Color("#ff0000")),
	}

	contexts := make([]string, 0)
	for s, k := range pi.GetContextByStatus() {
		if k > 0 {
			ss, ok := changesMap[s]
			if !ok {
				ss = s
			}
			st, ok := stylesMap[s]
			if !ok {
				st = lipgloss.NewStyle()
			}
			contexts = append(contexts, st.Render(fmt.Sprintf("%d%s", k, ss)))
		}
	}
	return contexts
}

func renderReviews(pi sv.PullRequestStatus) []string {
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

	changesMap := map[string]string{
		"APPROVED":          "👌",
		"CHANGES_REQUESTED": "♺",
		"COMMENTED":         "💬",
	}

	stylesMap := map[string]lipgloss.Style{
		"APPROVED":          lipgloss.NewStyle().Foreground(lipgloss.Color("#00ff00")),
		"CHANGES_REQUESTED": lipgloss.NewStyle().Foreground(lipgloss.Color("#ff0000")),
		"COMMENTED":         lipgloss.NewStyle().Foreground(lipgloss.Color("#e0e0e0")),
	}

	for s, byStatus := range reviewCount {
		var stats string
		stats = fmt.Sprintf("%d", len(byStatus))
		ss, ok := changesMap[s]
		if !ok {
			ss = s
		}
		style, ok := stylesMap[s]
		if !ok {
			style = lipgloss.NewStyle().Foreground(lipgloss.Color("#ffffff"))
		}
		reviews = append(reviews, style.Render(fmt.Sprintf("%s %s", stats, ss)))
	}
	return reviews
}

func setupTable() tea.Msg {
	return setupTableMsg{}
}

func (p PrStatusView) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	cmds := make([]tea.Cmd, 0)
	var pp tea.Model = p
	switch m := msg.(type) {
	case spinner.TickMsg:
		if !p.loaded {
			s, cmd := p.loadingStatus.Update(m)
			p.loadingStatus = s
			cmds = append(cmds, cmd)
			pp = p
		}
	case tea.WindowSizeMsg:
		// save window info
		p.w = m.Width
		p.h = m.Height
		pp = p
		cmds = append(cmds, setupTable)
	case showStatusError:
		p_, cmd := m.Update(p)
		pp = p_
		cmds = append(cmds, cmd)
	case finishedLoadingMsg:
		p_, cmd := m.Update(p)
		pp = p_
		cmds = append(cmds, cmd)
	case loadPrStatusMsg:
		p_, cmd := m.Update(p)
		pp = p_
		cmds = append(cmds, cmd)
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
	case "enter":
		pr := p.pullRequests[p.statusTable.GetHighlightedRowIndex()]
		if pr.GetRepository() == p.sv.GetRepositoryFullName() {
			if prr, err := p.sv.GetPullRequest(fmt.Sprintf("%d", pr.GetId())); err == nil {
				if err := ui.ShowPr(prr); err != nil {
					cmds = append(cmds, showStatusErrorCmd(fmt.Sprintf("Error while showing load pr %d : %e", pr.GetId(), err)))
				}
			} else {
				cmds = append(cmds, showStatusErrorCmd(fmt.Sprintf("Cannot load pr %d : %e", pr.GetId(), err)))
			}
		} else {
			cmds = append(cmds, showStatusErrorCmd(fmt.Sprintf("Repo '%s' doesn't match with '%s'", pr.GetRepository(), p.sv.GetRepositoryFullName())))
		}
	case "runes":
		switch r := m.Runes[0]; r {
		case 'q':
			cmds = append(cmds, tea.Quit)
		case 'r':
			cmds = append(cmds, loadPrStatusCmd)
		}
	}

	// delegate table
	statusTable_, cmd := pp.statusTable.Update(m)
	cmds = append(cmds, cmd)
	pp.statusTable = statusTable_

	return pp, tea.Batch(cmds...)
}

func (p PrStatusView) View() string {
	if !p.loaded {
		return lipgloss.JoinHorizontal(lipgloss.Center, p.loadingStatus.View(), " Loading PR status ")
	}

	if !p.ready {
		return "... initializing ..."
	}

	return p.statusTable.View()
}

func RunPrStatusView(s sv.Sv) error {

	view := PrStatusView{
		sv:            s,
		w:             0,
		h:             0,
		ready:         false,
		loaded:        false,
		loadingStatus: spinner.New(),
		asyncMsg:      make(chan tea.Cmd),
	}

	prg := tea.NewProgram(view)

	go func() {
		for cmd := range view.asyncMsg {
			prg.Send(cmd())
		}
	}()

	defer close(view.asyncMsg)

	if err := prg.Start(); err != nil {
		return err
	}

	return nil

}
