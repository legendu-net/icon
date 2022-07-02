package cmd

import (
	"github.com/spf13/cobra"
	"os"
)

var rootCmd = &cobra.Command{
	Use:              "icon",
	Short:            "Install and configure tools.",
	TraverseChildren: true,
	/*
		Run: func(cmd *cobra.Command, args []string) {
		},
	*/
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().Bool("sudo", false, "Run commands using sudo.")
}
