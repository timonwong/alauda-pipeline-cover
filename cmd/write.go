package cmd

import (
	"errors"
	"log"
	"strconv"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/timonwong/alauda-pipeline-cover/constants"
	"github.com/timonwong/alauda-pipeline-cover/covertool"
)

// writeCmd represents the write command
var writeCmd = &cobra.Command{
	Use:   "write coverage",
	Short: "Write coverage data",
	Args: func(cmd *cobra.Command, args []string) error {
		err := cobra.ExactArgs(1)(cmd, args)
		if err != nil {
			return err
		}

		_, err = strconv.ParseFloat(args[0], 64)
		if err != nil {
			return errors.New("coverage must in float")
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		cover, err := covertool.New(
			viper.GetString(constants.RootAPIBase), viper.GetString(constants.RootAPIToken), viper.GetString(constants.RootProjectID))
		if err != nil {
			return err
		}

		coverage, _ := strconv.ParseFloat(args[0], 64)
		return cover.Write(cmd.Context(),
			viper.GetString(constants.RootGitRef), viper.GetString(constants.WriteGitSHA), coverage)
	},
}

func init() {
	rootCmd.AddCommand(writeCmd)

	writeCmd.Flags().String(constants.WriteGitSHA, "", "The git SHA hash for target ref")
	if err := viper.BindPFlag(constants.WriteGitSHA, writeCmd.Flags().Lookup(constants.WriteGitSHA)); err != nil {
		log.Fatalf("failed to bind flag: %v", err)
	}
}