package branches

import (
	"fmt"

	"github.com/fatih/color"
)

// buildAndPrintStatus constructs and prints the status line.
func buildStatus(currentBranch string, localBranches, remoteBranches []string) string {
	// Colors
	blue := color.New(color.FgBlue).SprintFunc()
	red := color.New(color.FgRed).SprintFunc()

	if currentBranch == "main" {
		currentBranch = blue(currentBranch)
	} else {
		currentBranch = red(currentBranch)
	}

	// Build the status line
	statusLine := fmt.Sprintf("[Current: %s] [Local: %s] [Remote: %s]",
		currentBranch,
		formatBranches(localBranches),
		formatBranches(remoteBranches),
	)

	return statusLine
}

// formatBranches formats the branches for display.
func formatBranches(branches []string) string {
	if len(branches) == 0 {
		return "None"
	}
	return fmt.Sprintf("%s", branches)
}
