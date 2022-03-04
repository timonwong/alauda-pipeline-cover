package cmd

import (
	"fmt"
	"log"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gopkg.in/guregu/null.v4"

	"github.com/timonwong/alauda-pipeline-cover/constants"
	"github.com/timonwong/alauda-pipeline-cover/coverreport"
	"github.com/timonwong/alauda-pipeline-cover/covertool"
)

// checkCmd represents the check command
var checkCmd = &cobra.Command{
	Use:    "check",
	Short:  "check coverage data",
	PreRun: prerunBindViperFlags,
	RunE: func(cmd *cobra.Command, args []string) error {
		apiToken := viper.GetString(constants.APIToken)
		gitRef := viper.GetString(constants.GitRef)

		var coverage null.Float
		if apiToken == "" || gitRef == "" {
			log.Printf("WARNING: flag %s or %s is not set, skip reading coverage from api", constants.APIToken, constants.GitRef)
		} else {
			tool, err := covertool.New(viper.GetString(constants.APIBase), apiToken, viper.GetString(constants.ProjectID))
			if err != nil {
				return fmt.Errorf("failed to initialize covertool: %w", err)
			}

			coverage, err = tool.Read(cmd.Context(), viper.GetString(constants.PipelineName), gitRef)
			if err != nil {
				return fmt.Errorf("unable to read coverage from project: %w", err)
			}

			log.Printf("Successfully load coverage coverage %.2f from project", coverage.ValueOrZero())
		}

		// Generate coverage report
		const packages = true
		report, err := coverreport.GenerateReport(viper.GetString(constants.CoverProfile), &coverreport.Configuration{
			SortBy: coverreport.SortByPackage,
			Order:  coverreport.OrderDesc,
		}, packages)
		if err != nil {
			return fmt.Errorf("unable to read coverage: %w", err)
		}

		// Print coverage table
		coverreport.PrintTable(report, os.Stdout, packages)
		// Force flush
		os.Stdout.WriteString("\n")
		os.Stdout.Sync()

		// Check threshold
		threshold := viper.GetFloat64(constants.DefaultThreshold)
		log.Printf("Choose larger coverage between %.2f (default) and %.2f", threshold, coverage.ValueOrZero())
		if coverage.Valid && coverage.Float64 > threshold {
			threshold = coverage.Float64
		}

		if report.Total.StmtCoverage < threshold {
			log.Fatalf("ERROR: Your coverage is below %.2f%%!", threshold)
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(checkCmd)

	checkCmd.Flags().String(constants.GitRef, "", "The git ref name for target branch")
	checkCmd.Flags().String(constants.CoverProfile, "coverage.out", "Coverage output file (default coverage.out)")
	checkCmd.Flags().Float64(constants.DefaultThreshold, 0, "The default coverage threshold")
	checkCmd.MarkFlagRequired(constants.CoverProfile) // nolint: errcheck
}
