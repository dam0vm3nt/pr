package ui

import (
	"errors"
	"fmt"
	"github.com/bluekeyes/go-gitdiff/gitdiff"
	"github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
	"github.com/erikgeiser/promptkit/confirmation"
	"github.com/itchyny/timefmt-go"
	"github.com/pterm/pterm"
	boxer "github.com/treilik/bubbleboxer"
	"github.com/vballestra/sv/cmd/ui/simpleEditor"
	"github.com/vballestra/sv/sv"
	"math"
	"os"
	"os/exec"
	"strings"
)

const (
	COMMENT_CATEGORY = bookmarkCategory("COMMENT")
	FILE_CATEGORY    = bookmarkCategory("FILE")
)

type Heading struct {
	line    int
	text    string
	lineEnd int
}

type pullRequestData struct {
	sv.PullRequest
	checks        []sv.Check
	reviews       []sv.Review
	prComments    []sv.Comment
	commentMap    map[string]map[int64][]sv.Comment
	files         []*gitdiff.File
	lastCommitId  string
	pendingReview sv.Review
}

func (d *pullRequestData) addComment(path string, old int64, new int64, isNew bool, comment sv.Comment) {
	n := old
	if isNew {
		n = -new
	}
	fileComments, ok := d.commentMap[path]
	if !ok {
		fileComments = make(map[int64][]sv.Comment)
	}
	lineComments, ok := fileComments[n]
	if !ok {
		lineComments = make([]sv.Comment, 0)
	}

	lineComments = append(lineComments, comment)
	fileComments[n] = lineComments
	d.commentMap[path] = fileComments
}

func loadPullRequestData(pr sv.PullRequest) (*pullRequestData, error) {
	if checks, err := pr.GetChecks(); err != nil {
		pterm.Warning.Println("Couldn't read the checks ", err)
		return nil, err
	} else if reviews, err := pr.GetReviews(); err != nil {
		pterm.Warning.Println("Couldn't read the checks ", err)
		return nil, err
	} else if pending, err := pr.GetPendingReview(); err != nil {
		pterm.Warning.Println("Couldn't read the pending review ", err)
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
			files,
			pr.GetLastCommitId(),
			pending}, nil
	}
}

type bookmarkCategory string

type bookmark struct {
	line int
	data interface{}
}

type PullRequestView struct {
	boxer       boxer.Boxer
	layoutMode  layoutMode
	pullRequest *pullRequestData
	ready       bool
	bookmarks   map[bookmarkCategory][]bookmark
	mainFocus   int

	dirty   bool
	xOffset int
}

var focusOrder = [...]viewAddress{CONTENT_ADDRESS, FILEVIEW_ADDRESS}

func NewView(pr sv.PullRequest) (*PullRequestView, error) {

	headings := make([][]Heading, HEADINGS)
	for l := 0; l < int(HEADINGS); l++ {
		headings[l] = make([]Heading, 0)
	}

	if data, err := loadPullRequestData(pr); err != nil {
		pterm.Debug.Println("Couldn't read pr ", err)
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
			maxChecks:      5,
		}

		content := contentView{
			data:         data,
			selectedLine: -1,
		}

		mode := newLayoutMode()

		fileView := fileList{data, 0, 0, false, 0, 0}

		if layout, err := initWidgetsLayout(&box, header, content, fileView, mode); err != nil {
			pterm.Fatal.Print(err)
			return nil, err
		} else {
			box.LayoutTree = layout

			prv := &PullRequestView{
				boxer:       box,
				layoutMode:  mode,
				pullRequest: data,
				bookmarks: map[bookmarkCategory][]bookmark{
					COMMENT_CATEGORY: make([]bookmark, 0),
					FILE_CATEGORY:    make([]bookmark, 0),
				},
				mainFocus: 0,
				dirty:     true,
				xOffset:   0}

			return prv, nil
		}
	}
}

