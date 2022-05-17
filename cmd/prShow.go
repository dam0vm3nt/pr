/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"errors"
	"fmt"
	"github.com/bluekeyes/go-gitdiff/gitdiff"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
	"github.com/erikgeiser/promptkit/confirmation"
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
	"github.com/vballestra/gobb-cli/sv"
	"golang.org/x/term"
	"log"
	"os"
	"strings"
)

type heading struct {
	line int
	text string
}

type pullRequestView struct {
	header         *lines
	content        *lines
	viewport       viewport.Model
	ready          bool
	headings       [][]heading
	currentHeading []int
	bookmarks      map[string][]int
	bookmarksData  map[int]interface{}
}

func (p *pullRequestView) addHeading(lev int, format string, args ...any) {
	w, _, _ := term.GetSize(int(os.Stdout.Fd()))
	var st lipgloss.Style
	switch lev {
	case FILE_LEVEL:
		st = lipgloss.NewStyle().Background(lipgloss.Color("#d040d0")).Foreground(lipgloss.Color("#ffffff")).Bold(true).Width(w)
	default:
		st = lipgloss.NewStyle().Background(lipgloss.Color("#909090")).Foreground(lipgloss.Color("#ffffff")).Width(w)
	}

	s := st.Render(fmt.Sprintf(format, args...))
	p.content.printf(s)
	p.headings[lev] = append(p.headings[lev], heading{len(*p.content) + 1, s})
}

func (p pullRequestView) Init() tea.Cmd {
	return nil
}

func (p *pullRequestView) MoveToNextHeading(L int) {
	if p.currentHeading[L] >= 0 && p.currentHeading[L] < len(p.headings[L])-1 {
		p.currentHeading[L] += 1
		p.viewport.YOffset = p.headings[L][p.currentHeading[L]].line - 1
	} else if len(p.headings[L]) > 0 {
		p.currentHeading[L] = 0
		p.viewport.YOffset = p.headings[L][p.currentHeading[L]].line - 1
	}
}

func (p *pullRequestView) MoveToPrevHeading(L int) {
	if len(p.headings[L]) > 0 && p.currentHeading[L] > 0 {
		p.currentHeading[L] -= 1
		p.viewport.YOffset = p.headings[L][p.currentHeading[L]].line - 1
	} else if len(p.headings[L]) > 0 {
		p.currentHeading[L] = len(p.headings[L]) - 1
		p.viewport.YOffset = p.headings[L][p.currentHeading[L]].line - 1
	}
}

