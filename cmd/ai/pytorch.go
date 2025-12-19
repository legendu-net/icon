package ai

import (
	"strings"

	"github.com/spf13/cobra"
	"legendu.net/icon/utils"
)

// Install and configure PyTorch.
func pytorch(cmd *cobra.Command, _ []string) {
	if utils.GetBoolFlag(cmd, "install") {
		cudaVersion := utils.GetStringFlag(cmd, "cuda-version")
		version := "cpu"
		if cudaVersion != "" {
			version = "cu" + strings.ReplaceAll(cudaVersion, ".", "")
		}
		command := utils.Format(`{pip_install} torch torchvision torchaudio \
				--extra-index-url https://download.pytorch.org/whl/{version}`, map[string]string{
			"pip_install": utils.BuildPipInstall(cmd),
			"version":     version,
		})
		utils.RunCmd(command)
	}
	if utils.GetBoolFlag(cmd, "config") {
	}
	if utils.GetBoolFlag(cmd, "uninstall") {
	}
}

var pyTorchCmd = &cobra.Command{
	Use:     "pytorch",
	Aliases: []string{"torch"},
	Short:   "Install and configure PyTorch.",
	//Args:  cobra.ExactArgs(1),
	Run: pytorch,
}

func ConfigPyTorchCmd(rootCmd *cobra.Command) {
	pyTorchCmd.Flags().BoolP("install", "i", false, "Install IPython.")
	pyTorchCmd.Flags().Bool("uninstall", false, "Uninstall IPython.")
	pyTorchCmd.Flags().BoolP("config", "c", false, "Configure IPython.")
	pyTorchCmd.Flags().Bool("no-backup", false, "Do not backup existing configuration files.")
	pyTorchCmd.Flags().Bool("copy", false, "Make copies (instead of symbolic links) of configuration files.")
	pyTorchCmd.Flags().String("cuda-version", "", "The version of CUDA. If not specified, the CPU version is used.")
	utils.AddPythonFlags(pyTorchCmd)
	rootCmd.AddCommand(pyTorchCmd)
}
