package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/timonwong/alauda-pipeline-cover/constants"
	"github.com/timonwong/alauda-pipeline-cover/covertool"
)

// readCmd represents the read command
var readCmd = &cobra.Command{
	Use:   "read",
	Short: "Read coverage data",
	RunE: func(cmd *cobra.Command, args []string) error {
		tool, err := covertool.New(
			viper.GetString(constants.APIBase), viper.GetString(constants.APIToken), viper.GetString(constants.ProjectID))
		if err != nil {
			return err
		}

		coverage, err := tool.Read(cmd.Context(), viper.GetString(constants.PipelineName), viper.GetString(constants.GitRef))
		if err != nil {
			return err
		}

		fmt.Printf("%.2f\n", coverage)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(readCmd)
}
