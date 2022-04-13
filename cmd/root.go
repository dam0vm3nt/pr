/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"context"
	"github.com/vballestra/gobb-cli/bitbucket"
	"github.com/vballestra/gobb-cli/sv"
	"net/http"
	"os"

	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "gobb-cli",
	Short: "Stupid CLI for Bitbucket",
	Long: `I need a CLI for managing my BB account, so I goggled it and couldn't find any.
Since I'm a developer I thought it would have been funny to create one. Here it is.`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	// Run: func(cmd *cobra.Command, args []string) { },
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

var _token, _username, _password *string

var account, repoSlug string

var githubToken string

func GetClient() (*bitbucket.APIClient, context.Context) {
	cfg := bitbucket.NewConfiguration()
	cfg.HTTPClient = &http.Client{}
	auth := bitbucket.BasicAuth{UserName: *_username, Password: *_password}
	ctx := context.WithValue(context.Background(), bitbucket.ContextBasicAuth, auth)
	return bitbucket.NewAPIClient(cfg), ctx
}

func GetSv() sv.Sv {
	if len(githubToken) > 0 {
		return sv.NewGitHubSv(githubToken)
	} else {
		return sv.NewBitBucketSv(*_username, *_password, repoSlug, account)
	}

}

func init() {
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	// rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.gobb-cli.yaml)")
	_username = rootCmd.PersistentFlags().StringP("username", "u", "", "Username")
	_password = rootCmd.PersistentFlags().StringP("password", "p", "", "Password")
	rootCmd.PersistentFlags().StringVarP(&account, "account", "a", "latchMaster", "Account")
	rootCmd.PersistentFlags().StringVarP(&repoSlug, "repository", "r", "latch-cortex", "Repository")
	rootCmd.PersistentFlags().StringVarP(&githubToken, "token", "t", "", "Github token")
	// Cobra also supports local flags, which will only run
	// when this action is called directly.

}
