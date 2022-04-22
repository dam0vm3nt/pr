/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"fmt"
	"github.com/bluekeyes/go-gitdiff/gitdiff"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
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
	headings       []heading
	currentHeading int
}

func (p *pullRequestView) addHeading(format string, args ...any) {
	s := fmt.Sprintf(format, args...)
	p.content.printf(s)
	p.headings = append(p.headings, heading{len(*p.content) + 1, s})
}

func (p pullRequestView) Init() tea.Cmd {
	return nil
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
			if p.currentHeading >= 0 && p.currentHeading < len(p.headings)-1 {
				p.currentHeading += 1
				p.viewport.YOffset = p.headings[p.currentHeading].line - 1
			} else if len(p.headings) > 0 {
				p.currentHeading = 0
				p.viewport.YOffset = p.headings[p.currentHeading].line - 1
			}

		case "p":
			if len(p.headings) > 0 && p.currentHeading > 0 {
				p.currentHeading -= 1
				p.viewport.YOffset = p.headings[p.currentHeading].line - 1
			} else if len(p.headings) > 0 {
				p.currentHeading = len(p.headings) - 1
				p.viewport.YOffset = p.headings[p.currentHeading].line - 1
			}

		}

	case tea.WindowSizeMsg:
		headerHeight := len(*p.header) + 1
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
	p.currentHeading = -1
	for n, h := range p.headings {
		if h.line <= p.viewport.YOffset+1 {
			p.currentHeading = n
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
	heading := ""
	if p.currentHeading >= 0 {
		heading = p.headings[p.currentHeading].text
	}

	return fmt.Sprintf("%s\n%s\n%s", header, heading, p.viewport.View())
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
	return &pullRequestView{header: &h, content: &c, headings: make([]heading, 0)}
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
						fmt.Printf("%s -> %s:\n", file.OldName, file.NewName)
						fn = file.NewName
					} else if file.IsDelete {
						fmt.Printf("DELETED %s\n", file.OldName)
						continue
					} else if file.IsNew {
						fmt.Printf("NEW %s:\n", file.NewName)
						fn = file.NewName
					} else if file.IsCopy {
						fmt.Printf("COPY %s -> %s:\n", file.OldName, file.NewName)
					} else {
						fmt.Printf("%s:\n", file.NewName)
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

							prv.addHeading("==O== ==N== (+%d, -%d,  O=%d, N=%d)", frag.LinesAdded, frag.LinesDeleted,
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

								prv.content.printf(style.Render(fmt.Sprintf("%05d %05d %s  %s", oldN, newN, ln.Op, ln.Line)))
								if haveFileComments {

									if commentsForLine, haveLineComments := commentsForFile[newN]; haveLineComments {
										prv.PrintComments(commentsForLine)
										prv.content.printf("-------")
										delete(commentsForFile, newN)
									} else if commentsForLine, haveLineComments := commentsForFile[-oldN]; haveLineComments {
										prv.PrintComments(commentsForLine)
										prv.content.printf("-------")
										delete(commentsForFile, newN)
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

func (prv *pullRequestView) removeLastHeader() {
	prv.headings = append(prv.headings, heading{len(*prv.content), prv.headings[len(prv.headings)-2].text})
}

func (prv *pullRequestView) PrintComments(comments []sv.Comment) {
	w, _, _ := term.GetSize(int(os.Stdout.Fd()))
	/*
		style := lipgloss.NewStyle().
			Bold(true).
			// Foreground(lipgloss.Color("#FAFAFA")).
			// Background(lipgloss.Color("#7D56F4")).
			PaddingTop(2).
			PaddingLeft(4).
			PaddingRight(4)
	*/

	r, _ := glamour.NewTermRenderer(
		// detect background color and pick either the default dark or light theme
		glamour.WithAutoStyle(),
		// wrap output at specific width
		glamour.WithWordWrap(w),
		glamour.WithEmoji(),
	)

	for _, comment := range comments {
		raw := comment.GetContent().GetRaw()
		reply := ""
		if id := comment.GetParentId(); id != nil {
			reply = fmt.Sprintf(" <- %d", id)
		}

		prv.addHeading("------- [%d%s] %s at %s ------", comment.GetId(), reply, comment.GetUser().GetDisplayName(), comment.GetCreatedOn())

		rawRendered, _ := r.Render(raw)
		prv.content.printf(rawRendered)
		prv.removeLastHeader()
	}
}

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
