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
	"github.com/vballestra/gobb-cli/sv"
)

// listBranchesCmd represents the listBranches command
var listBranchesCmd = &cobra.Command{
	Use:     "list",
	Short:   "List all the branches",
	Long:    `List of all active branches`,
	Aliases: []string{"ls"},
	Run: func(cmd *cobra.Command, args []string) {
		c, ctx := GetClient()
		refs, _, err := c.RefsApi.RepositoriesWorkspaceRepoSlugRefsBranchesGet(ctx, repoSlug, account, &bitbucket.RefsApiRepositoriesWorkspaceRepoSlugRefsBranchesGetOpts{
			Q: optional.NewString(fmt.Sprintf(`name ~ "%s"`, branchesCmdNameFilter)),
		})
		if err != nil {
			pterm.Fatal.Println(err)
		}
		td := pterm.TableData{[]string{"Name"}}

		var pager sv.Paginated[bitbucket.Branch, bitbucket.PaginatedBranches]
		pager = PaginatedBranches{&refs}

		items := sv.Paginate(ctx, pager)
		for itm := range items {
			td = append(td, []string{itm.Name})
		}
		for _, ref := range refs.Values {
			td = append(td, []string{ref.Name})
		}
		if err2 := pterm.DefaultTable.WithHasHeader(true).WithData(td).Render(); err2 != nil {
			pterm.Fatal.Println(err2)
		}
	},
}

type PaginatedBranches struct {
	*bitbucket.PaginatedBranches
}

func (branches PaginatedBranches) GetContainer() *bitbucket.PaginatedBranches {
	return branches.PaginatedBranches
}

func (branches PaginatedBranches) GetNext() string {
	return branches.Next
}

func (branches PaginatedBranches) GetValues() []bitbucket.Branch {
	return branches.Values
}

func (branches PaginatedBranches) GetPages() int32 {
	return branches.Size
}

var branchesCmdNameFilter string

func init() {
	branchesCmd.AddCommand(listBranchesCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// listBranchesCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// listBranchesCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	listBranchesCmd.Flags().StringVarP(&branchesCmdNameFilter, "text", "t", "", "Filter by text")
}
