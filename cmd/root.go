package cmd

import (
	"log"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
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
	cobra.OnInitialize(func() {
		viper.AutomaticEnv()
		viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))

		postInitCommands(rootCmd.Commands())
	})

	rootCmd.PersistentFlags().String(constants.APIBase, "https://gitlab.com/api/v4", "Base API URL for gitlab")
	rootCmd.PersistentFlags().String(constants.APIToken, "", "GitLab API Token")
	rootCmd.PersistentFlags().String(constants.ProjectID, "", "Gitlab Project ID")
	rootCmd.PersistentFlags().String(constants.PipelineName, "default", "Pipeline name (default: default)")
	rootCmd.PersistentFlags().String(constants.GitRef, "", "The git ref name for target branch")
	viper.SetDefault(constants.PipelineName, "default")
	rootCmd.MarkPersistentFlagRequired(constants.ProjectID)    // nolint: errcheck
	rootCmd.MarkPersistentFlagRequired(constants.PipelineName) // nolint: errcheck
	rootCmd.MarkPersistentFlagRequired(constants.GitRef)       // nolint: errcheck
}

func postInitCommands(commands []*cobra.Command) {
	for _, cmd := range commands {
		presetRequiredFlags(cmd)
		if cmd.HasSubCommands() {
			postInitCommands(cmd.Commands())
		}
	}
}

func presetRequiredFlags(cmd *cobra.Command) {
	if err := viper.BindPFlags(cmd.Flags()); err != nil {
		log.Fatalf("failed to bind flag: %v", err)
	}

	cmd.Flags().VisitAll(func(f *pflag.Flag) {
		if viper.IsSet(f.Name) && viper.GetString(f.Name) != "" {
			if err := cmd.Flags().Set(f.Name, viper.GetString(f.Name)); err != nil {
				log.Fatalf("failed to set flag: %v", err)
			}
		}
	})
}
