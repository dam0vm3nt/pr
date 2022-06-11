/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
	"github.com/vballestra/sv/cmd/ui"
	"log"
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

		sv := GetSv()

		if forcePrCheck {
			if err := sv.Fetch(); err != nil {
				pterm.Warning.Println("An issue occurred while fetching the repository: ", err)
			}
		}

		for _, id := range args {
			if pr, err := sv.GetPullRequest(id); err != nil {
				pterm.Error.Println(err)
			} else if err = ui.ShowPr(pr); err != nil {
				pterm.Warning.Println("Cannot render pr ", pr.GetId(), " because ", err)
			}
		}
	},
}

var forcePrCheck bool

func init() {
	prCmd.AddCommand(prShowCmd)

	// Here you will define your flags and configuration settings.

	prCmd.PersistentFlags().BoolVarP(&forcePrCheck, "fetch", "f", false, "Force repository fetch")
	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// prShowCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// prShowCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

}
