package utils

import (
	"os"
	"path/filepath"
)

// isGitRepository checks if a directory is a Git repository.
func IsGitRepository(path string) bool {
	gitPath := filepath.Join(path, ".git")
	_, err := os.Stat(gitPath)
	return err == nil
}
