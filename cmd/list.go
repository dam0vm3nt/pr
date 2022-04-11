/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"fmt"
	"github.com/antihax/optional"
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
	"github.com/vballestra/gobb-cli/bitbucket"
	"log"
)

// listCmd represents the list command
var listCmd = &cobra.Command{
	Use:     "list",
	Short:   "List all PRs",
	Long:    `List all PRs`,
	Aliases: []string{"ls"},
	Run: func(cmd *cobra.Command, args []string) {

		c, ctx := GetClient()

		vars := bitbucket.PullrequestsApiRepositoriesWorkspaceRepoSlugPullrequestsGetOpts{
			State: optional.NewString("ACTIVE"),
		}
		if len(prsQuery) > 0 {
			vars.Q = optional.NewString(prsQuery)
		}
		prs, resp, err := c.PullrequestsApi.RepositoriesWorkspaceRepoSlugPullrequestsGet(ctx, repoSlug, account, &vars)
		if err != nil || resp.StatusCode != 200 {
			log.Fatalf("Something has occurred : %s", err)
		}

		data := pterm.TableData{{"ID", "Title", "Branch", "Author", "State", "Created At"}}

		for pr := range Paginate[bitbucket.Pullrequest, bitbucket.PaginatedPullrequests](ctx, PaginatedPullrequests{&prs}) {
			branch := pr.Source.Branch.(map[string]interface{})
			data = append(data, []string{
				fmt.Sprintf("%5d", pr.Id), pr.Title, branch["name"].(string), fmt.Sprintf("%s", pr.Author.DisplayName), pr.State, pr.CreatedOn.String(),
			})
		}

		rerr := pterm.DefaultTable.WithHasHeader().WithData(data).Render()
		if rerr != nil {
			pterm.Fatal.Println(rerr)
		}

	},
}

type PaginatedPullrequests struct {
	*bitbucket.PaginatedPullrequests
}

func (p PaginatedPullrequests) GetContainer() *bitbucket.PaginatedPullrequests {
	return p.PaginatedPullrequests
}

func (p PaginatedPullrequests) GetNext() string {
	return p.Next
}

func (p PaginatedPullrequests) GetPages() int32 {
	return p.Size
}

func (p PaginatedPullrequests) GetValues() []bitbucket.Pullrequest {
	return p.Values
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
