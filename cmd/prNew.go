/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
)

// prNewCmd represents the prNew command
var prNewCmd = &cobra.Command{
	Use:   "new",
	Short: "Create a new PR",
	Long:  `Create a new PR`,
	Run: func(cmd *cobra.Command, args []string) {
		sv := GetSv()

		if newPrDestBranch == "" {
			if currentBranch, err := sv.GetCurrentBranch(); err != nil {
				pterm.Fatal.Printfln("no head branch was provided or could be inferred: %v", err)
			} else {
				newPrDestBranch = currentBranch
			}
		}

		var newPrDescriptionOpt *string
		if newPrDescription == "" {
			newPrDescriptionOpt = nil
		} else {
			newPrDescriptionOpt = &newPrDescription
		}

		if pr, err := sv.CreatePullRequest(newPrBaseBranch, newPrDestBranch, newPrTitle, newPrDescriptionOpt, newPrLabels, newPrReviwers); err != nil {
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
