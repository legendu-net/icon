package icon

import (
	"log"
	"os"

	"github.com/spf13/cobra"
)

var completionCmd = &cobra.Command{
	Use:   "completion [bash|zsh|fish]",
	Short: "Generate completion script",
	Long: `To load completions:

Bash:

$ icon completion bash > /etc/bash_completion.d/icon

Zsh:

# If your autocomplete is enabled in zshrc...
$ echo "autoload -U compinit; compinit" >> ~/.zshrc
# Then to load completions for each session, execute once:
$ icon completion zsh > "${fpath[1]}/_icon"

Fish:

$ icon completion fish > ~/.config/fish/completions/icon.fish
`,
	DisableFlagsInUseLine: true,
	Args:                  cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
	ValidArgs:             []string{"bash", "zsh", "fish"},
	Run: func(cmd *cobra.Command, args []string) {
		switch args[0] {
		case "bash":
			err := cmd.Root().GenBashCompletion(os.Stdout)
			if err != nil {
				log.Fatalf("Failed to generate completion script for bash: %v", err)
			}
		case "zsh":
			err := cmd.Root().GenZshCompletion(os.Stdout)
			if err != nil {
				log.Fatalf("Failed to generate completion script for bash: %v", err)
			}
		case "fish":
			err := cmd.Root().GenFishCompletion(os.Stdout, true)
			if err != nil {
				log.Fatalf("Failed to generate completion script for bash: %v", err)
			}
		case "powershell":
			err := cmd.Root().GenPowerShellCompletion(os.Stdout)
			if err != nil {
				log.Fatalf("Failed to generate completion script for bash: %v", err)
			}
		}
	},
}

func ConfigCompletionCmd(rootCmd *cobra.Command) {
	rootCmd.AddCommand(completionCmd)
}
