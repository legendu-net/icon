package utils

import (
	"context"
	"log"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"

	"github.com/elliotchance/orderedmap/v2"
	"golang.org/x/sys/unix"
)

// RunCmd executes a command (using pwsh or bash based on the OS) in the terminal.
//
// @param cmd The command to execute as a string.
// @param env Optional environment variables to set for the command execution.
//
// @example RunCmd("ls -l", "MY_VAR=myvalue")
func RunCmd(cmd string, env ...string) {
	command := exec.CommandContext(context.Background(), "bash", "-c", cmd)
	command.Env = append(os.Environ(), env...)
	command.Stdin = os.Stdin
	command.Stdout = os.Stdout
	command.Stderr = os.Stderr
	err := command.Run()
	if err != nil {
		log.Fatal("ERROR - ", err, ": when running the command:\n", cmd)
	}
}

// Format replaces placeholders in a string with values from a map.
//
// This function takes a command string `cmd` and a map `hmap` as input. It
// iterates through the map, replacing each occurrence of a placeholder in
// `cmd` with the corresponding value from the map. Placeholders are denoted
// by the format "{key}" where "key" is a key in the `hmap`.
//
// @param cmd The command string with placeholders.
// @param hmap A map where keys are placeholder names and values are their replacements.
//
// @return The formatted command string with placeholders replaced by their values.
//
// @example Format("Hello, {name}!", map[string]string{"name": "World"}) // Returns "Hello, World!"
func Format(cmd string, hmap map[string]string) string {
	for key, val := range hmap {
		cmd = strings.ReplaceAll(cmd, "{"+key+"}", val)
	}
	return cmd
}

// Max returns the larger of two integers.
//
// @param x The first integer.
// @param y The second integer.
//
// @return The larger of the two integers.
func Max(x int, y int) int {
	if x >= y {
		return x
	}
	return y
}

// GetCurrentUser retrieves the current user's information.
//
// @return A pointer to a `user.User` struct representing the current user.
func GetCurrentUser() *user.User {
	currentUser, err := user.Current()
	if err != nil {
		log.Fatal("ERROR - ", err)
	}
	return currentUser
}

// Returns "sudo" or "" depending on whether sudo is accessible by the current user.
func sudo() string {
	_, err := exec.LookPath("sudo")
	if err != nil {
		return ""
	}
	RunCmd("sudo true")
	return "sudo"
}

// GetCommandPrefix determines the appropriate command prefix for running commands.
//
// @param forceSudo A boolean indicating whether to force the use of sudo.
// @param pathPerms A map where keys are file paths and values are permission
//
// @return The appropriate command prefix ("sudo" or an empty string).
func GetCommandPrefix(forceSudo bool, pathPerms map[string]uint32) string {
	switch runtime.GOOS {
	case "darwin", "linux":
		if GetCurrentUser().Uid != "0" {
			if forceSudo {
				return sudo()
			}
			for path, perm := range pathPerms {
				path = NormalizePath(path)
				for !ExistsPath(path) {
					path = filepath.Dir(path)
					perm |= unix.X_OK
				}
				if unix.Access(path, perm) != nil {
					return sudo()
				}
			}
		}
	}
	return ""
}

// ExistsCommand checks if a command exists in the system's PATH.
//
// @param cmd The name of the command to check for.
//
// @return true if the command exists in the PATH, false otherwise.
func ExistsCommand(cmd string) bool {
	cmd = NormalizePath(cmd)
	_, err := exec.LookPath(cmd)
	return err == nil
}

// IfElseString returns one of two strings based on a boolean condition.
//
// @param b The boolean condition to evaluate.
// @param t The string to return if `b` is true.
// @param f The string to return if `b` is false.
//
// @return `t` if `b` is true, `f` otherwise.
func IfElseString(b bool, t string, f string) string {
	if b {
		return t
	}
	return f
}

// Using Homebrew to install packages.
//
// @param pkgs A slice of strings representing the packages to install.
func BrewInstallSafe(pkgs []string) {
	for _, pkg := range pkgs {
		command := Format("brew install --force {pkg} || brew link --overwrite --force {pkg}", map[string]string{
			"pkg": pkg,
		})
		RunCmd(command)
	}
}

// Update map1 using map2.
//
// @param map1 The target map to be updated.
// @param map2 The source map from which to take updates.
func UpdateMap(map1 orderedmap.OrderedMap[string, any], map2 orderedmap.OrderedMap[string, any]) {
	for _, key2 := range map2.Keys() {
		val2, _ := map2.Get(key2)
		val1, map1HasKey2 := map1.Get(key2)
		if !map1HasKey2 {
			map1.Set(key2, val2)
			continue
		}
		switch t2 := val2.(type) {
		case orderedmap.OrderedMap[string, any]:
			switch t1 := val1.(type) {
			case orderedmap.OrderedMap[string, any]:
				UpdateMap(t1, t2)
			default:
				map1.Set(key2, val2)
			}
		default:
			map1.Set(key2, val2)
		}
	}
}

// ParseInt converts a string to an int64.
//
// @param str The string to convert to an int64.
//
// @return The int64 representation of the string.
func ParseInt(str string) int64 {
	i, err := strconv.ParseInt(str, 10, 64)
	if err != nil {
		log.Fatalf("Error converting string to int64: %v\n", err)
	}
	return i
}
