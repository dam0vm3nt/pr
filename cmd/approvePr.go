/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"fmt"
	"github.com/pterm/pterm"
	"log"
	"strconv"

	"github.com/spf13/cobra"
)

// approvePrCmd represents the approvePr command
var approvePrCmd = &cobra.Command{
	Use:   "approve",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			log.Fatalln("No ID supplied")
		}

		c, ctx := GetClient()

		for _, pr := range args {
			prId, _ := strconv.ParseInt(pr, 10, 32)
			prId32 := int32(prId)
			pterm.DefaultSpinner.Start(fmt.Sprintf("Approving %d", prId32))

			if pr, _, err := c.PullrequestsApi.RepositoriesWorkspaceRepoSlugPullrequestsPullRequestIdApprovePost(ctx, prId32, repoSlug, account); err == nil {
				pterm.DefaultSpinner.Success(fmt.Sprintf("Approved %s", pr.State))
			} else {
				pterm.DefaultSpinner.Fail(err)
				log.Fatalln(err)
			}
		}

	},
}

func init() {
	prCmd.AddCommand(approvePrCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// approvePrCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// approvePrCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
