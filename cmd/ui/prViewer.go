package ui

import (
	"errors"
	"fmt"
	"github.com/bluekeyes/go-gitdiff/gitdiff"
	"github.com/charmbracelet/bubbles/viewport"
	"github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
	"github.com/erikgeiser/promptkit/confirmation"
	"github.com/pterm/pterm"
	boxer "github.com/treilik/bubbleboxer"
	"github.com/vballestra/sv/sv"
	"strings"
)

type Heading struct {
	line int
	text string
}

type pullRequestData struct {
	sv.PullRequest
	checks     []sv.Check
	reviews    []sv.Review
	prComments []sv.Comment
	commentMap map[string]map[int64][]sv.Comment
	files      []*gitdiff.File
}

func loadPullRequestData(pr sv.PullRequest) (*pullRequestData, error) {
	if checks, err := pr.GetChecks(); err != nil {
		pterm.Warning.Println("Couldn't read the checks ", err)
		return nil, err
	} else if reviews, err := pr.GetReviews(); err != nil {
		pterm.Warning.Println("Couldn't read the checks ", err)
		return nil, err
	} else if prComments, commentMap, err := pr.GetCommentsByLine(); err != nil {
		return nil, err
	} else if files, err := pr.GetDiff(); err != nil {
		return nil, err
	} else {
		return &pullRequestData{pr,
			checks,
			reviews,
			prComments,
			commentMap,
			files}, nil
	}
}

type PullRequestView struct {
	boxer boxer.Boxer
	//header         *lines
	content     contentView
	pullRequest *pullRequestData
	ready       bool
	//headings       [][]Heading
	//currentHeading []int
	bookmarks     map[string][]int
	bookmarksData map[int]interface{}

	dirty   bool
	xOffset int
	header  pullRequestHeader
}

func NewView(pr sv.PullRequest) (*PullRequestView, error) {

	headings := make([][]Heading, HEADINGS)
	for l := 0; l < HEADINGS; l++ {
		headings[l] = make([]Heading, 0)
	}

	if data, err := loadPullRequestData(pr); err != nil {
		pterm.Warning.Println("Couldn't read pr ", err)
		return nil, err
	} else {
		box := boxer.Boxer{
			HandleMsg: true,
		}

		header := pullRequestHeader{
			data:           data,
			headings:       headings,
			currentHeading: make([]int, HEADINGS),
		}

		content := contentView{
			data: data,
		}

		layout := layoutWidgets(&box, header, content)

		box.LayoutTree = layout

		prv := &PullRequestView{
			boxer:       box,
			pullRequest: data,
			content:     content,
			header:      header,
			bookmarks: map[string][]int{
				"COMMENT": make([]int, 0),
			},
			bookmarksData: make(map[int]interface{}),
			dirty:         true,
			xOffset:       0}

		return prv, nil
	}
}

func layoutWidgets(box *boxer.Boxer, header pullRequestHeader, content contentView) boxer.Node {
	layout := boxer.Node{
		VerticalStacked: true,
		SizeFunc: func(node boxer.Node, height int) []int {
			headerHeight := header.measureHeight()
			return []int{
				headerHeight,
				height - headerHeight,
			}
		},
		Children: []boxer.Node{
			box.CreateLeaf("header", header),
			box.CreateLeaf("view", content),
		},
	}
	return layout
}

func ptr[T string | uint | int](s T) *T {
	return &s
}

func (prv *PullRequestView) PrintComments(comments []sv.Comment, w int) {
	bg := "#7D56F4"
	fg := "#FAFAFA"
	style := lipgloss.NewStyle().
		Italic(true).
		Foreground(lipgloss.Color(fg)).
		Background(lipgloss.Color(bg)).
		Align(lipgloss.Left).
		Width(w)

	style2 := style.Copy().
		PaddingLeft(2).
		PaddingRight(2).
		Align(lipgloss.Left)

	st := glamour.DraculaStyleConfig

	st.Document.BackgroundColor = ptr(bg)
	st.Document.Color = ptr(fg)
	st.Document.Margin = ptr(uint(0))
	st.Link.BackgroundColor = ptr(bg)

	r, _ := glamour.NewTermRenderer(
		// detect background color and pick either the default dark or light theme
		//glamour.WithAutoStyle(),
		// wrap output at specific width
		glamour.WithWordWrap(w),
		glamour.WithEmoji(),
		glamour.WithStyles(st),
	)

	for _, comment := range comments {
		raw := comment.GetContent().GetRaw()
		reply := ""
		if id := comment.GetParentId(); id != nil {
			reply = fmt.Sprintf(" <- %d", id)
		}

		prv.addBookmark("COMMENT", comment)
		prv.addHeading(w, COMMIT_LEVEL, style.Render(
			fmt.Sprintf("------- [%d%s] %s at %s ------",
				comment.GetId(), reply,
				comment.GetUser().GetDisplayName(),
				comment.GetCreatedOn())))

		rawRendered, _ := r.Render(raw)
		prv.content.printf(style2.Render(rawRendered))
		prv.removeLastHeader(COMMIT_LEVEL)
	}
}

