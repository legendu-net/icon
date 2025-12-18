package utils

import (
	"log"
	"path/filepath"
	"strings"
)

// ConfigBash configures the Bash shell environment.
//
// This function performs the following configurations to the Bash shell:
//   - configure the shell's PATH environment variable smartly
//   - set the environment variables VISUAL and EDITOR to nvim with a fallback to vim.
func ConfigBash() {
	bashConfigFile := GetBashConfigFile()
	ConfigShellPath(bashConfigFile)
	AppendToTextFile(
		bashConfigFile,
		`
if which nvim > /dev/null; then
	export VISUAL=nvim
	export EDITOR=nvim
else
	export VISUAL=vim
	export EDITOR=vim
fi
`,
		true,
	)
	if IsLinux() {
		sourceIn := `
# source in ~/.bashrc
if [[ -f $HOME/.bashrc ]]; then
	. $HOME/.bashrc
fi
`
		AppendToTextFile(
			filepath.Join(UserHomeDir(), ".bash_profile"),
			sourceIn,
			true,
		)
	}
}

// Configure shell to add a path into the environment variable PATH.
// @param paths: Absolute paths to add into PATH.
// @param config_file: The path of a shell's configuration file.
func ConfigShellPath(configFile string) {
	if GetLinuxDistID() == "idx" {
		return
	}
	text := ReadFileAsString(configFile)
	if !strings.Contains(text, ". /scripts/path.sh") && !strings.Contains(text, "\n_PATHS=(\n") {
		text = `
# set $PATH
_PATHS=(
	$(ls -d $HOME/*/bin 2> /dev/null)
	$(ls -d $HOME/.*/bin 2> /dev/null)
	$(ls -d $HOME/Library/Python/3.*/bin 2> /dev/null)
	$(ls -d /usr/local/*/bin 2> /dev/null)
	$(ls -d /opt/*/bin 2> /dev/null)
)
for ((_i=${#_PATHS[@]}-1; _i>=0; _i--)); do
	_PATH=${_PATHS[$_i]}
	if [[ -d $_PATH && ! "$PATH" =~ (^$_PATH:)|(:$_PATH:)|(:$_PATH$) ]]; then
		export PATH=$_PATH:$PATH
	fi
done
`
		AppendToTextFile(configFile, text, true)
	}
	log.Printf("%s is configured to insert common bin paths into $PATH.", configFile)
}

// GetBashConfigFile returns the path to the Bash configuration file based on the operating system.
//
// @return The path to the Bash configuration file as a string.
func GetBashConfigFile() string {
	home := UserHomeDir()
	file := ".bash_profile"
	if IsLinux() {
		file = ".bashrc"
	}
	return filepath.Join(home, file)
}
