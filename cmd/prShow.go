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
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			log.Fatalln("No ID supplied")
		}

		c, err := GetClient()
		if err != nil {
			log.Fatalf("Counldn't get client : %s\n", err)
		}

		for _, id := range args {
			n, _ := strconv.ParseInt(id, 10, 64)
			pr, resp, err2 := c.PullRequests.Get(account, repoSlug, n)

			if err2 != nil || resp.StatusCode != 200 {
				continue
			}

			fmt.Printf("#%d %s (%s)\n%s -> %s\n\nDiffs:\n\n", *pr.ID, *pr.Title, *pr.Author.DisplayName,
				*pr.Source.Branch.Name, *pr.Destination.Branch.Name)

			rawDiff, resp, err4 := c.PullRequests.GetDiffRaw(account, repoSlug, *pr.ID)
			diffStr := rawDiff.String()
			if err4 != nil {
				log.Fatalf("Error while getting raw diff : %s", err4)
			}
			files, _, err5 := gitdiff.Parse(bytes.NewBufferString(resp.Body))
			if err5 != nil {
				log.Fatalf("Couldn't parse diff : %s", err5)
			}
			rawDiff.Reset()
			log.Printf("Str: '%s'\n", diffStr)

			fmt.Printf("Diff of %d files:\n\n", len(files))
			for _, file := range files {
				fmt.Printf("%s -> %s:\n", file.OldName, file.NewName)
				for _, frag := range file.TextFragments {

					fmt.Printf("%s (+%d, -%d,  O=%d, N=%d)\n", frag.Comment, frag.LinesAdded, frag.LinesDeleted,
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
