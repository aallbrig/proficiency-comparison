package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show version information",
	Run:   runVersion,
}

func runVersion(cmd *cobra.Command, args []string) {
	fmt.Printf("edu-stats version %s\n", Version)
	fmt.Println("Educational Statistics CLI Tool")
	fmt.Println("https://github.com/aallbrig/proficiency-comparison")
}