func fillLine(s string, w int) string {
	r := strings.NewReplacer("\n", "", "\r", "", "\t", "    ")
	s = r.Replace(s)
	l := min1(w, len(s))
	s = s[0:l]
	for len(s) < w {
		s = s + strings.Repeat(" ", w-len(s))
	}
	return s
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

func (p *PullRequestView) getHeaderView() (pullRequestHeader, bool) {
	return getModel[pullRequestHeader](&p.boxer, HEADER_ADDRESS)
}

func (p *PullRequestView) getContentView() (contentView, bool) {
	return getModel[contentView](&p.boxer, CONTENT_ADDRESS)
}

func (p *PullRequestView) getFileListView() (fileList, bool) {
	return getModel[fileList](&p.boxer, FILEVIEW_ADDRESS)
}

func (p *PullRequestView) nextFocus() {
	var i int
	for i = (p.mainFocus + 1) % len(focusOrder); !p.isVisible(focusOrder[i]); i = (i + 1) % len(focusOrder) {
	}
	p.mainFocus = i
}

func (p *PullRequestView) isVisible(view viewAddress) bool {
	switch view {
	case FILEVIEW_ADDRESS:
		return p.layoutMode.showFileView
	default:
		return true
	}
}

func (p *PullRequestView) currentFocus() viewAddress {
	return focusOrder[p.mainFocus]
}

func withView[T tea.Model](p *PullRequestView, address viewAddress, action func(T) (T, error)) error {
	if m, ok := getModel[T](&p.boxer, address); ok {
		if m, err := action(m); err != nil {
			return err
		} else {
			p.boxer.ModelMap[string(address)] = m
			return nil
		}
	} else {
		return errors.New(fmt.Sprintf("Cannot find view %s", address))
	}
}

type usage struct {
	value interface{}
	count int
}

var viewGuards = make(map[viewAddress]usage)

func removeGuard(address viewAddress) {
	if u, ok := viewGuards[address]; ok {
		u.count--
		if u.count > 0 {
			viewGuards[address] = u
		} else {
			delete(viewGuards, address)
		}
	}
}

func withViewPtr[T tea.Model](p *PullRequestView, address viewAddress, action func(*T) error) error {
	if u, ok := viewGuards[address]; ok {
		u.count += 1
		viewGuards[address] = u
		defer removeGuard(address)
		return action(u.value.(*T))
	}

	if m, ok := getModel[T](&p.boxer, address); ok {
		viewGuards[address] = usage{value: &m, count: 0}
		defer removeGuard(address)
		if err := action(&m); err != nil {
			return err
		} else {
			p.boxer.ModelMap[string(address)] = m
			return nil
		}
	} else {
		return fmt.Errorf("Cannot find view '%s'", address)
	}
}

func (p *PullRequestView) withHeaderView(action func(pullRequestHeader) (pullRequestHeader, error)) error {
	return withView(p, HEADER_ADDRESS, action)
}

func (p *PullRequestView) withContentView(action func(view contentView) (contentView, error)) error {
	return withView(p, CONTENT_ADDRESS, action)
}

func (p *PullRequestView) withStatusBarView(action func(view statusBar) (statusBar, error)) error {
	return withView(p, STATUSBAR_ADDRESS, action)
}

func (p *PullRequestView) withFileListView(action func(view fileList) (fileList, error)) error {
	return withView(p, FILEVIEW_ADDRESS, action)
}

func (p *PullRequestView) withHeaderViewPtr(action func(*pullRequestHeader) error) error {
	return withViewPtr(p, HEADER_ADDRESS, action)
}

func (p *PullRequestView) withContentViewPtr(action func(*contentView) error) error {
	return withViewPtr(p, CONTENT_ADDRESS, action)
}

func (p *PullRequestView) withFileListViewPtr(action func(*fileList) error) error {
	return withViewPtr(p, FILEVIEW_ADDRESS, action)
}

type viewAddress string

const (
	HEADER_ADDRESS    viewAddress = "header"
	CONTENT_ADDRESS   viewAddress = "view"
	FILEVIEW_ADDRESS  viewAddress = "files"
	STATUSBAR_ADDRESS viewAddress = "statusBar"
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
	box.ModelMap[string(STATUSBAR_ADDRESS)] = newStatusBar()

	return layoutWidgets(box, mode)
}

func layoutWidgets(box *boxer.Boxer, mode layoutMode) (boxer.Node, error) {

	if header, headerNode, ok := getModelAndNode[pullRequestHeader](box, HEADER_ADDRESS); ok {
		if _, contentNode, ok := getModelAndNode[contentView](box, CONTENT_ADDRESS); ok {
			if _, listNode, ok := getModelAndNode[fileList](box, FILEVIEW_ADDRESS); ok {
				if _, statusNode, ok := getModelAndNode[statusBar](box, STATUSBAR_ADDRESS); ok {
					var bottomNode *boxer.Node

					if mode.showFileView {
						bottomNode = &boxer.Node{
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
						bottomNode = contentNode
					}

					layout := boxer.CreateNoBorderNode()
					layout.VerticalStacked = true
					layout.SizeFunc = func(node boxer.Node, height int) []int {
						headerHeight := header.measureHeight()
						return []int{
							headerHeight,
							height - 1 - headerHeight,
							1,
						}
					}
					layout.Children = []boxer.Node{
						*headerNode,
						*bottomNode,
						*statusNode,
					}
					return layout, nil
				}
			}
		}
	}

	return *new(boxer.Node), errors.New("Something's not ok")
}

func ptr[T string | uint | int](s T) *T {
	return &s
}

func (prv *PullRequestView) PrintComments(content *contentView, header *pullRequestHeader, comments []sv.Comment, w int) {
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
			reply = fmt.Sprintf(" <- %s", id)
		}

		prv.addBookmark(content, COMMENT_CATEGORY, comment)
		addHeading(content, header, w, COMMIT_LEVEL, style.Render(
			fmt.Sprintf("------- [%s%s] %s at %s ------",
				comment.GetId(), reply,
				comment.GetUser().GetDisplayName(),
				comment.GetCreatedOn())))

		rawRendered, _ := r.Render(raw)
		content.printf(style2.Render(
			strings.ReplaceAll(
				strings.ReplaceAll(rawRendered, "\t", "    "), "\r", "\n")))

		// Print reactions
		reactions := make([]string, 0)
		reactionIcons := map[string]string{
			"THUMBS_UP":   "👍",
			"THUMBS_DOWN": "👎",
			"LAUGH":       "😁",
			"HOORAY":      "🕺",
			"CONFUSED":    "🤔",
			"HEART":       "🫶",
			"ROCKET":      "🚀",
			"EYES":        "👀",
		}
		for r, u := range comment.GetReactions() {
			icon, ok := reactionIcons[r]
			if !ok {
				icon = r
			}
			reactions = append(reactions, fmt.Sprintf("%s(%d)", icon, len(u)))
		}
		content.printf(style2.Render(strings.Join(reactions, " ")))

		prv.closeLastHeader(header, content, COMMIT_LEVEL)
	}
}

