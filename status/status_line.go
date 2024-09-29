package status

import (
	"fmt"

	"github.com/fatih/color"
)

// buildStatusLine constructs the status line with colors and the repository information.
func buildStatusLine(branch string, ahead, behind, modified, added, deleted, untracked int) string {
	// Colors for different statuses
	green := color.New(color.FgGreen).SprintFunc()
	yellow := color.New(color.FgYellow).SprintFunc()
	red := color.New(color.FgRed).SprintFunc()
	blue := color.New(color.FgBlue).SprintFunc()
	magenta := color.New(color.FgMagenta).SprintFunc()

	// Set branch color: green if "main", magenta otherwise
	var branchColorFunc func(a ...interface{}) string
	if branch == "main" {
		branchColorFunc = green
	} else {
		branchColorFunc = magenta
	}

	// Build the status line
	statusLine := fmt.Sprintf("[%s] ", branchColorFunc(branch))
	if ahead > 0 || behind > 0 {
		statusLine += fmt.Sprintf("[%s↑/%s↓] ", green(ahead), red(behind))
	}
	if modified > 0 {
		statusLine += fmt.Sprintf("[%s %s] ", yellow("Modified:"), yellow(modified))
	}
	if added > 0 {
		statusLine += fmt.Sprintf("[%s %s] ", green("Added:"), green(added))
	}
	if deleted > 0 {
		statusLine += fmt.Sprintf("[%s %s] ", red("Deleted:"), red(deleted))
	}
	if untracked > 0 {
		statusLine += fmt.Sprintf("[%s %s] ", blue("Untracked:"), blue(untracked))
	}

	return statusLine
}