func (p *pullRequestView) MoveToNextBookmark(b string) error {
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

func (p *pullRequestView) MoveToPrevBookmark(b string) error {
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

func CurrentBookmark(p *pullRequestView, b string) (int, interface{}) {
	for n, l := range p.bookmarks[b] {
		if l == p.viewport.YOffset {
			return n, p.bookmarksData[n]
		}
	}
	return 0, nil
}

func (p pullRequestView) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch k := msg.String(); k {
		case "ctrl+c":
			return p, tea.Quit
		case "q":
			return p, tea.Quit
		case "esc":
			return p, tea.Quit
		case "n":
			p.MoveToNextHeading(COMMIT_LEVEL)
		case "p":
			p.MoveToPrevHeading(COMMIT_LEVEL)
		case "N":
			p.MoveToNextHeading(FILE_LEVEL)
		case "P":
			p.MoveToPrevHeading(FILE_LEVEL)
		case "c":
			p.MoveToNextBookmark("COMMENT")
		case "C":
			p.MoveToPrevBookmark("COMMENT")
		case "r":
			if _, data := CurrentBookmark(&p, "COMMENT"); data != nil {
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
			// Render the viewport one line below the header.
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

func (p pullRequestView) View() string {
	if !p.ready {
		return "\n  Initializing..."
	}
	header := strings.Join(*p.header, "\n")
	// Find the first header
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

func (prv *lines) printf(format string, args ...any) {
	res := fmt.Sprintf(format, args...)
	ln := strings.Split(res, "\n")
	if len(ln) > 1 && len(ln[len(ln)-1]) == 0 {
		ln = ln[:len(ln)-1]
	}
	*prv = append(*prv, ln...)
}

func newView() *pullRequestView {

	h := make(lines, 0)
	c := make(lines, 0)

	headings := make([][]heading, HEADINGS)
	for l := 0; l < HEADINGS; l++ {
		headings[l] = make([]heading, 0)
	}

	return &pullRequestView{
		header:         &h,
		content:        &c,
		headings:       headings,
		currentHeading: make([]int, HEADINGS),
		bookmarks: map[string][]int{
			"COMMENT": make([]int, 0),
		},
		bookmarksData: make(map[int]interface{})}
}

// prShowCmd represents the prShow command
var prShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Shows one PR details",
	Long:  `Shows one PR with all details`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			log.Fatalln("No ID supplied")
		}

		sv := GetSv()

		for _, id := range args {

			prv := newView()

			if pr, err := sv.GetPullRequest(id); err != nil {
				pterm.Error.Println(err)
			} else {
				sourceBranch := pr.GetBranch().GetName()
				destBranch := pr.GetBase().GetName()
				prv.header.printf("#%d %s (%s)\n%s -> %s\tStatus: %s", pr.GetId(), pr.GetTitle(), pr.GetAuthor().GetDisplayName(),
					sourceBranch, destBranch, pr.GetState())

				if checks, err := pr.GetChecks(); err == nil {
					for _, chk := range checks {
						prv.header.printf("* %s : %s (%s)", chk.GetStatus(), chk.GetName(), chk.GetUrl())
					}
				} else {
					pterm.Warning.Println("Couldn't read the checks ", err)
				}

				if reviews, err := pr.GetReviews(); err == nil {
					for _, rev := range reviews {
						prv.header.printf("* %s : %s (%s)", rev.GetState(), rev.GetAuthor(), rev.GetSubmitedAt())
					}
				} else {
					pterm.Warning.Println("Couldn't read the checks ", err)
				}

				prComments, commentMap, err2 := pr.GetCommentsByLine()
				if err2 != nil {
					continue
				}

				prv.PrintComments(prComments)

				files, err3 := pr.GetDiff()
				if err3 != nil {
					continue
				}

				//fmt.Printf("Diff of %d files:\n\n", len(files))
				//prv.header.printf("Diff of %d files:\n\n", len(files))

				for _, file := range files {
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

					commentsForFile, haveFileComments := commentMap[fn]

					if file.IsBinary {
						prv.content.printf("\nBINARY FILE\n")
					} else {
						w, _, _ := term.GetSize(int(os.Stdout.Fd()))
						styleAdd := lipgloss.NewStyle().
							Bold(true).
							Inline(true).
							Foreground(lipgloss.Color("#000000")).
							Background(lipgloss.Color("#007E00e0")).
							Width(w)
						styleNorm := lipgloss.NewStyle()
						styleDel := lipgloss.NewStyle().
							Bold(true).
							Inline(true).
							Foreground(lipgloss.Color("#ffffff")).
							Background(lipgloss.Color("#7e0000e0")).
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

			}

			// Show Pr
			p := tea.NewProgram(
				prv,
				tea.WithAltScreen(),       // use the full size of the terminal in its "alternate screen buffer"
				tea.WithMouseCellMotion(), // turn on mouse support so we can track the mouse wheel
			)

			if err := p.Start(); err != nil {
				fmt.Println("could not run program:", err)
				os.Exit(1)
			}
		}
	},
}

func (prv *pullRequestView) removeLastHeader(lev int) {
	prv.headings[lev] = append(prv.headings[lev], heading{len(*prv.content), prv.headings[lev][len(prv.headings[lev])-2].text})
}

func ptr[T string | uint | int](s T) *T {
	return &s
}

func (prv *pullRequestView) PrintComments(comments []sv.Comment) {
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

		prv.AddBookmark("COMMENT", comment)
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

func (prv *pullRequestView) AddBookmark(b string, data interface{}) {
	prv.bookmarksData[len(prv.bookmarks[b])] = data
	prv.bookmarks[b] = append(prv.bookmarks[b], len(*prv.content)+1)
}

const (
	FILE_LEVEL = iota
	COMMIT_LEVEL
	HEADINGS
)

func init() {
	prCmd.AddCommand(prShowCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// prShowCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// prShowCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

}
