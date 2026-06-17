package utils

import (
	"log"

	"gopkg.in/yaml.v3"
)

// UserConfig holds the user identity (name and email) shared across tools such
// as Git, jj and gopass. It is single-sourced from ~/.config/icon-data/user.yaml.
type UserConfig struct {
	UserName  string `yaml:"userName"`
	UserEmail string `yaml:"userEmail"`
}

// ReadUserConfig reads and validates the shared user identity from
// ~/.config/icon-data/user.yaml. It terminates the program if the file is
// missing, cannot be parsed, or does not define both userName and userEmail.
func ReadUserConfig() UserConfig {
	file := "~/.config/icon-data/user.yaml"
	if !ExistsFile(file) {
		log.Fatalf("The user configuration file %s does not exist.", file)
	}
	var cfg UserConfig
	if err := yaml.Unmarshal(ReadFile(file), &cfg); err != nil {
		log.Fatalf("Error parsing %s: %v", file, err)
	}
	if cfg.UserName == "" {
		log.Fatalf("userName is not configured in %s.", file)
	}
	if cfg.UserEmail == "" {
		log.Fatalf("userEmail is not configured in %s.", file)
	}
	return cfg
}
