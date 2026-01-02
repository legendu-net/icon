package utils

import (
	"github.com/spf13/cobra"
)

// BuildPipUninstall constructs a command to uninstall a Python package using pip.
//
// @param cmd A pointer to a Cobra command object.
//
// @return A string representing the pip uninstall command.
func BuildPipUninstall(cmd *cobra.Command) string {
	python := GetStringFlag(cmd, "python")
	return Format("{python} -m pip uninstall", map[string]string{
		"python": python,
	})
}

// BuildPipInstall constructs a command to install a Python package using pip.
//
// @param cmd A pointer to a Cobra command object.
//
// @return A string representing the pip install command.
func BuildPipInstall(cmd *cobra.Command) string {
	python := GetStringFlag(cmd, "python")
	if LookPath(python) == "" {
		return ""
	}
	user := ""
	if GetBoolFlag(cmd, "user") {
		user = "--user"
	}
	extraPipOptions := GetStringSliceFlag(cmd, "extra-pip-options")
	options := ""
	for _, option := range extraPipOptions {
		options += "--" + option
	}
	return Format("PIP_BREAK_SYSTEM_PACKAGES=1 {python} -m pip install {user} {options}", map[string]string{
		"python":  python,
		"user":    user,
		"options": options,
	})
}

// AddPythonFlags adds common Python-related flags to a Cobra command.
//
// These flags are commonly used when working with Python packages and installations.
//
// @param cmd A pointer to the Cobra command to which the flags will be added.
func AddPythonFlags(cmd *cobra.Command) {
	cmd.Flags().String("python", "python3", "Path to the python3 command.")
	cmd.Flags().Bool("user", false, "Install Python packages to user's local directory.")
	cmd.Flags().StringSlice("extra-pip-options", []string{}, "Extra options (separated by comma) to pass to pip.")
}
