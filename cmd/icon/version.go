package icon

import (
	"fmt"

	"github.com/spf13/cobra"
)

// Show the version of icon.
func version(_ *cobra.Command, _ []string) {
	fmt.Println("0.32.1")
}

var VersionCmd = &cobra.Command{
	Use:     "version",
	Aliases: []string{"v"},
	Short:   "Show the version of icon.",
	Run:     version,
}

func init() {
}
