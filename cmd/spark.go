package cmd

import (
    "fmt"
    "github.com/spf13/cobra"
)

var sparkCmd = &cobra.Command{
    Use:   "spark",
    Aliases: []string{},
    Short:  "Install and configure Spark.",
    //Args:  cobra.ExactArgs(1),
    Run: func(cmd *cobra.Command, args []string) {
    },
}

func init() {
    rootCmd.AddCommand(sparkCmd)
}
