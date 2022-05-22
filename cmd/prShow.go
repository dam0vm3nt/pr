/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
	"github.com/vballestra/sv/cmd/ui"
	"log"
	"os"
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

		for _, id := range args {

			if pr, err := sv.GetPullRequest(id); err != nil {
				pterm.Error.Println(err)
			} else if prv, err := ui.NewView(pr); err != nil {
				pterm.Warning.Println("Cannot render pr ", pr.GetId(), " because ", err)
				continue
			} else {
				// Show Pr
				p := tea.NewProgram(
					prv,
					tea.WithAltScreen(),       // use the full size of the terminal in its "alternate screen buffer"
					tea.WithMouseCellMotion(), // turn on mouse support so we can track the mouse wheel
				)

				if err := p.Start(); err != nil {
					fmt.Println("could not run program:", err)
					os.Exit(1)
				}
			}
		}
	},
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
