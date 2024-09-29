package cleanup

import (
	"github.com/fatih/color"
)

// constructStatusLine constructs a status line for the repository
func constructStatusLine(changes, currentBranch string, prunedBranches, branchesCount int) string {
	if changes == "" && branchesCount == 1 && currentBranch == "main" && prunedBranches == 0 {
		return ""
	}
	statusLine := ""
	if changes != "" {
		statusLine += color.BlueString("[Non Committed Changes: %s]", changes)
	}
	if branchesCount != 1 {
		statusLine += color.RedString("[Too many branches: %d]", branchesCount)
	}
	if currentBranch != "main" {
		statusLine += color.MagentaString("[Current Branch: %s]", currentBranch)
	}
	if prunedBranches != 0 {
		statusLine += color.YellowString("[Pruned Branches: %d]", prunedBranches)
	}

	return statusLine
}
