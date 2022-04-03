/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"bytes"
	"fmt"
	"github.com/bluekeyes/go-gitdiff/gitdiff"
	"github.com/spf13/cobra"
	"log"
	"strconv"
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

		c, ctx := GetClient()

		for _, id := range args {
			n, _ := strconv.ParseInt(id, 10, 32)
			pr, resp, err2 := c.PullrequestsApi.RepositoriesWorkspaceRepoSlugPullrequestsPullRequestIdGet(ctx, int32(n), repoSlug, account)

			if err2 != nil || resp.StatusCode != 200 {
				continue
			}

			sourceBranch := pr.Source.Branch.(map[string]interface{})["name"].(string)
			destBranch := pr.Destination.Branch.(map[string]interface{})["name"].(string)
			fmt.Printf("#%d %s (%s)\n%s -> %s\n\nDiffs:\n\n", pr.Id, pr.Title, pr.Author.DisplayName,
				sourceBranch, destBranch)

			diff, _, err4 := c.PullrequestsApi.RepositoriesWorkspaceRepoSlugPullrequestsPullRequestIdDiffGet(ctx, pr.Id, repoSlug, account)
			if err4 != nil {
				log.Fatalf("Error while getting raw diff : %s", err4)
			}
			files, _, err5 := gitdiff.Parse(bytes.NewBuffer(diff))
			if err5 != nil {
				log.Fatalf("Couldn't parse diff : %s", err5)
			}

			fmt.Printf("Diff of %d files:\n\n", len(files))
			for _, file := range files {
				if file.IsRename {
					fmt.Printf("%s -> %s:\n", file.OldName, file.NewName)
				} else if file.IsDelete {
					fmt.Printf("DELETED %s\n", file.OldName)
					continue
				} else if file.IsNew {
					fmt.Printf("NEW %s:\n", file.NewName)
				} else if file.IsCopy {
					fmt.Printf("COPY %s -> %s:\n", file.OldName, file.NewName)
				} else {
					fmt.Printf("%s:\n", file.NewName)
				}

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
	},
}

func GetPtr[T any](ptr *T, def T) T {
	if ptr == nil {
		return def
	}
	return *ptr
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
