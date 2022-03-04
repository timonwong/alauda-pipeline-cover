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

	addGlobalStringFlag(constants.APIBase, "https://gitlab.com/api/v4", "Base API URL for gitlab")
	addGlobalStringFlag(constants.APIToken, "", "GitLab API Token")
	addGlobalStringFlag(constants.ProjectID, "", "Gitlab Project ID")
	addGlobalStringFlag(constants.PipelineName, "alauda-pipeline-cover", "Pipeline name (default: alauda-pipeline-cover)")
	rootCmd.MarkPersistentFlagRequired(constants.ProjectID)    // nolint: errcheck
	rootCmd.MarkPersistentFlagRequired(constants.PipelineName) // nolint: errcheck
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
	cmd.Flags().VisitAll(func(flag *pflag.Flag) {
		requiredAnnotation, found := flag.Annotations[cobra.BashCompOneRequiredFlag]
		if !found || requiredAnnotation[0] != "true" {
			return
		}
		if viper.IsSet(flag.Name) && viper.GetString(flag.Name) != "" {
			if err := cmd.Flags().Set(flag.Name, viper.GetString(flag.Name)); err != nil {
				log.Fatalf("failed to set flag: %v", err)
			}
		}
	})
}

func addGlobalStringFlag(name, value, usage string) {
	rootCmd.PersistentFlags().String(name, value, usage)
	if err := viper.BindPFlag(name, rootCmd.PersistentFlags().Lookup(name)); err != nil {
		log.Fatalf("failed to bind flag: %v", err)
	}
	if value != "" {
		viper.SetDefault(name, value)
	}
}

func prerunBindViperFlags(cmd *cobra.Command, args []string) {
	if err := viper.BindPFlags(cmd.NonInheritedFlags()); err != nil {
		log.Fatalf("failed to bind flags: %v", err)
	}
}
