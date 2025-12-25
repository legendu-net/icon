package virtualization

import (
	"github.com/spf13/cobra"
	"legendu.net/icon/utils"
)

// Configure a KVM virtual machine.
func kvm(cmd *cobra.Command, _ []string) {
	if utils.GetBoolFlag(cmd, "install") {
	}
	if utils.GetBoolFlag(cmd, "config") {
		if utils.IsLinux() {
			if utils.IsDebianUbuntuSeries() {
				cmd := utils.Format(`{prefix} dmesg | grep -q 'DMI: QEMU' \
							&& {prefix} apt-get update \
							&& {prefix} apt-get install {yesStr} spice-vdagent`, map[string]string{
					"prefix": utils.GetCommandPrefix(true, map[string]uint32{}),
				})
				utils.RunCmd(cmd)
			} else if utils.IsFedoraSeries() {
				cmd := utils.Format(`{prefix} dmesg | grep -q 'DMI: QEMU' \
							&& {prefix} dnf install {yesStr} spice-vdagent`, map[string]string{
					"prefix": utils.GetCommandPrefix(true, map[string]uint32{}),
				})
				utils.RunCmd(cmd)
			}
		}
	}
}

var kvmCmd = &cobra.Command{
	Use:     "kvm",
	Aliases: []string{},
	Short:   "Install and configure KVM related tools.",
	Run:     kvm,
}

func ConfigKVMCmd(rootCmd *cobra.Command) {
	kvmCmd.Flags().BoolP("install", "i", false, "Install KVM related tools.")
	kvmCmd.Flags().BoolP("config", "c", false, "Configure KVM related tools.")
	kvmCmd.Flags().BoolP("yes", "y", false, "Automatically yes to prompt questions.")
	kvmCmd.Flags().Bool("no-backup", false, "Do not backup existing configuration files.")
	kvmCmd.Flags().Bool("copy", false, "Make copies (instead of symbolic links) of configuration files.")
	kvmCmd.Flags().BoolP("uninstall", "u", false, "Uninstall KVM related tools.")
	rootCmd.AddCommand(kvmCmd)
}
