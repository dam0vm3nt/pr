/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"errors"
	"fmt"
	"github.com/ktrysmt/go-bitbucket"
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

func GetClient() (*bitbucket.Client, error) {
	var c *bitbucket.Client

	if len(*_token) != 0 {
		c = bitbucket.NewOAuthbearerToken(*_token)
	} else if len(*_username) != 0 && len(*_password) != 0 {
		c = bitbucket.NewBasicAuth(*_username, *_password)
	} else {
		return nil, errors.New(fmt.Sprintf("No credentials given, cannot auth"))
	}

	return c, nil
}

func init() {
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	// rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.gobb-cli.yaml)")
	_token = rootCmd.PersistentFlags().StringP("token", "t", "", "Token")
	_username = rootCmd.PersistentFlags().StringP("username", "u", "", "Username")
	_password = rootCmd.PersistentFlags().StringP("password", "p", "", "Password")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.

}
