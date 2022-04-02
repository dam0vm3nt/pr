/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"log"

	"github.com/ktrysmt/go-bitbucket"
)

var authCmd = &cobra.Command{
	Use:   "auth",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		c, errc := GetClient()
		if errc != nil {
			log.Fatalf("While getting the client %s", errc)
		}

		lo := &bitbucket.RepositoriesOptions{Role: "member", Owner: "latchMaster"}
		repos, err := c.Repositories.ListForAccount(lo)
		if err != nil {
			fmt.Printf("Error occurred : %s\n", err)
			return
		}

		fmt.Printf("Found : %d\n", repos.Size)
		for i := int32(0); i < repos.Size; i++ {
			repo := repos.Items[i]
			fmt.Printf(" - %s\n", repo.Name)
		}
	},
}

func init() {
	rootCmd.AddCommand(authCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// authCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// authCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
