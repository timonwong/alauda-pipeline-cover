package cmd

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/timonwong/alauda-pipeline-cover/constants"
	"github.com/timonwong/alauda-pipeline-cover/covertool"
)

// writeCmd represents the write command
var writeCmd = &cobra.Command{
	Use:    "write coverage",
	Short:  "Write coverage data",
	PreRun: prerunBindViperFlags,
	Args:   cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cover, err := covertool.New(
			viper.GetString(constants.APIBase), viper.GetString(constants.APIToken), viper.GetString(constants.ProjectID))
		if err != nil {
			return fmt.Errorf("failed to initialize covertool: %w", err)
		}

		coverage, err := strconv.ParseFloat(args[0], 64)
		if err != nil {
			return errors.New("coverage must in float")
		}

		return cover.Write(cmd.Context(),
			viper.GetString(constants.PipelineName), viper.GetString(constants.GitRef), viper.GetString(constants.GitSHA), coverage)
	},
}

func init() {
	rootCmd.AddCommand(writeCmd)

	writeCmd.Flags().String(constants.GitRef, "", "The git ref name for target branch")
	writeCmd.Flags().String(constants.GitSHA, "", "Optional git SHA hash for target ref")
	writeCmd.MarkFlagRequired(constants.GitRef) // nolint: errcheck
}
