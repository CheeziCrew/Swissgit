package status

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/fatih/color"
)

// ScanAndCheckStatus scans the specified directory and prints branch information.
func ScanAndCheckStatus(rootPath string, allFlag bool, verbose bool) {
	if allFlag {
		// Get all subdirectories in the specified directory
		entries, err := os.ReadDir(rootPath)
		if err != nil {
			fmt.Printf("Error reading directory: %v\n", err)
			return
		}

		for _, entry := range entries {
			if entry.IsDir() {
				subDirPath := filepath.Join(rootPath, entry.Name())
				if isGitRepository(subDirPath) {
					CheckStatus(subDirPath, verbose)
				}
			}
		}
	} else {
		// Process only the current directory
		if isGitRepository(rootPath) {
			clean := CheckStatus(rootPath, verbose)
			if clean {
				green := color.New(color.FgGreen).SprintFunc()
				statusMessage := fmt.Sprintf("%s: Updated status:", filepath.Base(rootPath))
				fmt.Printf("\r%s [%s] %s\n", statusMessage, green("✔"), "No changes.")
			}
		} else {
			fmt.Printf("Directory %s is not a Git repository.\n", rootPath)
		}
	}
}

// isGitRepository checks if a directory is a Git repository.
func isGitRepository(path string) bool {
	gitPath := filepath.Join(path, ".git")
	_, err := os.Stat(gitPath)
	return err == nil
}
