package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/timonwong/alauda-pipeline-cover/constants"
	"github.com/timonwong/alauda-pipeline-cover/coverreport"
	"github.com/timonwong/alauda-pipeline-cover/covertool"
)

// compareCmd represents the compare command
var compareCmd = &cobra.Command{
	Use: "compare",
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

		// Generate coverage report
		const packages = true
		report, err := coverreport.GenerateReport(viper.GetString(constants.CoverProfile), &coverreport.Configuration{
			SortBy: coverreport.SortByPackage,
			Order:  coverreport.OrderDesc,
		}, packages)
		if err != nil {
			return err
		}

		// Print coverage table
		coverreport.PrintTable(report, os.Stdout, packages)

		// Check threshold
		threshold := viper.GetFloat64(constants.DefaultThreshold)
		if coverage.Valid && coverage.Float64 > threshold {
			threshold = coverage.Float64
		}

		if report.Total.StmtCoverage < threshold {
			fmt.Printf("ERROR: Your coverage is below %.2f%%\n!", threshold)
			os.Exit(1)
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(compareCmd)

	compareCmd.Flags().String(constants.CoverProfile, "coverage.out", "Coverage output file (default coverage.out)")
	compareCmd.Flags().Float64(constants.DefaultThreshold, 0, "The default coverage threshold")
	compareCmd.MarkFlagRequired(constants.CoverProfile) // nolint: errcheck
}
