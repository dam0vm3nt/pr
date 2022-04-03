/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"encoding/json"
	"fmt"
	"github.com/antihax/optional"
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
	"github.com/vballestra/gobb-cli/bitbucket"
	"io/ioutil"
	"log"
	"net/http"
)

// listCmd represents the list command
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all PRs",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {

		c, ctx := GetClient()

		vars := bitbucket.PullrequestsApiRepositoriesWorkspaceRepoSlugPullrequestsGetOpts{
			State: optional.NewString("ACTIVE"),
		}
		prs, resp, err := c.PullrequestsApi.RepositoriesWorkspaceRepoSlugPullrequestsGet(ctx, repoSlug, account, &vars)
		if err != nil || resp.StatusCode != 200 {
			log.Fatalf("Something has occurred : %s", err)
		}

		data := pterm.TableData{{"ID", "Title", "Branch", "Author"}}

		for cont := true; cont && prs.Page <= maxPages; cont = prs.Next != "" {
			for _, pr := range prs.Values {
				branch := pr.Source.Branch.(map[string]interface{})
				data = append(data, []string{
					fmt.Sprintf("%5d", pr.Id), pr.Title, branch["name"].(string), pr.Author.DisplayName,
				})
				//fmt.Printf(" - %d : %s\n   (%s) Author: %s\n", *pr.ID, *pr.Title, *pr.Source.Branch.Name, *pr.Author.DisplayName)
			}

			if prs.Next != "" {
				req, _ := http.NewRequestWithContext(ctx, "GET", prs.Next, nil)
				if auth, ok := ctx.Value(bitbucket.ContextBasicAuth).(bitbucket.BasicAuth); ok {
					req.SetBasicAuth(auth.UserName, auth.Password)
				}

				resp2, _ := http.DefaultClient.Do(req)
				bb, _ := ioutil.ReadAll(resp2.Body)
				json.Unmarshal(bb, &prs)

			}
		}
		rerr := pterm.DefaultTable.WithHasHeader().WithData(data).Render()
		if rerr != nil {
			pterm.Fatal.Println(rerr)
		}

	},
}

var maxPages int32

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
}
