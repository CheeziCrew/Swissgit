package ops

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// Config holds user-configurable settings for swissgit.
type Config struct {
	// ChangeTypes overrides the default PR change categories.
	ChangeTypes []string `json:"changeTypes,omitempty"`

	// TargetBranch overrides the default PR target branch (default: "main").
	TargetBranch string `json:"targetBranch,omitempty"`

	// TemplatePath points to a custom PR body template file.
	// If empty, the embedded default template is used.
	TemplatePath string `json:"templatePath,omitempty"`
}

// DefaultConfig returns the built-in defaults.
func DefaultConfig() Config {
	return Config{
		ChangeTypes:  defaultChangeTypes,
		TargetBranch: "main",
	}
}

// LoadConfig reads ~/.swissgit/config.json.
// Returns defaults for any missing or invalid fields.
func LoadConfig() Config {
	cfg := DefaultConfig()

	home, err := os.UserHomeDir()
	if err != nil {
		return cfg
	}

	data, err := os.ReadFile(filepath.Join(home, ".swissgit", "config.json"))
	if err != nil {
		return cfg
	}

	var user Config
	if err := json.Unmarshal(data, &user); err != nil {
		return cfg
	}

	if len(user.ChangeTypes) > 0 {
		cfg.ChangeTypes = user.ChangeTypes
	}
	if user.TargetBranch != "" {
		cfg.TargetBranch = user.TargetBranch
	}
	if user.TemplatePath != "" {
		cfg.TemplatePath = user.TemplatePath
	}

	return cfg
}
