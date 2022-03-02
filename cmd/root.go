package cmd

import (
	"log"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/timonwong/alauda-pipeline-cover/constants"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use: "alauda-pipeline-cover",
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))

	bindViper := func(name string) {
		if err := viper.BindPFlag(name, rootCmd.PersistentFlags().Lookup(name)); err != nil {
			log.Fatalf("failed to bind flag: %v", err)
		}
	}

	addStringFlag := func(name, value, usage string) {
		rootCmd.PersistentFlags().String(name, value, usage)
		bindViper(name)
	}

	addStringFlag(constants.RootAPIBase, "https://gitlab.com/api/v4", "Base API URL for gitlab")
	addStringFlag(constants.RootAPIToken, "", "GitLab API Token")
	addStringFlag(constants.RootProjectID, "", "Gitlab Project ID")
	addStringFlag(constants.RootGitRef, "", "The git ref name for target branch")
}
