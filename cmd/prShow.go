/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"fmt"
	"github.com/bluekeyes/go-gitdiff/gitdiff"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
	"github.com/vballestra/gobb-cli/sv"
	"log"
)

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

			if pr, err := sv.GetPullRequest(id); err != nil {
				pterm.Error.Println(err)
			} else {
				sourceBranch := pr.GetBranch().GetName()
				destBranch := pr.GetBase().GetName()
				fmt.Printf("#%d %s (%s)\n%s -> %s\n\nDiffs:\n\n", pr.GetId(), pr.GetTitle(), pr.GetAuthor().GetDisplayName(),
					sourceBranch, destBranch)

				prComments, commentMap, err2 := pr.GetCommentsByLine()
				if err2 != nil {
					continue
				}

				PrintComments(prComments)

				files, err3 := pr.GetDiff()
				if err3 != nil {
					continue
				}

				fmt.Printf("Diff of %d files:\n\n", len(files))

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
						fmt.Printf("\nBINARY FILE\n")
					} else {
						for _, frag := range file.TextFragments {

							fmt.Printf("==O== ==N== (+%d, -%d,  O=%d, N=%d)\n", frag.LinesAdded, frag.LinesDeleted,
								frag.OldLines, frag.NewLines)
							oldN := frag.OldPosition
							newN := frag.NewPosition

							for _, ln := range frag.Lines {
								fmt.Printf("%05d %05d %s  %s", oldN, newN, ln.Op, ln.Line)
								if haveFileComments {
									if commentsForLine, haveLineComments := commentsForFile[newN]; haveLineComments {
										PrintComments(commentsForLine)
										pterm.Println("-------")

										// Remove the comment as it shouldn't be printed anymore
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

		}
	},
}

func PrintComments(comments []sv.Comment) {
	style := lipgloss.NewStyle().
		Bold(true).
		// Foreground(lipgloss.Color("#FAFAFA")).
		// Background(lipgloss.Color("#7D56F4")).
		PaddingTop(2).
		PaddingLeft(4).
		PaddingRight(4)
	r, _ := glamour.NewTermRenderer(
		// detect background color and pick either the default dark or light theme
		glamour.WithAutoStyle(),
		// wrap output at specific width
		glamour.WithWordWrap(200),
	)

	for _, comment := range comments {
		raw := comment.GetContent().GetRaw()
		reply := ""
		if id := comment.GetParentId(); id != nil {
			reply = fmt.Sprintf(" <- %d", id)
		}

		rawRendered, _ := r.Render(raw)
		fmt.Print(style.Render(fmt.Sprintf("------- [%d%s] %s at %s ------%s", comment.GetId(), reply, comment.GetUser().GetDisplayName(), comment.GetCreatedOn(), rawRendered)))
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
