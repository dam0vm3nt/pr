/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"github.com/ktrysmt/go-bitbucket"
	"github.com/spf13/cobra"
	"log"
)

// listCmd represents the list command
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all PRs",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		c, errc := GetClient()
		if errc != nil {
			log.Fatalf("While getting the client %s", errc)
		}

		po := &bitbucket.PullRequestsOptions{Owner: "latchMaster", RepoSlug: "latch-cortex"}
		prs, err := c.Repositories.PullRequests.Gets(po)
		if err != nil {
			log.Fatalf("Something has occurred : %s", err)
		}
		log.Printf("Found %s\n", prs)
	},
}

func init() {
	prCmd.AddCommand(listCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// listCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// listCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
