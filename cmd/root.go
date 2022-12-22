/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"context"
	"fmt"
	"github.com/go-git/go-git/v5"
	"github.com/pterm/pterm"
	"github.com/vballestra/sv/bitbucket"
	"github.com/vballestra/sv/sv"
	"net/http"
	"os"
	"regexp"

	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "sv",
	Short: "Stupid CLI for any sv system",
	Long: `I need a CLI for managing my repos regardless the provider (BB, GH, GL) account, so I goggled it and couldn't find any.
Since I'm a developer I thought it would have been funny to create one. Here it is.`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	// Run: func(cmd *cobra.Command, args []string) { },
	PersistentPreRun: setupRepo,
}

var defaultOrigin = "origin"

var origins = map[OriginType]*regexp.Regexp{
	GitHubOriginType:    regexp.MustCompile("^git@github.com:(?P<org>[^/]+)/(?P<repo>.+)(\\.git)$?"),
	BitbucketOriginType: regexp.MustCompile("^git@bitbucket.org:(?P<org>[^/]+)/(?P<repo>.+)(\\.git)$"),
}

var localRepository *git.Repository
var origin *git.Remote

func setupRepo(cmd *cobra.Command, args []string) {
	var err error
	localRepository, err = git.PlainOpen(localRepo)
	if err != nil {
		pterm.Fatal.Println("Cannot open local repo", localRepository)
	}

	// analyze remote
	origin, err = localRepository.Remote(defaultOrigin)
	if err != nil {
		pterm.Warning.Println(fmt.Sprintf("Cannot read %s remote url", defaultOrigin), err)
	}

	url := origin.Config().URLs[0]
	// Check if local repo is github

	for tp, re := range origins {
		if m := re.FindStringSubmatch(url); m != nil {
			subexp := make(map[string]string)
			for i, name := range re.SubexpNames() {
				if name == "" {
					continue
				}
				subexp[name] = m[i]
			}
			if account == "" {
				account = subexp["org"]
			}
			if repoSlug == "" {
				repoSlug = subexp["repo"]
			}
			originType = tp
			break
		}
	}
}

type OriginType int

const (
	GitHubOriginType OriginType = iota
	BitbucketOriginType
	UnknownOriginType
)

var originType = UnknownOriginType

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

var localRepo string

var sshKeyComment string

func GetClient() (*bitbucket.APIClient, context.Context) {
	cfg := bitbucket.NewConfiguration()
	cfg.HTTPClient = &http.Client{}
	auth := bitbucket.BasicAuth{UserName: *_username, Password: *_password}
	ctx := context.WithValue(context.Background(), bitbucket.ContextBasicAuth, auth)
	return bitbucket.NewAPIClient(cfg), ctx
}

func GetSv() sv.Sv {
	if len(githubToken) > 0 {
		if originType != GitHubOriginType {
			pterm.Warning.Println("Remote '%s' mismatches with origin url : %s", defaultOrigin, origin.Config().URLs[0])
		}
		return sv.NewGitHubSv(githubToken, localRepo, sshKeyComment, account, repoSlug)
	} else {
		if originType != BitbucketOriginType {
			pterm.Warning.Println("Remote '%s' mismatches with origin url : %s", defaultOrigin, origin.Config().URLs[0])
		}
		return sv.NewBitBucketSv(*_username, *_password, repoSlug, account, localRepo)
	}

}

func init() {
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	wd, err := os.Getwd()
	if err != nil {
		pterm.Warning.Println("Cannot get wd", wd)
	}
	// rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.gobb-cli.yaml)")
	_username = rootCmd.PersistentFlags().StringP("username", "u", "", "Username")
	_password = rootCmd.PersistentFlags().StringP("password", "p", "", "Password")
	rootCmd.PersistentFlags().StringVarP(&account, "account", "a", "", "Account (default value will be deduced from the local repo)")
	rootCmd.PersistentFlags().StringVarP(&repoSlug, "repository", "r", "", "Repository (Account (default value will be deduced from the local repo)")
	rootCmd.PersistentFlags().StringVarP(&githubToken, "token", "t", os.Getenv("GITHUB_TOKEN"), "Github token")
	rootCmd.PersistentFlags().StringVarP(&localRepo, "workspace", "w", wd, "Local copy")
	rootCmd.PersistentFlags().StringVar(&defaultOrigin, "remote", "origin", "Default origin to use")
	rootCmd.PersistentFlags().StringVarP(&sshKeyComment, "ssh-key-comment", "K", ".*", "REGEXP that should match with the SSH key to be used")
	// Cobra also supports local flags, which will only run
	// when this action is called directly.

}
