package cmd

import (
	"log"
	"os"

	"github.com/spf13/cobra"
	"legendu.net/icon/cmd/ai"
	"legendu.net/icon/cmd/bigdata"
	"legendu.net/icon/cmd/dev"
	"legendu.net/icon/cmd/filesystem"
	"legendu.net/icon/cmd/ide"
	"legendu.net/icon/cmd/jupyter"
	"legendu.net/icon/cmd/misc"
	"legendu.net/icon/cmd/network"
	"legendu.net/icon/cmd/shell"
	"legendu.net/icon/cmd/virtualization"
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
	Args:                  cobra.ExactValidArgs(1),
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

var rootCmd = &cobra.Command{
	Use:              "icon",
	Short:            "Install and configure tools.",
	TraverseChildren: true,
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
	rootCmd.AddCommand(network.DownloadGitHubReleaseCmd)
	rootCmd.AddCommand(network.SshClientCmd)
	rootCmd.AddCommand(network.SshServerCmd)
	rootCmd.AddCommand(jupyter.IpythonCmd)
	rootCmd.AddCommand(jupyter.JLabVimCmd)
	rootCmd.AddCommand(jupyter.JupyterBookCmd)
	rootCmd.AddCommand(jupyter.GanymedeCmd)
	rootCmd.AddCommand(shell.AlacrittyCmd)
	rootCmd.AddCommand(shell.AtuinCmd)
	rootCmd.AddCommand(shell.BashItCmd)
	rootCmd.AddCommand(shell.FishCmd)
	rootCmd.AddCommand(shell.HyperCmd)
	rootCmd.AddCommand(shell.NushellCmd)
	rootCmd.AddCommand(shell.ZellijCmd)
	rootCmd.AddCommand(dev.GitCmd)
	rootCmd.AddCommand(dev.GolangCmd)
	rootCmd.AddCommand(dev.RustCmd)
	rootCmd.AddCommand(dev.BytehoundCmd)
	rootCmd.AddCommand(dev.PytypeCmd)
	rootCmd.AddCommand(ide.NeovimCmd)
	rootCmd.AddCommand(ide.HelixCmd)
	rootCmd.AddCommand(ide.FirenvimCmd)
	rootCmd.AddCommand(ide.VscodeCmd)
	rootCmd.AddCommand(ai.PyTorchCmd)
	rootCmd.AddCommand(bigdata.SparkCmd)
	rootCmd.AddCommand(virtualization.DockerCmd)
	rootCmd.AddCommand(misc.KeepassXCCmd)
	rootCmd.AddCommand(filesystem.RipCmd)
	rootCmd.AddCommand(completionCmd)
}
