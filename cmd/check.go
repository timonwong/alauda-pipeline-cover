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
	RunE:   runCheck,
}

func runCheck(cmd *cobra.Command, args []string) error {
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

	cfg, err := readCoverCheckConfig()
	if err != nil {
		log.Fatalf("unable to read config: %v", err)
	}

	// Generate coverage report
	packages := cfg.GetString("mode") == "packages"
	report, err := coverreport.GenerateReport(viper.GetString(constants.CoverProfile), &coverreport.Configuration{
		Root:       cfg.GetString("root"),
		Exclusions: cfg.GetStringSlice("excludes"),
		SortBy:     cfg.GetString("sort_by"),
		Order:      cfg.GetString("order"),
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

	leeway := viper.GetFloat64(constants.Leeway)
	if report.Total.StmtCoverage < threshold-leeway {
		log.Fatalf("ERROR: Your coverage is below %.2f%% (leeway=%.2f%%)!", threshold, leeway)
	}

	return nil
}

func readCoverCheckConfig() (*viper.Viper, error) {
	v := viper.New()

	v.SetDefault("sort_by", coverreport.SortByPackage)
	v.SetDefault("order", coverreport.OrderDesc)
	v.SetDefault("mode", "packages")

	v.SetConfigName(".covercheck")
	v.SetConfigType("yaml")
	v.AddConfigPath(".")
	err := v.ReadInConfig()
	if err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, err
		}
	}

	return v, nil
}

func init() {
	rootCmd.AddCommand(checkCmd)

	checkCmd.Flags().String(constants.GitRef, "", "The git ref name for target branch")
	checkCmd.Flags().String(constants.CoverProfile, "coverage.out", "Coverage output file (default coverage.out)")
	checkCmd.Flags().Float64(constants.DefaultThreshold, 0, "The default coverage threshold")
	checkCmd.Flags().Float64(constants.Leeway, 0, "Allow coverage to drop by leeway")
	checkCmd.MarkFlagRequired(constants.CoverProfile) // nolint: errcheck
}
