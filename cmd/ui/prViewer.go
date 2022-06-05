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
	"math"
	"strings"
)

const COMMENT_CATEGORY = "COMMENT"

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
	boxer         boxer.Boxer
	layoutMode    layoutMode
	content       contentView
	fileList      fileList
	pullRequest   *pullRequestData
	ready         bool
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
			maxReviews:     4,
			maxChecks:      2,
		}

		content := contentView{
			data: data,
		}

		mode := newLayoutMode()

		fileView := fileList{data, 0, 0}

		if layout, err := initWidgetsLayout(&box, header, content, fileView, mode); err != nil {
			pterm.Fatal.Print(err)
			return nil, err
		} else {
			box.LayoutTree = layout

			prv := &PullRequestView{
				boxer:       box,
				layoutMode:  mode,
				pullRequest: data,
				content:     content,
				header:      header,
				fileList:    fileView,
				bookmarks: map[string][]int{
					COMMENT_CATEGORY: make([]int, 0),
				},
				bookmarksData: make(map[int]interface{}),
				dirty:         true,
				xOffset:       0}

			return prv, nil
		}
	}
}

type fileList struct {
	pullRequestData *pullRequestData
	w               int
	h               int
}

func (f fileList) Init() tea.Cmd {
	return nil
}

func (f fileList) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch m := msg.(type) {
	case tea.WindowSizeMsg:
		f.w = m.Width
		f.h = m.Height
	}
	return f, nil
}

func fillLine(s string, w int) string {
	l := min1(w, len(s))
	s = s[0:l]
	for len(s) < w {
		s = s + " "
	}
	return s
}

func (f fileList) View() string {
	s := lipgloss.NewStyle().Width(f.w).Inline(true)
	l := make([]string, 0)
	for _, fn := range f.pullRequestData.files {
		l = append(l, s.Render(fillLine(fn.NewName, f.w)))
		if len(l) >= f.h {
			break
		}
	}
	return strings.Join(l, "\n")
}

func getModel[T tea.Model](box *boxer.Boxer, name viewAddress) (T, bool) {
	model, ok := box.ModelMap[string(name)]
	return model.(T), ok
}

func getModelAndNode[T tea.Model](box *boxer.Boxer, name viewAddress) (T, *boxer.Node, bool) {
	if model, ok := getModel[T](box, name); ok {
		node := box.CreateLeaf(string(name), model)
		return model, &node, true
	} else {
		return model, nil, false
	}
}

type viewAddress string

const (
	HEADER_ADDRESS   viewAddress = "header"
	CONTENT_ADDRESS  viewAddress = "view"
	FILEVIEW_ADDRESS viewAddress = "files"
)

type layoutMode struct {
	showFileView bool
}

func newLayoutMode() layoutMode {
	return layoutMode{
		showFileView: false,
	}
}

func (mode layoutMode) withFileView(visible bool) layoutMode {
	return layoutMode{
		showFileView: visible,
	}
}

func initWidgetsLayout(box *boxer.Boxer, header pullRequestHeader, content contentView, fileView fileList, mode layoutMode) (boxer.Node, error) {

	box.ModelMap = make(map[string]tea.Model)

	box.ModelMap[string(HEADER_ADDRESS)] = header
	box.ModelMap[string(CONTENT_ADDRESS)] = content
	box.ModelMap[string(FILEVIEW_ADDRESS)] = fileView

	return layoutWidgets(box, mode)
}

func layoutWidgets(box *boxer.Boxer, mode layoutMode) (boxer.Node, error) {

	if header, headerNode, ok := getModelAndNode[pullRequestHeader](box, HEADER_ADDRESS); ok {
		if _, contentNode, ok := getModelAndNode[contentView](box, CONTENT_ADDRESS); ok {
			if _, listNode, ok := getModelAndNode[fileList](box, FILEVIEW_ADDRESS); ok {
				var bottomNode boxer.Node

				if mode.showFileView {
					bottomNode = boxer.Node{
						VerticalStacked: false,
						Children: []boxer.Node{
							*listNode, *contentNode,
						},
						SizeFunc: func(node boxer.Node, width int) []int {
							return []int{
								width / 3,
								width - (width / 3),
							}
						},
					}
				} else {
					bottomNode = *contentNode
				}

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
						*headerNode,
						bottomNode,
					},
				}
				return layout, nil
			}
		}
	}

	return *new(boxer.Node), errors.New("Something's not ok")
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

		prv.addBookmark(COMMENT_CATEGORY, comment)
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

func getNodeRecur(nodes []boxer.Node, name viewAddress) *boxer.Node {
	for _, nd := range nodes {
		if nd.IsLeaf() {
			if nd.GetAddress() == string(name) {
				return &nd
			}
		} else {
			if res := getNodeRecur(nd.Children, name); res != nil {
				return res
			}
		}
	}
	return nil
}

func (p *PullRequestView) getNode(name viewAddress) *boxer.Node {
	return getNodeRecur(p.boxer.LayoutTree.Children, name)
}