func (c *contentView) currentLine() int {
	return len(*c.content)
}

func (prv *PullRequestView) closeLastHeader(header *pullRequestHeader, contentView *contentView, lev headingsLevel) {
	header.headings[lev][len(header.headings[lev])-1].lineEnd = contentView.currentLine()
}

func (prv *PullRequestView) addBookmark(contentView *contentView, b bookmarkCategory, data interface{}) {
	prv.bookmarks[b] = append(prv.bookmarks[b], bookmark{contentView.currentLine() + 1, data})
}

func addHeading(content *contentView, header *pullRequestHeader, w int, lev headingsLevel, format string, args ...any) int {
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
	content.printf(s)
	h := Heading{content.currentLine() + 1, s, -1}
	header.headings[lev] = append(header.headings[lev], h)
	return len(header.headings[lev]) - 1
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

type direction int

const (
	NEXT direction = iota
	PREV
)

type moveToHeadingMsg struct {
	level headingsLevel
	dir   direction
}

func moveToHeadingCmd(lev headingsLevel, dir direction) tea.Cmd {
	return func() tea.Msg {
		return moveToHeadingMsg{lev, dir}
	}
}

func (p *PullRequestView) moveToNextHeading(L headingsLevel) {
	p.withContentViewPtr(func(content *contentView) error {
		return p.withHeaderViewPtr(func(header *pullRequestHeader) error {
			if header.currentHeading[L] >= 0 && header.currentHeading[L] < len(header.headings[L])-1 {
				header.currentHeading[L] += 1
				content.viewport.YOffset = header.headings[L][header.currentHeading[L]].line - 1
			} else if len(header.headings[L]) > 0 {
				header.currentHeading[L] = 0
				content.viewport.YOffset = header.headings[L][header.currentHeading[L]].line - 1
			}
			return nil
		})
	})
}

func (p *PullRequestView) moveToPrevHeading(L headingsLevel) {
	p.withContentViewPtr(func(content *contentView) error {
		return p.withHeaderViewPtr(func(header *pullRequestHeader) error {
			if len(header.headings[L]) > 0 && header.currentHeading[L] > 0 {
				header.currentHeading[L] -= 1
				content.viewport.YOffset = header.headings[L][header.currentHeading[L]].line - 1
			} else if len(header.headings[L]) > 0 {
				header.currentHeading[L] = len(header.headings[L]) - 1
				content.viewport.YOffset = header.headings[L][header.currentHeading[L]].line - 1
			}
			return nil
		})
	})
}

type moveToBookmarkMsg struct {
	cat bookmarkCategory
	dir direction
}

func moveToNextPrevBookmarkCmd(cat bookmarkCategory, dir direction) tea.Cmd {
	return func() tea.Msg {
		return moveToBookmarkMsg{cat, dir}
	}
}

func (p *PullRequestView) moveToNextBookmark(b bookmarkCategory) error {
	return p.withContentViewPtr(func(content *contentView) error {
		bookmarks := p.bookmarks[b]
		if len(bookmarks) == 0 {
			return errors.New("No bookmarks")
		}
		for _, l := range bookmarks {
			if l.line > content.viewport.YOffset {
				content.viewport.YOffset = l.line
				return nil
			}
		}
		content.viewport.YOffset = bookmarks[0].line
		return nil
	})
}

func (p *PullRequestView) moveToBookmark(b bookmarkCategory, ordinal int) error {
	return p.withContentViewPtr(func(content *contentView) error {
		bookmarks := p.bookmarks[b]
		if len(bookmarks) == 0 || ordinal >= len(bookmarks) {
			return errors.New("No bookmarks")
		}
		content.viewport.YOffset = bookmarks[ordinal].line
		return nil
	})
}

func (p *PullRequestView) moveToPrevBookmark(b bookmarkCategory) error {
	return p.withContentViewPtr(func(content *contentView) error {

		bookmarks := p.bookmarks[b]
		if len(bookmarks) == 0 {
			return errors.New("No bookmarks")
		}
		for i := len(bookmarks) - 1; i >= 0; i-- {
			l := bookmarks[i]
			if l.line < content.viewport.YOffset {
				content.viewport.YOffset = l.line
				return nil
			}
		}
		content.viewport.YOffset = bookmarks[len(bookmarks)-1].line
		return nil
	})
}

func currentBookmark(p *PullRequestView, b bookmarkCategory) (int, interface{}) {
	if content, ok := p.getContentView(); ok {
		return bookmarkAt(p, b, content.viewport.YOffset)
	} else {
		return 0, nil
	}
}

func bookmarkAt(p *PullRequestView, b bookmarkCategory, line int) (int, interface{}) {
	for n, l := range p.bookmarks[b] {
		if l.line == line {
			return n, l.data
		}
	}
	return 0, nil
}

func currentBookmark2(p *PullRequestView, b bookmarkCategory) (int, interface{}) {
	bookmarks := p.bookmarks[b]
	for n, l := range bookmarks {
		content, _ := p.getContentView()
		l1 := math.MaxInt
		if n+1 < len(bookmarks) {
			l1 = bookmarks[n+1].line
		}
		if l.line <= content.viewport.YOffset && content.viewport.YOffset < l1 {
			return n, l.data
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

type focusChangedMsg struct {
	newFocus viewAddress
}

func (p *PullRequestView) reloadPullRequest() tea.Cmd {
	if pr, err := loadPullRequestData(p.pullRequest.PullRequest); err == nil {
		p.pullRequest = pr
		p.withContentViewPtr(func(view *contentView) error {
			view.data = pr
			return nil
		})
		p.withHeaderViewPtr(func(header *pullRequestHeader) error {
			header.data = pr
			return nil
		})
		p.withFileListViewPtr(func(list *fileList) error {
			list.pullRequestData = pr
			return nil
		})
		return tea.Batch(tea.ClearScrollArea, renderPrCmd)
	} else {
		return showErrCmd(err)
	}
}

func focusChanged(address viewAddress) func() tea.Msg {
	return func() tea.Msg {
		return focusChangedMsg{address}
	}
}

type moveHorizontallyMsg struct {
	offset int
}

func moveHorizontallyCmd(offset int) tea.Cmd {
	return func() tea.Msg {
		return moveHorizontallyMsg{offset: offset}
	}
}

func (p PullRequestView) Update(m tea.Msg) (tea.Model, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	switch msg := m.(type) {
	case clearStatusMsg,
		showStatusMsg:
		if err := p.withStatusBarView(func(sb statusBar) (statusBar, error) {
			newBar, cmd := sb.Update(m)
			cmds = append(cmds, cmd)
			return newBar.(statusBar), nil
		}); err != nil {
			pterm.Warning.Println("Coudln't process a message to status bar: ", err)
		}

	case renderPrMsg:
		p.dirty = true
		p.ready = true
		p.renderPullRequest()
	case lineCommandMsg:
		switch msg.cmd {
		case newComment:
			input := confirmation.New(fmt.Sprintf("Want to comment line %05d/%05d", msg.code.old, msg.code.new), confirmation.Yes)
			if yes, err := input.RunPrompt(); yes && err == nil {
				// Do nothing for now
				if comment, err := launchEditor("",
					simpleEditor.WithWidth{pterm.GetTerminalWidth()},
					simpleEditor.WithTitle{"New Comment"},
					simpleEditor.WithPlaceholder{"Edit comment"}); err == nil && comment != "" {
					isNew := msg.code.code.New()
					fn := getFileName(msg.code.file)
					lineNum := int(msg.code.new)
					if !isNew {
						lineNum = int(msg.code.old)
					}
					if newComment, err := p.pullRequest.CreateComment(fn, msg.code.commitId, lineNum, isNew, comment); err != nil {
						pterm.Warning.Println("Couldn't add: ", err)
					} else {
						p.pullRequest.addComment(fn, msg.code.old, msg.code.new, isNew, newComment)
						return p, tea.Batch(tea.ClearScrollArea, renderPrCmd)
					}
				}

			}
			return p, tea.ClearScrollArea
		case replyComment:
			if _, data := bookmarkAt(&p, COMMENT_CATEGORY, msg.line); data != nil {
				if comment, ok := data.(sv.Comment); ok {
					input := confirmation.New(fmt.Sprintf("Want to reply to comment by %s", comment.GetUser().GetDisplayName()), confirmation.Yes)
					if yes, err := input.RunPrompt(); yes && err == nil {
						// Do nothing for now
						if replyText, err := launchEditor("",
							simpleEditor.WithWidth{pterm.GetTerminalWidth()},
							simpleEditor.WithTitle{"Reply to Comment"},
							simpleEditor.WithPlaceholder{"Edit comment"}); err == nil {
							if _, err := p.pullRequest.ReplyToComment(comment, replyText); err != nil {
								return p, showErrCmd(err)
							} else {
								return p, p.reloadPullRequest()
							}
						}
					}
					return p, tea.ClearScrollArea
				}
			} else {
				return p, showErrCmd(fmt.Errorf("No comments found at line %d", msg.line))
			}
		}

	case moveToBookmarkMsg:
		switch msg.dir {
		case NEXT:
			if err := p.moveToNextBookmark(msg.cat); err != nil {
				cmds = append(cmds, showErrCmd(err))
			}
		case PREV:
			if err := p.moveToPrevBookmark(msg.cat); err != nil {
				cmds = append(cmds, showErrCmd(err))
			}
		}
	case moveToHeadingMsg:
		switch msg.dir {
		case NEXT:
			p.moveToNextHeading(msg.level)
		case PREV:
			p.moveToPrevHeading(msg.level)
		}
	case moveHorizontallyMsg:
		if p.xOffset+msg.offset >= 0 {
			p.xOffset += msg.offset
			return p, renderPrCmd
		}
	case tea.KeyMsg:

		switch k := msg.String(); k {
		case "tab":
			p.nextFocus()
			cmds = append(cmds, focusChanged(p.currentFocus()))
		case "ctrl+c":
			return p, tea.Quit
		case "q":
			return p, tea.Quit
		case "esc":
			return p, tea.Quit
		case "R":
			if rev := p.pullRequest.pendingReview; rev == nil {
				if rev, err := p.pullRequest.StartReview(); err != nil {
					return p, showErrCmd(err)
				} else {
					p.pullRequest.pendingReview = rev
					return p, renderPrCmd
				}
			} else if yes, err := confirmation.New(fmt.Sprintf("Want to request changes for PR %s ?", rev.GetId()), confirmation.Yes).RunPrompt(); !yes || err != nil {
				return p, nil
			} else if text, err := launchEditor("",
				simpleEditor.WithWidth{pterm.GetTerminalWidth()},
				simpleEditor.WithTitle{"Request changes"},
				simpleEditor.WithPlaceholder{"Edit review comment"}); err != nil {
				return p, nil
			} else if err := rev.RequestChanges(&text); err != nil {
				return p, showErrCmd(err)
			} else {
				return p, p.reloadPullRequest()
			}
		case "C":
			if rev := p.pullRequest.pendingReview; rev != nil {
				if yes, err := confirmation.New(fmt.Sprintf("Want to cancel rev %s ?", rev.GetId()), confirmation.Yes).RunPrompt(); !yes || err != nil {
					return p, nil
				} else if err := rev.Cancel(); err != nil {
					return p, showErrCmd(err)
				} else {
					p.pullRequest.pendingReview = nil
					return p, renderPrCmd
				}
			}
		case "S":
			if rev := p.pullRequest.pendingReview; rev != nil {
				if yes, err := confirmation.New(fmt.Sprintf("Want to submit rev %s ?", rev.GetId()), confirmation.Yes).RunPrompt(); !yes || err != nil {
					return p, nil
				} else if err := rev.Close(nil); err != nil {
					return p, showErrCmd(err)
				} else {
					p.pullRequest.pendingReview = nil
					return p, p.reloadPullRequest()
				}
			}
		case "M":
			if yes, err := confirmation.New(fmt.Sprintf("Want to merge rev %v ?", p.pullRequest.GetId()), confirmation.Yes).RunPrompt(); !yes || err != nil {
				return p, nil
			} else if err := p.pullRequest.Merge(); err != nil {
				return p, showErrCmd(err)
			} else {
				return p, p.reloadPullRequest()
			}

		case "A":
			if rev := p.pullRequest.pendingReview; rev != nil {
				if text, err := launchEditor("",
					simpleEditor.WithWidth{pterm.GetTerminalWidth()},
					simpleEditor.WithTitle{"Request changes"},
					simpleEditor.WithPlaceholder{"Edit review comment"}); err != nil {
					return p, showErrCmd(err)
				} else if yes, err := confirmation.New(fmt.Sprintf("Want to approve rev %v ?", rev.GetId()), confirmation.Yes).RunPrompt(); !yes || err != nil {
					return p, nil
				} else if err := rev.Approve(&text); err != nil {
					return p, showErrCmd(err)
				} else {
					p.pullRequest.pendingReview = nil
					return p, p.reloadPullRequest()
				}
			}
		case "v":
			newMode := p.layoutMode.withFileView(!p.layoutMode.showFileView)
			if tree, err := layoutWidgets(&p.boxer, newMode); err == nil {
				p.boxer.LayoutTree = tree
				p.layoutMode = newMode
				p.ready = false
				cmds = append(cmds, tea.ClearScrollArea, RefreshSizeCmd)
			}
			if !p.isVisible(p.currentFocus()) {
				p.nextFocus()
			}

		default:
			switch p.currentFocus() {
			case CONTENT_ADDRESS:
				if err := p.withContentView(func(content contentView) (contentView, error) {
					newView, cmd := content.Update(msg)
					cmds = append(cmds, cmd)
					return newView.(contentView), nil
				}); err != nil {
					cmds = append(cmds, showErrCmd(err))
				}
			case FILEVIEW_ADDRESS:
				if err := p.withFileListView(func(view fileList) (fileList, error) {
					newView, cmd := view.Update(msg)
					cmds = append(cmds, cmd)
					return newView.(fileList), nil
				}); err != nil {
					cmds = append(cmds, showErrCmd(err))
				}
			}
		}

	case tea.WindowSizeMsg:
		newBox, cmd := p.boxer.Update(msg)
		// Update from model
		p.boxer = newBox.(boxer.Boxer)
		cmds = append(cmds, cmd, renderPrCmd)
	//p, cmds = p.propagateEvent(msg, cmds)
	case focusChangedMsg:
		p, cmds = p.propagateEvent(msg, cmds)
	case fileSelectedMsg:
		if msg.move {
			p.moveToBookmark(FILE_CATEGORY, msg.ordinal)
		}
		p, cmds = p.propagateEvent(msg, cmds)
	default:
		p.withContentViewPtr(func(content *contentView) error {
			content.viewport, cmd = content.viewport.Update(msg)
			cmds = append(cmds, cmd)
			return nil
		})

	}

	// Handle keyboard and mouse events in the viewport

	// Update last headings
	p.withHeaderViewPtr(func(header *pullRequestHeader) error {
		content, _ := p.getContentView()
		for l := 0; l < len(header.currentHeading); l++ {
			header.currentHeading[l] = -1
			for n, h := range header.headings[l] {
				if h.line <= content.viewport.YOffset+1 && (content.viewport.YOffset < h.lineEnd || h.lineEnd == -1) {
					header.currentHeading[l] = n
				}
			}
		}
		return nil
	})

	// Eventually send a file selected event
	if flv, ok := p.getFileListView(); ok {
		if ord, val := currentBookmark2(&p, FILE_CATEGORY); val != nil && ord != flv.selectedLine {
			cmds = append(cmds, fileSelected(ord, false))
		}
	}

	return p, tea.Batch(cmds...)
}

var useEditor = false

func launchEditor(initialText string, opts ...simpleEditor.Opts) (string, error) {

	if useEditor {
		if file, err := os.CreateTemp("", "comment-*.txt"); err == nil {
			if _, err = file.WriteString(initialText); err != nil {
				return "", err
			}
			file.Close()
			defer os.Remove(file.Name())

			// Run editor
			var editorPath string
			if editorPath = os.Getenv("EDITOR"); editorPath == "" {
				editorPath = "vim"
			}

			cmd := exec.Command(editorPath, file.Name())
			cmd.Stdin = os.Stdin
			cmd.Stdout = os.Stdout
			if err = cmd.Start(); err != nil {
				return "", err
			}
			if err = cmd.Wait(); err != nil {
				return "", err
			}

			// Read back the file and write it
			if content, err := os.ReadFile(file.Name()); err == nil {
				return string(content[:]), nil
			} else {
				return "", err
			}

		} else {
			return "", err
		}
	} else {
		opts = append(opts, simpleEditor.WithValue{initialText})
		return simpleEditor.RunEditor(opts...)
	}
}

func (p PullRequestView) propagateEvent(msg tea.Msg, cmds []tea.Cmd) (PullRequestView, []tea.Cmd) {
	// Recursively update the sub-widgets
	p.withFileListView(func(view fileList) (fileList, error) {
		newFile, cmd := view.Update(msg)
		view = newFile.(fileList)
		cmds = append(cmds, cmd)
		return view, nil
	})

	p.withContentView(func(view contentView) (contentView, error) {
		newFile, cmd := view.Update(msg)
		view = newFile.(contentView)
		cmds = append(cmds, cmd)
		return view, nil
	})

	p.withHeaderView(func(view pullRequestHeader) (pullRequestHeader, error) {
		newFile, cmd := view.Update(msg)
		view = newFile.(pullRequestHeader)
		cmds = append(cmds, cmd)
		return view, nil
	})

	newBox, cmd := p.boxer.Update(msg)
	p.boxer = newBox.(boxer.Boxer)
	cmds = append(cmds, cmd)
	return p, cmds
}

func (p PullRequestView) View() string {
	if !p.ready {
		return "\n  Initializing..."
	}

	return p.boxer.View()
}

type lines []string

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

type headingsLevel int

const (
	FILE_LEVEL headingsLevel = iota
	COMMIT_LEVEL
	HEADINGS
)

func getFileName(file *gitdiff.File) string {
	fn := file.OldName
	if file.IsRename {
		fn = file.NewName
	} else if file.IsNew {
		fn = file.NewName
	}

	return fn
}

func (prv *PullRequestView) renderPullRequest() {
	if !prv.dirty {
		return
	}
	prv.withContentViewPtr(func(content *contentView) error {
		return prv.withHeaderViewPtr(func(header *pullRequestHeader) error {
			h := make(lines, 0)
			content.clear()

			// clear bookmarks
			prv.bookmarks[FILE_CATEGORY] = make([]bookmark, 0)
			prv.bookmarks[COMMENT_CATEGORY] = make([]bookmark, 0)

			header.header = &h

			pr := prv.pullRequest
			sourceBranch := pr.GetBranch().GetName()
			destBranch := pr.GetBase().GetName()
			header.header.printf("#%d %s (%s)\n%s -> %s Status: %s", pr.GetId(), pr.GetTitle(), pr.GetAuthor().GetDisplayName(),
				sourceBranch, destBranch, pr.GetState())

			for n, chk := range prv.pullRequest.checks {
				header.header.printf("> %s : %s (%s)", chk.GetStatus(), chk.GetName(), chk.GetUrl())
				if n >= header.maxChecks-1 {
					break
				}
			}

			for n, rev := range prv.pullRequest.reviews {
				header.header.printf("* %s : %s (%s)", rev.GetState(), rev.GetAuthor(), rev.GetSubmitedAt())
				if n >= header.maxReviews-1 {
					break
				}
			}

			pendingReviewStyle := lipgloss.NewStyle().Width(content.viewport.Width).Background(lipgloss.Color("#00e0e0")).ColorWhitespace(true).Foreground(lipgloss.Color("#000000"))
			if perev := pr.pendingReview; perev != nil {
				header.header.printf(pendingReviewStyle.Render(fmt.Sprintf("PENDING REVIEW %s SUBMITTED AT %s (C='CANCEL', S='SUBMIT', A='APPROVE', R='REQ. CHANGES')", perev.GetId(), timefmt.Format(perev.GetSubmitedAt(), "%02d/%02m/%Y %H:%M:%S"))))
			} else {
				header.header.printf("NO PENDING REVIEW (R='Create a new one')")
			}

			prv.PrintComments(content, header, prv.pullRequest.prComments, content.viewport.Width)

			//fmt.Printf("Diff of %d files:\n\n", len(files))
			//header.printf("Diff of %d files:\n\n", len(files))

			for _, file := range prv.pullRequest.files {
				prv.addBookmark(content, FILE_CATEGORY, file)

				fn := file.OldName
				if file.IsRename {
					addHeading(content, header, content.viewport.Width, FILE_LEVEL, "%s -> %s:", file.OldName, file.NewName)
					fn = file.NewName
				} else if file.IsDelete {
					addHeading(content, header, content.viewport.Width, FILE_LEVEL, "DELETED %s", file.OldName)
					continue
				} else if file.IsNew {
					addHeading(content, header, content.viewport.Width, FILE_LEVEL, "NEW %s:", file.NewName)
					fn = file.NewName
				} else if file.IsCopy {
					addHeading(content, header, content.viewport.Width, FILE_LEVEL, "COPY %s -> %s:", file.OldName, file.NewName)
				} else {
					addHeading(content, header, content.viewport.Width, FILE_LEVEL, "%s:", file.NewName)
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
					content.printf("\nBINARY FILE\n")
				} else {
					w := content.viewport.Width
					styleAdd := lipgloss.NewStyle().
						Bold(true).
						Foreground(lipgloss.Color("#ffffff")).
						Background(lipgloss.Color("#005E00e0")).
						Width(w).MaxHeight(1)
					styleNorm := lipgloss.NewStyle().
						Foreground(lipgloss.Color("#999999")).
						Background(lipgloss.Color("#000000")).
						Width(w).MaxHeight(1)
					styleDel := lipgloss.NewStyle().
						Bold(true).
						Foreground(lipgloss.Color("#ffffff")).
						Background(lipgloss.Color("#5e0000")).
						Width(w).MaxHeight(1)

					for _, frag := range file.TextFragments {

						addHeading(content, header, content.viewport.Width, COMMIT_LEVEL, "==O== ==N== (+%d, -%d,  O=%d, N=%d)", frag.LinesAdded, frag.LinesDeleted,
							frag.OldLines, frag.NewLines)
						oldN := frag.OldPosition
						newN := frag.NewPosition

						for pos, ln := range frag.Lines {
							content.saveLine(prv.pullRequest.GetLastCommitId(), oldN, newN, pos+1, file, ln)
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

							rendered := style.Render(fillLine(pterm.RemoveColorFromString(fmt.Sprintf("%05d %05d %s  %s", oldN, newN, ln.Op, escaped)), w))

							content.printf(rendered)
							if haveFileComments {
								if commentsForLine, haveLineComments := commentsForFile[-newN]; ln.Op != gitdiff.OpDelete && haveLineComments {
									prv.PrintComments(content, header, commentsForLine, content.viewport.Width)
									delete(commentsForFile, -newN)
								}

								if commentsForLine, haveLineComments := commentsForFile[oldN]; ln.Op != gitdiff.OpAdd && haveLineComments {
									prv.PrintComments(content, header, commentsForLine, content.viewport.Width)
									delete(commentsForFile, oldN)
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
				content.updateViewportWithContent()
			}
			prv.dirty = false
			return nil
		})
	})

}

var asyncMsg chan tea.Msg

func sendAsyncMsg(msg tea.Msg) {
	asyncMsg <- msg
}

func sendAsyncCmd(cmd tea.Cmd) {
	sendAsyncMsg(cmd())
}

func ShowPr(pr sv.PullRequest) error {
	asyncMsg = make(chan tea.Msg)
	defer close(asyncMsg)

	if prv, err := NewView(pr); err != nil {
		return err
	} else {
		// Show Pr
		p := tea.NewProgram(
			prv,
			tea.WithAltScreen(),       // use the full size of the terminal in its "alternate screen buffer"
			tea.WithMouseCellMotion(), // turn on mouse support so we can track the mouse wheel
		)

		// Start processing any async msg
		go func() {
			for msg := range asyncMsg {
				p.Send(msg)
			}
		}()

		if err := p.Start(); err != nil {
			fmt.Println("could not run program:", err)
		}
		return nil
	}
}
