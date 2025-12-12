package utils

import (
	"log"

	"github.com/spf13/cobra"
)

// GetBoolFlag retrieves the boolean value of a flag from a Cobra command.
//
// @param cmd  A pointer to a Cobra command object.
// @param flag The name of the flag to retrieve.
//
// @return The boolean value of the flag.
func GetBoolFlag(cmd *cobra.Command, flag string) bool {
	b, err := cmd.Flags().GetBool(flag)
	if err != nil {
		log.Fatal("ERROR - ", err)
	}
	return b
}

// GetIntFlag retrieves the integer value of a flag from a Cobra command.
//
// @param cmd  A pointer to a Cobra command object.
// @param flag The name of the flag to retrieve.
//
// @return The integer value of the flag.
func GetIntFlag(cmd *cobra.Command, flag string) int {
	i, err := cmd.Flags().GetInt(flag)
	if err != nil {
		log.Fatal("ERROR - ", err)
	}
	return i
}

// GetStringFlag retrieves the string value of a flag from a Cobra command.
//
// @param cmd  A pointer to a Cobra command object.
// @param flag The name of the flag to retrieve.
//
// @return The string value of the flag.
func GetStringFlag(cmd *cobra.Command, flag string) string {
	s, err := cmd.Flags().GetString(flag)
	if err != nil {
		log.Fatal("ERROR - ", err)
	}
	return s
}

// GetStringSliceFlag retrieves the string slice value of a flag from a Cobra command.
//
// @param cmd  A pointer to a Cobra command object.
// @param flag The name of the flag to retrieve.
//
// @return The string slice value of the flag.
func GetStringSliceFlag(cmd *cobra.Command, flag string) []string {
	ss, err := cmd.Flags().GetStringSlice(flag)
	if err != nil {
		log.Fatal("ERROR - ", err)
	}
	return ss
}

// BuildYesFlag constructs a string flag for commands that require confirmation.
//
// @param cmd A pointer to a Cobra command object.
//
// @return "-y" if the "yes" flag is set, otherwise an empty string.
func BuildYesFlag(cmd *cobra.Command) string {
	return IfElseString(GetBoolFlag(cmd, "yes"), "-y", "")
}
