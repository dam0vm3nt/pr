/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"github.com/charmbracelet/lipgloss"
	"github.com/pterm/pterm"
	"github.com/vballestra/sv/cmd/ui/statusView"
	"strings"

	"github.com/spf13/cobra"
)

// prStatusCmd represents the prStatus command
var prStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Get status of your PRs",
	Long:  `Get status of your PRs`,
	Run: func(cmd *cobra.Command, args []string) {
		sv := GetSv()

		if c, err := sv.PullRequestStatus(); err != nil {
			pterm.Fatal.Println(err)
			return
		} else {
			if interactive {
				if err := statusView.RunPrStatusView(sv); err != nil {
					pterm.Fatal.Println(err)
				}
				return
			}

			data := pterm.TableData{{"ID", "Title", "Author", "Repository", "Branch", "State", "Reviews", "Checks", "Contexts"}}
			mineStyle := lipgloss.NewStyle().Bold(true).ColorWhitespace(true).Foreground(lipgloss.Color("#ff0000"))
			theirStyle := lipgloss.NewStyle().Bold(true).ColorWhitespace(true).Foreground(lipgloss.Color("#00ffff"))

			for pr := range c {
				checks := make([]string, 0)
				for s, k := range pr.GetChecksByStatus() {
					if k > 0 {
						checks = append(checks, fmt.Sprintf("%s: %d", s, k))
					}
				}

				contexts := make([]string, 0)
				for s, k := range pr.GetContextByStatus() {
					if k > 0 {
						contexts = append(contexts, fmt.Sprintf("%s: %d", s, k))
					}
				}

				reviewCount := make(map[string]map[string]int)
				for _, r := range pr.GetReviews() {
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
					if verboseStatus {
						byAuthStats := make([]string, 0)
						for a, c := range byStatus {
							byAuthStats = append(byAuthStats, fmt.Sprintf("%s(%d)", a, c))
						}
						stats = strings.Join(byAuthStats, "|")
					} else {
						stats = fmt.Sprintf("%d", len(byStatus))
					}
					reviews = append(reviews, fmt.Sprintf("%s: %s", s, stats))
				}

				var idStr string

				if pr.IsMine() {
					idStr = mineStyle.Render(
						fmt.Sprintf("%5d", pr.GetId()))
				} else {
					idStr = theirStyle.Render(
						fmt.Sprintf("%5d", pr.GetId()))
				}

				data = append(data, []string{
					idStr, pr.GetTitle(), pr.GetAuthor(), pr.GetRepository(), pr.GetBranchName(), pr.GetStatus(), strings.Join(reviews, ", "), strings.Join(checks, ", "),
					strings.Join(contexts, ", "),
				})
			}

			rerr := pterm.DefaultTable.WithHasHeader().WithData(data).Render()
			if rerr != nil {
				pterm.Fatal.Println(rerr)
			}
		}
	},
}

var verboseStatus = false
var interactive = false

func init() {
	prCmd.AddCommand(prStatusCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// prStatusCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	prStatusCmd.Flags().BoolVarP(&verboseStatus, "verbose", "v", false, "Shows a more verbose status")
	prStatusCmd.Flags().BoolVarP(&interactive, "interactive", "i", false, "Interactively shows prs")
}
