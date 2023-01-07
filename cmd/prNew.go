/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"github.com/antihax/optional"
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
	"github.com/vballestra/sv/sv"
)

func optionalString(s string) optional.String {
	if s == "" {
		return optional.EmptyString()
	} else {
		return optional.NewString(s)
	}
}

// prNewCmd represents the prNew command
var prNewCmd = &cobra.Command{
	Use:   "new",
	Short: "Create a new PR",
	Long:  `Create a new PR`,
	Run: func(cmd *cobra.Command, args []string) {
		sv2 := GetSv()

		a := sv.CreatePullRequestArgs{
			BaseBranch:          optionalString(newPrBaseBranch),
			HeadBranch:          optionalString(newPrDestBranch),
			Title:               optionalString(newPrTitle),
			Description:         optionalString(newPrDescription),
			Labels:              newPrLabels,
			Reviewers:           newPrReviwers,
			CreateMissingLabels: true,
		}

		if pr, err := sv2.CreatePullRequest(a); err != nil {
			pterm.Fatal.Printfln("couldn't create a new pr : %v", err)
		} else {
			pterm.Info.Printfln("New Pr created.\nId: %v\nStatus: %v", pr.GetId(), pr.GetStatus())
		}
	},
}

var newPrTitle = ""
var newPrDescription = ""

var newPrBaseBranch = ""

var newPrDestBranch = ""

var newPrReviwers = make([]string, 0)

var newPrLabels = make([]string, 0)

func init() {
	prCmd.AddCommand(prNewCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// prStatusCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	prNewCmd.Flags().StringVarP(&newPrTitle, "title", "T", "", "The new PR title")
	prNewCmd.Flags().StringVarP(&newPrDescription, "description", "d", "", "New PR optional description")
	prNewCmd.Flags().StringVarP(&newPrBaseBranch, "base", "b", "", "The base branch (required)")
	prNewCmd.Flags().StringVarP(&newPrDestBranch, "head", "H", "", "The (optional) head branch")
	prNewCmd.Flags().StringSliceVarP(&newPrReviwers, "reviewer", "R", []string{}, "Optional list of reviewers requested")
	prNewCmd.Flags().StringSliceVarP(&newPrLabels, "label", "l", []string{}, "Optional list of labels")
}
