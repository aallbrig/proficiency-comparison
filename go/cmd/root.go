package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

const Version = "1.0.0"

var rootCmd = &cobra.Command{
	Use:   "edu-stats-cli",
	Short: "Educational Statistics CLI Tool",
	Long: `A command-line tool for downloading and managing US educational statistics.
	
Downloads data from authoritative sources including World Bank, US Census Bureau,
NCES, NAEP, and ECLS. Stores data in SQLite and generates assets for Hugo website.`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.AddCommand(statusCmd)
	rootCmd.AddCommand(allCmd)
	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(upgradeCmd)
}