func (c *contentView) currentLine() int {
	return len(*c.content)
}

func (prv *PullRequestView) removeLastHeader(lev int) {
	prv.header.headings[lev] = append(prv.header.headings[lev], Heading{prv.content.currentLine(), prv.header.headings[lev][len(prv.header.headings[lev])-2].text})
}

func (prv *PullRequestView) addBookmark(b string, data interface{}) {
	prv.bookmarksData[len(prv.bookmarks[b])] = data
	prv.bookmarks[b] = append(prv.bookmarks[b], prv.content.currentLine()+1)
}

func (p *PullRequestView) addHeading(w int, lev int, format string, args ...any) {
	var st lipgloss.Style
	switch lev {
	case FILE_LEVEL:
		st = lipgloss.NewStyle().
			Background(lipgloss.Color("#d040d0")).
			Foreground(lipgloss.Color("#ffffff")).
			Bold(true).
			Width(w)
	default:
		st = lipgloss.NewStyle().
			Background(lipgloss.Color("#909090")).
			Foreground(lipgloss.Color("#ffffff")).
			Width(w)
	}

	s := st.Render(fmt.Sprintf(format, args...))
	p.content.printf(s)
	p.header.headings[lev] = append(p.header.headings[lev], Heading{p.content.currentLine() + 1, s})
}

func (prv *lines) printf(format string, args ...any) {
	res := fmt.Sprintf(format, args...)
	ln := strings.Split(res, "\n")
	if len(ln) > 1 && len(ln[len(ln)-1]) == 0 {
		ln = ln[:len(ln)-1]
	}
	*prv = append(*prv, ln...)
}

func (p PullRequestView) Init() tea.Cmd {
	return nil
}

func (p *PullRequestView) moveToNextHeading(L int) {
	if p.header.currentHeading[L] >= 0 && p.header.currentHeading[L] < len(p.header.headings[L])-1 {
		p.header.currentHeading[L] += 1
		p.content.viewport.YOffset = p.header.headings[L][p.header.currentHeading[L]].line - 1
	} else if len(p.header.headings[L]) > 0 {
		p.header.currentHeading[L] = 0
		p.content.viewport.YOffset = p.header.headings[L][p.header.currentHeading[L]].line - 1
	}
}

func (p *PullRequestView) moveToPrevHeading(L int) {
	if len(p.header.headings[L]) > 0 && p.header.currentHeading[L] > 0 {
		p.header.currentHeading[L] -= 1
		p.content.viewport.YOffset = p.header.headings[L][p.header.currentHeading[L]].line - 1
	} else if len(p.header.headings[L]) > 0 {
		p.header.currentHeading[L] = len(p.header.headings[L]) - 1
		p.content.viewport.YOffset = p.header.headings[L][p.header.currentHeading[L]].line - 1
	}
}

func (p *PullRequestView) moveToNextBookmark(b string) error {
	bookmarks := p.bookmarks[b]
	if len(bookmarks) == 0 {
		return errors.New("No bookmarks")
	}
	for _, l := range bookmarks {
		if l > p.content.viewport.YOffset {
			p.content.viewport.YOffset = l
			return nil
		}
	}
	p.content.viewport.YOffset = bookmarks[0]
	return nil
}

func (p *PullRequestView) moveToPrevBookmark(b string) error {
	bookmarks := p.bookmarks[b]
	if len(bookmarks) == 0 {
		return errors.New("No bookmarks")
	}
	for i := len(bookmarks) - 1; i >= 0; i-- {
		l := bookmarks[i]
		if l < p.content.viewport.YOffset {
			p.content.viewport.YOffset = l
			return nil
		}
	}
	p.content.viewport.YOffset = bookmarks[len(bookmarks)-1]
	return nil
}

func currentBookmark(p *PullRequestView, b string) (int, interface{}) {
	for n, l := range p.bookmarks[b] {
		if l == p.content.viewport.YOffset {
			return n, p.bookmarksData[n]
		}
	}
	return 0, nil
}

