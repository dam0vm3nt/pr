package statusView

import (
	"fmt"
	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/evertras/bubble-table/table"
	"github.com/pterm/pterm"
	"github.com/vballestra/sv/cmd/ui"
	"github.com/vballestra/sv/sv"
	"os/exec"
	"strings"
	"time"
)

type PrStatusView struct {
	sv           sv.Sv
	pullRequests []sv.PullRequestStatus
	w            int
	h            int

	ready  bool
	loaded bool

	statusTable      table.Model
	asyncMsg         chan tea.Cmd
	loadingStatus    spinner.Model
	countdownStatus  progress.Model
	loadingCountdown int
	showingPr        bool
	isMonitoring     bool
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

var mineStyle = lipgloss.NewStyle().Bold(true).ColorWhitespace(true).Foreground(lipgloss.Color("#40e040"))
var theirStyle = lipgloss.NewStyle().Bold(true).ColorWhitespace(true).Foreground(lipgloss.Color("#a0a0a0"))
var notLocalStyle = lipgloss.NewStyle().Italic(true).ColorWhitespace(true).Foreground(lipgloss.Color("#e0e0e0"))

func (m setupTableMsg) Update(p PrStatusView) (PrStatusView, tea.Cmd) {
	rows := make([]table.Row, 0)
	for _, pi := range p.pullRequests {

		isLocal := pi.GetRepository() == p.sv.GetRepositoryFullName()

		checks := renderChecks(pi)

		contexts := renderContexts(pi)

		reviews := renderReviews(pi)

		style := (func() lipgloss.Style {
			if !isLocal {
				return notLocalStyle
			} else if pi.IsMine() {
				return mineStyle
			} else {
				return theirStyle
			}
		})()

		row := table.NewRow(table.RowData{
			colId:         style.Render(fmt.Sprintf("%5d", pi.GetId())),
			colTitle:      style.Render(pi.GetTitle()),
			colAuthor:     style.Render(pi.GetAuthor()),
			colRepository: style.Render(pi.GetRepository()),
			colBranch:     style.Render(pi.GetBranchName()),
			colState:      pi.GetStatus(),
			colReviews:    strings.Join(reviews, " "),
			colChecks:     strings.Join(checks, ", "),
			colContexts:   strings.Join(contexts, " "),
		})
		rows = append(rows, row)
	}
	p.statusTable = table.New([]table.Column{
		table.NewColumn(colId, "ID", 5),
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
		"SUCCESS":     "✔",
		"FAILURE":     "⤫",
		"IN_PROGRESS": "☯",
		"QUEUED":      "䷄",
	}

	stylesMap := map[string]lipgloss.Style{
		"SUCCESS": lipgloss.NewStyle().Foreground(lipgloss.Color("#00ff00")),
		"FAILURE": lipgloss.NewStyle().Foreground(lipgloss.Color("#ff0000")),
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
		"SUCCESS": "✔",
		"ERROR":   "⤫",
		"PENDING": "☯",
	}

	stylesMap := map[string]lipgloss.Style{
		"SUCCESS": lipgloss.NewStyle().Foreground(lipgloss.Color("#00ff00")),
		"ERROR":   lipgloss.NewStyle().Foreground(lipgloss.Color("#ff0000")),
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

type showPrMsg struct {
	pullRequest sv.PullRequest
}

func showPrCmd(s sv.Sv, id string) tea.Cmd {
	return func() tea.Msg {
		if prr, err := s.GetPullRequest(id); err == nil {
			return showPrMsg{prr}
		} else {
			return showStatusErrorCmd(fmt.Sprintf("Cannot load pr %s : %e", id, err))()
		}
	}
}

type modelAndCmd struct {
	model tea.Model
	cmd   tea.Cmd
}

type fetchAndReloadMsg struct {
	pr sv.PullRequest
}

func fetchAndReloadCmd(pr sv.PullRequest) tea.Cmd {
	return func() tea.Msg {
		return fetchAndReloadMsg{pr}
	}
}

func execGitFetch() error {
	if path, err := exec.LookPath("git"); err != nil {
		return err
	} else {
		cmd := exec.Command(path, "fetch")
		// cmd.Stdin = os.Stdin
		// cmd.Stdout = os.Stdout
		if err = cmd.Start(); err != nil {
			return err
		}
		if err = cmd.Wait(); err != nil {
			if err, ok := err.(*exec.ExitError); ok {
				if err.ExitCode() != 1 {
					return err
				} else {
					// We can ignore exit code 1 from fetch.
					return nil
				}
			}
			return err
		}

		return nil
	}

}

func forceFetch(repo sv.Sv) error {
	if err := repo.Fetch(); err == nil {
		return nil
	} else if err := execGitFetch(); err == nil {
		return nil
	} else {
		return err
	}
}

func (m showPrMsg) Update(view PrStatusView) (tea.Model, tea.Cmd) {
	w := make(chan modelAndCmd)
	go func() {
		cmds := make([]tea.Cmd, 0)
		if err := ui.ShowPr(m.pullRequest); err != nil {
			if _, ok := err.(*sv.MissingCommitError); ok {
				// Let's try updating the archive
				if err := forceFetch(view.sv); err == nil {
					cmds = append(cmds, showPrCmd(view.sv, fmt.Sprintf("%d", m.pullRequest.GetId())))
				} else {
					cmds = append(cmds, showStatusErrorCmd(fmt.Sprintf("Error while showing load pr %d : %s", m.pullRequest.GetId(), err)))
				}
			} else {
				cmds = append(cmds, showStatusErrorCmd(fmt.Sprintf("Error while showing load pr %d : %s", m.pullRequest.GetId(), err)))
			}
		} else {
			cmds = append(cmds, tea.ClearScrollArea)
		}

		view.loaded = true
		view.showingPr = false
		w <- modelAndCmd{view, tea.Batch(cmds...)}
		close(w)
	}()

	res := <-w
	return res.model, res.cmd
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
		p.countdownStatus = progress.New(progress.WithScaledGradient("#FF7CCB", "#FDFF8C"),
			progress.WithoutPercentage(),
			(func(min int, max int) progress.Option {
				if m.Width-min > max {
					return progress.WithWidth(max)
				} else if m.Width < min {
					return progress.WithWidth(0)
				} else {
					return progress.WithWidth(m.Width - min)
				}
			})(15, 30))
		pp = p
		cmds = append(cmds, setupTable)
	case delayMsg:
		p_, cmd := m.Update(p)
		pp = p_
		cmds = append(cmds, cmd)
	case showPrMsg:
		p_, cmd := m.Update(p)
		pp = p_
		cmds = append(cmds, cmd)
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

type delayMsg struct {
	initialCounter int
	counter        int
	delay          time.Duration
	cmd            tea.Cmd
}

func delayCmd(initialCounter int, counter int, delay time.Duration, cmd tea.Cmd) tea.Cmd {
	return func() tea.Msg {
		return delayMsg{initialCounter, counter, delay, cmd}
	}
}

func (m delayMsg) Update(p PrStatusView) (tea.Model, tea.Cmd) {
	p.loadingCountdown = m.counter
	if p.isMonitoring {
		go func() {
			var newCounter int
			if m.counter == 0 {
				p.asyncMsg <- m.cmd
				newCounter = m.initialCounter
			} else {
				newCounter = m.counter - 1
			}
			time.Sleep(m.delay)
			p.asyncMsg <- delayCmd(m.initialCounter, newCounter, m.delay, m.cmd)
		}()
	}
	return p, nil
}

func (p PrStatusView) handleKey(m tea.KeyMsg) (tea.Model, tea.Cmd) {
	if p.showingPr {
		return p, nil
	}
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
			go func() { p.asyncMsg <- showPrCmd(p.sv, fmt.Sprintf("%d", pr.GetId())) }()
			cmds = append(cmds, p.loadingStatus.Tick)
			p.showingPr = true
			p.loaded = false
			pp = p
		} else {
			cmds = append(cmds, showStatusErrorCmd(fmt.Sprintf("Repo '%s' doesn't match with '%s'", pr.GetRepository(), p.sv.GetRepositoryFullName())))
		}
	case "runes":
		switch r := m.Runes[0]; r {
		case 'q':
			cmds = append(cmds, tea.Quit)
		case 'r':
			cmds = append(cmds, loadPrStatusCmd)
		case 'm':
			pp.isMonitoring = !p.isMonitoring
			if pp.isMonitoring {
				cmds = append(cmds, delayCmd(30, 30, time.Second, loadPrStatusCmd))
			}
		}
	}

	// delegate table
	statusTable_, cmd := pp.statusTable.Update(m)
	cmds = append(cmds, cmd)
	pp.statusTable = statusTable_

	return pp, tea.Batch(cmds...)
}

func (p PrStatusView) View() string {
	verts := make([]string, 0)

	if p.isMonitoring && p.loadingCountdown > 0 {
		verts = append(verts, fmt.Sprintf("Countdown : %d %s", p.loadingCountdown, p.countdownStatus.ViewAs(float64(p.loadingCountdown)/30.0)))
	}

	if !p.loaded {
		verts = append(verts, lipgloss.JoinHorizontal(lipgloss.Center, p.loadingStatus.View(), " Loading PR status "))
	} else {
		verts = append(verts, "")
	}

	if !p.ready {
		verts = append(verts, "... initializing ...")
	} else {
		verts = append(verts, p.statusTable.View())
	}

	line := strings.Repeat(" ", p.w)
	res := strings.Split(strings.ReplaceAll(lipgloss.JoinVertical(lipgloss.Left, verts...), "\r\n", "\n"), "\n")
	for len(res) < p.h {
		res = append(res, line)
	}

	return strings.Join(res, "\n")
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
