package cmd

import (
	"github.com/spf13/cobra"
)

var stepCmd = &cobra.Command{
	Use:   "step",
	Short: "Run individual pipeline steps",
	Long: `Run individual steps of the data pipeline independently.
	
This allows you to re-run specific steps without running the entire pipeline.
Useful for testing or regenerating specific outputs.`,
}

func init() {
	// Add all individual step commands under 'step'
	stepCmd.AddCommand(checkSchemaCmd)
	stepCmd.AddCommand(downloadWorldBankCmd)
	stepCmd.AddCommand(downloadCensusCmd)
	stepCmd.AddCommand(downloadNCESCmd)
	stepCmd.AddCommand(downloadNAEPCmd)
	stepCmd.AddCommand(downloadECLSCmd)
	stepCmd.AddCommand(generateAssetsCmd)
	
	// Add flags that steps might need
	stepCmd.PersistentFlags().StringVar(&years, "years", "1970-2025", "Year range to download (format: YYYY-YYYY)")
	stepCmd.PersistentFlags().BoolVar(&dryRun, "dry-run", false, "Simulate without downloading data")
}
