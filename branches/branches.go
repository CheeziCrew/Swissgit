package branches

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/CheeziCrew/swissgo/utils"
)

// ScanAndPrintBranches scans the specified directory and prints branch information.
func ScanAndPrintBranches(rootPath string, allFlag bool, verbose bool) {
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
				if utils.IsGitRepository(subDirPath) {
					PrintBranches(subDirPath, verbose)
				}
			}
		}
	} else {
		// Process only the current directory
		if utils.IsGitRepository(rootPath) {
			PrintBranches(rootPath, verbose)
		} else {
			fmt.Printf("Directory %s is not a Git repository.\n", rootPath)
		}
	}
}
