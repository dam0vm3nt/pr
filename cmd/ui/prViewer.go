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
	"github.com/vballestra/sv/sv"
	"golang.org/x/term"
	"os"
	"strings"
)

type Heading struct {
	line int
	text string
}

type PullRequestView struct {
	header         *lines
	content        *lines
	pullRequest    sv.PullRequest
	viewport       viewport.Model
	ready          bool
	headings       [][]Heading
	currentHeading []int
	bookmarks      map[string][]int
	bookmarksData  map[int]interface{}
	checks         []sv.Check
	reviews        []sv.Review
	prComments     []sv.Comment
	commentMap     map[string]map[int64][]sv.Comment
	files          []*gitdiff.File
	dirty          bool
	xOffset        int
}

func ptr[T string | uint | int](s T) *T {
	return &s
}

func (prv *PullRequestView) PrintComments(comments []sv.Comment) {
	w, _, _ := term.GetSize(int(os.Stdout.Fd()))

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
		prv.addHeading(COMMIT_LEVEL, style.Render(
			fmt.Sprintf("------- [%d%s] %s at %s ------",
				comment.GetId(), reply,
				comment.GetUser().GetDisplayName(),
				comment.GetCreatedOn())))

		rawRendered, _ := r.Render(raw)
		prv.content.printf(style2.Render(rawRendered))
		prv.removeLastHeader(COMMIT_LEVEL)
	}
}

func (prv *PullRequestView) removeLastHeader(lev int) {
	prv.headings[lev] = append(prv.headings[lev], Heading{len(*prv.content), prv.headings[lev][len(prv.headings[lev])-2].text})
}

func (prv *PullRequestView) addBookmark(b string, data interface{}) {
	prv.bookmarksData[len(prv.bookmarks[b])] = data
	prv.bookmarks[b] = append(prv.bookmarks[b], len(*prv.content)+1)
}

func (p *PullRequestView) addHeading(lev int, format string, args ...any) {
	w, _, _ := term.GetSize(int(os.Stdout.Fd()))
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
	p.headings[lev] = append(p.headings[lev], Heading{len(*p.content) + 1, s})
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
	if p.currentHeading[L] >= 0 && p.currentHeading[L] < len(p.headings[L])-1 {
		p.currentHeading[L] += 1
		p.viewport.YOffset = p.headings[L][p.currentHeading[L]].line - 1
	} else if len(p.headings[L]) > 0 {
		p.currentHeading[L] = 0
		p.viewport.YOffset = p.headings[L][p.currentHeading[L]].line - 1
	}
}

func (p *PullRequestView) moveToPrevHeading(L int) {
	if len(p.headings[L]) > 0 && p.currentHeading[L] > 0 {
		p.currentHeading[L] -= 1
		p.viewport.YOffset = p.headings[L][p.currentHeading[L]].line - 1
	} else if len(p.headings[L]) > 0 {
		p.currentHeading[L] = len(p.headings[L]) - 1
		p.viewport.YOffset = p.headings[L][p.currentHeading[L]].line - 1
	}
}

func (p *PullRequestView) moveToNextBookmark(b string) error {
	bookmarks := p.bookmarks[b]
	if len(bookmarks) == 0 {
		return errors.New("No bookmarks")
	}
	for _, l := range bookmarks {
		if l > p.viewport.YOffset {
			p.viewport.YOffset = l
			return nil
		}
	}
	p.viewport.YOffset = bookmarks[0]
	return nil
}

func (p *PullRequestView) moveToPrevBookmark(b string) error {
	bookmarks := p.bookmarks[b]
	if len(bookmarks) == 0 {
		return errors.New("No bookmarks")
	}
	for i := len(bookmarks) - 1; i >= 0; i-- {
		l := bookmarks[i]
		if l < p.viewport.YOffset {
			p.viewport.YOffset = l
			return nil
		}
	}
	p.viewport.YOffset = bookmarks[len(bookmarks)-1]
	return nil
}

