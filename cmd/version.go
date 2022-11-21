package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// Update icon.
func version(cmd *cobra.Command, args []string) {
	fmt.Println("0.13.0")
}

var versionCmd = &cobra.Command{
	Use:     "version",
	Aliases: []string{"v"},
	Short:   "Show the version of icon.",
	Run:     version,
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