func (p PullRequestView) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	p.renderPullRequest()

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch k := msg.String(); k {
		case "ctrl+c":
			return p, tea.Quit
		case "q":
			return p, tea.Quit
		case "esc":
			return p, tea.Quit
		case "right":
			p.xOffset += 4
			p.dirty = true
			p.renderPullRequest()
			// return p, tea.ClearScrollArea
		case "left":
			if p.xOffset >= 4 {
				p.xOffset -= 4
				p.dirty = true
			}
			p.renderPullRequest()
			// return p, tea.ClearScrollArea
		case "n":
			p.moveToNextHeading(COMMIT_LEVEL)
		case "p":
			p.moveToPrevHeading(COMMIT_LEVEL)
		case "N":
			p.moveToNextHeading(FILE_LEVEL)
		case "P":
			p.moveToPrevHeading(FILE_LEVEL)
		case "c":
			p.moveToNextBookmark("COMMENT")
		case "C":
			p.moveToPrevBookmark("COMMENT")
		case "r":
			if _, data := currentBookmark(&p, "COMMENT"); data != nil {
				if comment, ok := data.(sv.Comment); ok {
					input := confirmation.New(fmt.Sprintf("Whant to reply to comment by %s", comment.GetUser().GetDisplayName()), confirmation.Yes)
					if ready, err := input.RunPrompt(); err != nil && ready {

					}
					return p, tea.ClearScrollArea
				}
			}
		}

	case tea.WindowSizeMsg:
		p, cmds = p.propagateEvent(msg, cmd, cmds)
		headerHeight := p.header.measureHeight()
		footerHeight := 0
		verticalMarginHeight := headerHeight + footerHeight + 1

		if !p.ready {
			// Since this program is using the full size of the viewport we
			// need to wait until we've received the window dimensions before
			// we can initialize the viewport. The initial dimensions come in
			// quickly, though asynchronously, which is why we wait for them
			// here.
			//p.viewport = viewport.New(msg.Width, msg.Height-verticalMarginHeight)
			p.content.viewport.YPosition = headerHeight
			p.content.viewport.HighPerformanceRendering = false
			p.content.updateViewportWithContent()
			p.ready = true

			// This is only necessary for high performance rendering, which in
			// most cases you won't need.
			//
			// Render the viewport one line below the Header.
			//p.content.viewport.YPosition = headerHeight + 1
		}
		p.content.viewport.Width = msg.Width
		p.content.viewport.Height = msg.Height - verticalMarginHeight
		p.dirty = true
		p.renderPullRequest()
	}

	// Handle keyboard and mouse events in the viewport
	p.content.viewport, cmd = p.content.viewport.Update(msg)

	// Update last headings
	for l := 0; l < len(p.header.currentHeading); l++ {
		p.header.currentHeading[l] = -1
		for n, h := range p.header.headings[l] {
			if h.line <= p.content.viewport.YOffset+1 {
				p.header.currentHeading[l] = n
			}
		}
	}

	cmds = append(cmds, cmd)

	p.updateModels()

	return p, tea.Batch(cmds...)
}

func (p PullRequestView) propagateEvent(msg tea.Msg, cmd tea.Cmd, cmds []tea.Cmd) (PullRequestView, []tea.Cmd) {
	// Recursively update the sub-widgets
	newContent, cmd := p.content.Update(msg)
	p.content = newContent.(contentView)
	cmds = append(cmds, cmd)

	newHeader, cmd := p.header.Update(msg)
	p.header = newHeader.(pullRequestHeader)
	cmds = append(cmds, cmd)

	newBox, cmd := p.boxer.Update(msg)
	p.boxer = newBox.(boxer.Boxer)
	cmds = append(cmds, cmd)
	return p, cmds
}

func (p *PullRequestView) updateModels() {
	p.boxer.ModelMap["header"] = p.header
	p.boxer.ModelMap["view"] = p.content

}

func (p PullRequestView) View() string {
	if !p.ready {
		return "\n  Initializing..."
	}

	return p.boxer.View()
}

type lines []string

type pullRequestHeader struct {
	data           *pullRequestData
	header         *lines
	currentHeading []int
	headings       [][]Heading
	width          int
	height         int
}

func (p pullRequestHeader) measureHeight() int {
	return len(p.data.checks) + len(p.data.reviews) + 4
}

func (p pullRequestHeader) Init() tea.Cmd {
	return nil
}

func (p pullRequestHeader) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		p.width = msg.Width
		p.height = p.measureHeight()
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

type contentView struct {
	data     *pullRequestData
	viewport viewport.Model
	content  *lines
}

func (cv *contentView) printf(line string, args ...any) {
	cv.content.printf(line, args...)
}

func (cv contentView) clear() contentView {
	c := make(lines, 0)
	cv.content = &c
	return cv
}

func (c contentView) Init() tea.Cmd {
	return nil
}

func (c contentView) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	return c, nil
}

func (c contentView) View() string {
	return c.viewport.View()
}

const (
	FILE_LEVEL = iota
	COMMIT_LEVEL
	HEADINGS
)

