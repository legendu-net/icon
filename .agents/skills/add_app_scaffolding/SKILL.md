---
name: add-app-scaffolding
description: Creates scaffolding for a new app installation and configuration command in the icon CLI project.
---

# Instructions

When the user asks to add support for installing and configuring a new app in `icon` using this skill, you should:

1. **Determine the app name:**

   - If the user didn't provide the name of the app (e.g., they just invoked the skill via `/add-app-scaffolding`), ask them for the name of the app first before proceeding.

2. **Determine the appropriate category:**

   - Examine the existing packages under `cmd/` (e.g., `ai`, `bigdata`, `dev`, `filesystem`, `ide`, `jupyter`, `misc`, `network`, `shell`, `virtualization`).
   - Choose the most appropriate category for the new app. If unsure, ask the user or default to `misc`.

3. **Create the command file:**

   - Create a new file `cmd/<category>/<app>.go` (e.g., `cmd/dev/nodejs.go`).
   - Generate the following scaffolding, replacing placeholders (`<category>`, `<app>`, `<App>`, `<alias>`) with the actual values. Use Title Case for `<App>` (e.g., `Nodejs`) and lowercase for `<app>` (e.g., `nodejs`).

   ```go
   package <category>

   import (
   	"github.com/spf13/cobra"
   	"legendu.net/icon/utils"
   )

   // Install and configure <App>.
   func <app>(cmd *cobra.Command, args []string) {
   	if utils.GetBoolFlag(cmd, "install") {
   		// TODO: implement installation
   	}
   	if utils.GetBoolFlag(cmd, "config") {
   		// TODO: implement configuration
   	}
   	if utils.GetBoolFlag(cmd, "uninstall") {
   		// TODO: implement uninstallation
   	}
   }

   var <app>Cmd = &cobra.Command{
   	Use:     "<app>",
   	Aliases: []string{"<alias>"},
   	Short:   "Install and configure <App>.",
   	//Args:  cobra.ExactArgs(1),
   	Run:     <app>,
   }

   func Config<App>Cmd(rootCmd *cobra.Command) {
   	<app>Cmd.Flags().BoolP("install", "i", false, "Install <App>.")
   	<app>Cmd.Flags().BoolP("uninstall", "u", false, "Uninstall <App>.")
   	<app>Cmd.Flags().BoolP("config", "c", false, "Configure <App>.")
   	<app>Cmd.Flags().Bool("no-backup", false, "Do not backup existing configuration files.")
   	<app>Cmd.Flags().Bool("copy", false, "Make copies (instead of symbolic links) of configuration files.")
   	rootCmd.AddCommand(<app>Cmd)
   }
   ```

   *Note: Remove the `Aliases` line if no alias is needed. If no alias is provided, just omit that line from the struct.*

4. **Register the command in `root.go`:**

   - Modify `cmd/root.go`.
   - Inside the `Execute()` function, find the section where commands are registered (e.g., `dev.ConfigGolangCmd(rootCmd)`).
   - Add the call to `<category>.Config<App>Cmd(rootCmd)` in the appropriate alphabetically sorted location for its package.

5. **Inform the user:**

   - Once the scaffolding is generated and registered, instruct the user that the code has been placed.
   - Remind them that the `install`, `config`, and `uninstall` logic inside `cmd/<category>/<app>.go` has been left empty for them to implement.
   - DO NOT attempt to write the implementation yourself.

6. **Format and verify:**

   - Run `golangci-lint fmt -d` to verify formatting, or ask the user to verify if they have linting setup.
