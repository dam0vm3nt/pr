/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"fmt"
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
	"log"
)

// listCmd represents the list command
var listCmd = &cobra.Command{
	Use:     "list",
	Short:   "List all PRs",
	Long:    `List all PRs`,
	Aliases: []string{"ls"},
	Run: func(cmd *cobra.Command, args []string) {

		c := GetSv()

		if prs, err := c.ListPullRequests(prsQuery); err != nil {
			log.Fatalf("Something has occurred : %s", err)
		} else {

			data := pterm.TableData{{"ID", "Title", "Branch", "Author", "State", "Created At"}}

			for pr := range prs {
				branch := pr.GetBranch()
				data = append(data, []string{
					fmt.Sprintf("%5d", pr.GetId()), pr.GetTitle(), branch.GetName(), fmt.Sprintf("%s", pr.GetAuthor().GetDisplayName()), pr.GetState(), pr.GetCreatedOn().String(),
				})
			}

			rerr := pterm.DefaultTable.WithHasHeader().WithData(data).Render()
			if rerr != nil {
				pterm.Fatal.Println(rerr)
			}
		}

	},
}

var maxPages int32
var prsQuery string

func init() {
	prCmd.AddCommand(listCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// listCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// listCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	listCmd.Flags().Int32Var(&maxPages, "max-pages", 5, "Max pages to retrieve")
	listCmd.Flags().StringVarP(&prsQuery, "query", "q", "", "Query")
}