func (prv *PullRequestView) renderPullRequest() {
	if !prv.dirty {
		return
	}
	h := make(lines, 0)
	prv.content = prv.content.clear()
	prv.header.header = &h

	pr := prv.pullRequest
	sourceBranch := pr.GetBranch().GetName()
	destBranch := pr.GetBase().GetName()
	prv.header.header.printf("#%d %s (%s)\n%s -> %s Status: %s", pr.GetId(), pr.GetTitle(), pr.GetAuthor().GetDisplayName(),
		sourceBranch, destBranch, pr.GetState())

	for _, chk := range prv.pullRequest.checks {
		prv.header.header.printf("* %s : %s (%s)", chk.GetStatus(), chk.GetName(), chk.GetUrl())
	}

	for _, rev := range prv.pullRequest.reviews {
		prv.header.header.printf("* %s : %s (%s)", rev.GetState(), rev.GetAuthor(), rev.GetSubmitedAt())
	}

	prv.PrintComments(prv.pullRequest.prComments, prv.content.viewport.Width)

	//fmt.Printf("Diff of %d files:\n\n", len(files))
	//prv.header.printf("Diff of %d files:\n\n", len(files))

	for _, file := range prv.pullRequest.files {
		fn := file.OldName
		if file.IsRename {
			prv.addHeading(prv.content.viewport.Width, FILE_LEVEL, "%s -> %s:", file.OldName, file.NewName)
			fn = file.NewName
		} else if file.IsDelete {
			prv.addHeading(prv.content.viewport.Width, FILE_LEVEL, "DELETED %s", file.OldName)
			continue
		} else if file.IsNew {
			prv.addHeading(prv.content.viewport.Width, FILE_LEVEL, "NEW %s:", file.NewName)
			fn = file.NewName
		} else if file.IsCopy {
			prv.addHeading(prv.content.viewport.Width, FILE_LEVEL, "COPY %s -> %s:", file.OldName, file.NewName)
		} else {
			prv.addHeading(prv.content.viewport.Width, FILE_LEVEL, "%s:", file.NewName)
		}

		commentsForFileOrig, haveFileComments := prv.pullRequest.commentMap[fn]

		// Clone it
		commentsForFile := make(map[int64][]sv.Comment)
		for k, v := range commentsForFileOrig {
			vv := make([]sv.Comment, len(v))
			for i, x := range v {
				vv[i] = x
			}
			commentsForFile[k] = vv
		}

		if file.IsBinary {
			prv.content.printf("\nBINARY FILE\n")
		} else {
			w := prv.content.viewport.Width
			styleAdd := lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("#ffffff")).
				Background(lipgloss.Color("#005E00e0")).
				Width(w).MaxHeight(1)
			styleNorm := lipgloss.NewStyle().Width(w).MaxHeight(1)
			styleDel := lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("#ffffff")).
				Background(lipgloss.Color("#5e0000e0")).
				Width(w).MaxHeight(1)

			for _, frag := range file.TextFragments {

				prv.addHeading(prv.content.viewport.Width, COMMIT_LEVEL, "==O== ==N== (+%d, -%d,  O=%d, N=%d)", frag.LinesAdded, frag.LinesDeleted,
					frag.OldLines, frag.NewLines)
				oldN := frag.OldPosition
				newN := frag.NewPosition

				for _, ln := range frag.Lines {
					var style lipgloss.Style
					switch ln.Op {
					case gitdiff.OpAdd:
						style = styleAdd
					case gitdiff.OpDelete:
						style = styleDel
					default:
						style = styleNorm
					}

					escaped := strings.ReplaceAll(ln.Line, "%", "%%")
					if len(escaped) >= prv.xOffset {
						escaped = escaped[prv.xOffset:]
					} else {
						escaped = ""
					}

					rendered := style.Render(fmt.Sprintf("%05d %05d %s  %s", oldN, newN, ln.Op, escaped))

					prv.content.printf(rendered)
					if haveFileComments {
						if commentsForLine, haveLineComments := commentsForFile[newN]; haveLineComments {
							prv.PrintComments(commentsForLine, prv.content.viewport.Width)
							delete(commentsForFile, newN)
						} else if commentsForLine, haveLineComments := commentsForFile[-oldN]; haveLineComments {
							prv.PrintComments(commentsForLine, prv.content.viewport.Width)
							delete(commentsForFile, -oldN)
						}
					}

					if ln.Op == gitdiff.OpAdd {
						oldN -= 1
					}
					if ln.Op == gitdiff.OpDelete {
						newN -= 1
					}
					newN += 1
					oldN += 1
				}
			}
		}
	}

	if prv.ready {
		prv.content.updateViewportWithContent()
	}
	prv.dirty = false
}

func (c *contentView) updateViewportWithContent() {
	c.viewport.SetContent(strings.Join(*c.content, "\n"))
}