func RefreshSizeCmd() tea.Msg {
	if w, h, err := pterm.GetTerminalSize(); err == nil {
		return tea.WindowSizeMsg{
			Width: w, Height: h,
		}
	} else {
		return nil
	}
}

type renderPrMsg struct {
}

func renderPrCmd() tea.Msg {
	return renderPrMsg{}
}

func (p PullRequestView) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	switch msg := msg.(type) {
	case renderPrMsg:
		p.dirty = true
		p.ready = true
		p.renderPullRequest()

	case tea.KeyMsg:

		switch k := msg.String(); k {
		case "ctrl+c":
			return p, tea.Quit
		case "q":
			return p, tea.Quit
		case "esc":
			return p, tea.Quit
		case "v":
			newMode := p.layoutMode.withFileView(!p.layoutMode.showFileView)
			if tree, err := layoutWidgets(&p.boxer, newMode); err == nil {
				p.boxer.LayoutTree = tree
				p.layoutMode = newMode
				p.ready = false
				cmds = append(cmds, tea.ClearScrollArea, RefreshSizeCmd)
			}
		case "right":
			p.xOffset += 4
			cmds = append(cmds, renderPrCmd)

		case "left":
			if p.xOffset >= 4 {
				p.xOffset -= 4
				cmds = append(cmds, renderPrCmd)
			}
		case "n":
			p.moveToNextHeading(COMMIT_LEVEL)
		case "p":
			p.moveToPrevHeading(COMMIT_LEVEL)
		case "N":
			p.moveToNextHeading(FILE_LEVEL)
		case "P":
			p.moveToPrevHeading(FILE_LEVEL)
		case "c":
			p.moveToNextBookmark(COMMENT_CATEGORY)
		case "C":
			p.moveToPrevBookmark(COMMENT_CATEGORY)
		case "r":
			if _, data := currentBookmark(&p, COMMENT_CATEGORY); data != nil {
				if comment, ok := data.(sv.Comment); ok {
					input := confirmation.New(fmt.Sprintf("Whant to reply to comment by %s", comment.GetUser().GetDisplayName()), confirmation.Yes)
					if ready, err := input.RunPrompt(); err != nil && ready {

					}
					return p, tea.ClearScrollArea
				}
			}
		default:
			p.content.viewport, cmd = p.content.viewport.Update(msg)
			cmds = append(cmds, cmd)
		}

	case tea.WindowSizeMsg:
		newBox, cmd := p.boxer.Update(msg)
		// Update from model
		p.updateFromModels()
		p.boxer = newBox.(boxer.Boxer)
		cmds = append(cmds, cmd, renderPrCmd)
	//p, cmds = p.propagateEvent(msg, cmds)
	default:
		p.content.viewport, cmd = p.content.viewport.Update(msg)
		cmds = append(cmds, cmd)

	}

	// Handle keyboard and mouse events in the viewport

	// Update last headings
	for l := 0; l < len(p.header.currentHeading); l++ {
		p.header.currentHeading[l] = -1
		for n, h := range p.header.headings[l] {
			if h.line <= p.content.viewport.YOffset+1 {
				p.header.currentHeading[l] = n
			}
		}
	}

	p.updateModels()

	return p, tea.Batch(cmds...)
}

func (p PullRequestView) propagateEvent(msg tea.Msg, cmds []tea.Cmd) (PullRequestView, []tea.Cmd) {
	// Recursively update the sub-widgets
	newFile, cmd := p.fileList.Update(msg)
	p.fileList = newFile.(fileList)
	cmds = append(cmds, cmd)

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
	p.boxer.ModelMap[string(HEADER_ADDRESS)] = p.header
	p.boxer.ModelMap[string(CONTENT_ADDRESS)] = p.content
	p.boxer.ModelMap[string(FILEVIEW_ADDRESS)] = p.fileList
}

func (p *PullRequestView) updateFromModels() {
	p.header, _ = getModel[pullRequestHeader](&p.boxer, HEADER_ADDRESS)
	p.content, _ = getModel[contentView](&p.boxer, CONTENT_ADDRESS)
	p.fileList, _ = getModel[fileList](&p.boxer, FILEVIEW_ADDRESS)
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
	maxReviews     int
	maxChecks      int
}

func min(a ...int) (int, int) {
	m := math.MaxInt
	p := -1
	for i, x := range a {
		if x < m {
			m = x
			p = i
		}
	}
	return m, p
}

func min1(a ...int) int {
	m, _ := min(a...)
	return m
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

	for n, chk := range prv.pullRequest.checks {
		prv.header.header.printf("* %s : %s (%s)", chk.GetStatus(), chk.GetName(), chk.GetUrl())
		if n >= prv.header.maxChecks-1 {
			break
		}
	}

	for n, rev := range prv.pullRequest.reviews {
		prv.header.header.printf("* %s : %s (%s)", rev.GetState(), rev.GetAuthor(), rev.GetSubmitedAt())
		if n >= prv.header.maxReviews-1 {
			break
		}
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
			copy(vv, v)

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