func currentBookmark(p *PullRequestView, b string) (int, interface{}) {
	for n, l := range p.bookmarks[b] {
		if l == p.viewport.YOffset {
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
		headerHeight := len(*p.header) + len(p.currentHeading)
		footerHeight := 0
		verticalMarginHeight := headerHeight + footerHeight

		const useHighPerformanceRenderer = false
		if !p.ready {
			// Since this program is using the full size of the viewport we
			// need to wait until we've received the window dimensions before
			// we can initialize the viewport. The initial dimensions come in
			// quickly, though asynchronously, which is why we wait for them
			// here.
			p.viewport = viewport.New(msg.Width, msg.Height-verticalMarginHeight)
			p.viewport.YPosition = headerHeight
			p.viewport.HighPerformanceRendering = useHighPerformanceRenderer
			p.viewport.SetContent(strings.Join(*p.content, "\n"))
			p.ready = true

			// This is only necessary for high performance rendering, which in
			// most cases you won't need.
			//
			// Render the viewport one line below the Header.
			p.viewport.YPosition = headerHeight + 1
		} else {
			p.viewport.Width = msg.Width
			p.viewport.Height = msg.Height - verticalMarginHeight
		}

		if useHighPerformanceRenderer {
			// Render (or re-render) the whole viewport. Necessary both to
			// initialize the viewport and when the window is resized.
			//
			// This is needed for high-performance rendering only.
			cmds = append(cmds, viewport.Sync(p.viewport))
		}
	}

	// Handle keyboard and mouse events in the viewport
	p.viewport, cmd = p.viewport.Update(msg)

	// Update last headings
	for l := 0; l < len(p.currentHeading); l++ {
		p.currentHeading[l] = -1
		for n, h := range p.headings[l] {
			if h.line <= p.viewport.YOffset+1 {
				p.currentHeading[l] = n
			}
		}
	}

	cmds = append(cmds, cmd)

	return p, tea.Batch(cmds...)
}

func (p PullRequestView) View() string {
	if !p.ready {
		return "\n  Initializing..."
	}

	header := strings.Join(*p.header, "\n")
	// Find the first Header
	heading := make([]string, len(p.currentHeading))
	for l := 0; l < len(p.currentHeading); l++ {
		if p.currentHeading[l] >= 0 {
			heading[l] = p.headings[l][p.currentHeading[l]].text
		} else {
			heading[l] = ""
		}
	}

	return fmt.Sprintf("%s\n%s\n%s", header, strings.Join(heading, "\n"), p.viewport.View())
}

type lines []string

func NewView(pr sv.PullRequest) (*PullRequestView, error) {

	headings := make([][]Heading, HEADINGS)
	for l := 0; l < HEADINGS; l++ {
		headings[l] = make([]Heading, 0)
	}

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
		prv := &PullRequestView{
			pullRequest:    pr,
			checks:         checks,
			reviews:        reviews,
			prComments:     prComments,
			commentMap:     commentMap,
			files:          files,
			headings:       headings,
			currentHeading: make([]int, HEADINGS),
			bookmarks: map[string][]int{
				"COMMENT": make([]int, 0),
			},
			bookmarksData: make(map[int]interface{}),
			dirty:         true,
			xOffset:       0}
		return prv, nil
	}
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
	c := make(lines, 0)
	prv.content = &c
	prv.header = &h

	pr := prv.pullRequest
	sourceBranch := pr.GetBranch().GetName()
	destBranch := pr.GetBase().GetName()
	prv.header.printf("#%d %s (%s)\n%s -> %s\tStatus: %s", pr.GetId(), pr.GetTitle(), pr.GetAuthor().GetDisplayName(),
		sourceBranch, destBranch, pr.GetState())

	for _, chk := range prv.checks {
		prv.header.printf("* %s : %s (%s)", chk.GetStatus(), chk.GetName(), chk.GetUrl())
	}

	for _, rev := range prv.reviews {
		prv.header.printf("* %s : %s (%s)", rev.GetState(), rev.GetAuthor(), rev.GetSubmitedAt())
	}

	prv.PrintComments(prv.prComments)

	//fmt.Printf("Diff of %d files:\n\n", len(files))
	//prv.header.printf("Diff of %d files:\n\n", len(files))

	for _, file := range prv.files {
		fn := file.OldName
		if file.IsRename {
			prv.addHeading(FILE_LEVEL, "%s -> %s:", file.OldName, file.NewName)
			fn = file.NewName
		} else if file.IsDelete {
			prv.addHeading(FILE_LEVEL, "DELETED %s", file.OldName)
			continue
		} else if file.IsNew {
			prv.addHeading(FILE_LEVEL, "NEW %s:", file.NewName)
			fn = file.NewName
		} else if file.IsCopy {
			prv.addHeading(FILE_LEVEL, "COPY %s -> %s:", file.OldName, file.NewName)
		} else {
			prv.addHeading(FILE_LEVEL, "%s:", file.NewName)
		}

		commentsForFileOrig, haveFileComments := prv.commentMap[fn]

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
			w, _, _ := term.GetSize(int(os.Stdout.Fd()))
			styleAdd := lipgloss.NewStyle().
				Bold(true).
				Inline(true).
				Foreground(lipgloss.Color("#ffffff")).
				Background(lipgloss.Color("#005E00e0")).
				Width(w)
			styleNorm := lipgloss.NewStyle()
			styleDel := lipgloss.NewStyle().
				Bold(true).
				Inline(true).
				Foreground(lipgloss.Color("#ffffff")).
				Background(lipgloss.Color("#5e0000e0")).
				Width(w)

			for _, frag := range file.TextFragments {

				prv.addHeading(COMMIT_LEVEL, "==O== ==N== (+%d, -%d,  O=%d, N=%d)", frag.LinesAdded, frag.LinesDeleted,
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
					prv.content.printf(style.Render(fmt.Sprintf("%05d %05d %s  %s", oldN, newN, ln.Op, escaped)))
					if haveFileComments {
						if commentsForLine, haveLineComments := commentsForFile[newN]; haveLineComments {
							prv.PrintComments(commentsForLine)
							delete(commentsForFile, newN)
						} else if commentsForLine, haveLineComments := commentsForFile[-oldN]; haveLineComments {
							prv.PrintComments(commentsForLine)
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
		prv.viewport.SetContent(strings.Join(*prv.content, "\n"))
	}
	prv.dirty = false
}
