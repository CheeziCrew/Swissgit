package pull_request

import (
	_ "embed"
	"fmt"
	"strings"
)

//go:embed PULL_REQUEST_TEMPLATE.md
var pullRequestTemplate string

// BuildPullRequestBody reads the pull request template, takes parameters, and updates the content.
func BuildPullRequestBody(changes []string, breakingChange bool) (string, error) {

	// Copy template content into a new string
	templateContent := pullRequestTemplate

	// Replace change types with the provided changes
	changeTypes := []string{
		"Bug fix",
		"New feature",
		"Removed feature",
		"Code style update (formatting etc.)",
		"Refactoring (no functional changes, no api changes)",
		"Build related changes",
		"Documentation content changes",
	}

	for _, change := range changeTypes {
		checkbox := fmt.Sprintf("- [ ] %s", change)
		if contains(changes, change) {
			checkbox = fmt.Sprintf("- [x] %s", change)
		}
		templateContent = strings.ReplaceAll(templateContent, fmt.Sprintf("- [ ] %s", change), checkbox)
	}

	// Handle breaking change
	if breakingChange {
		templateContent = strings.ReplaceAll(templateContent, "- [ ] Yes (I have stepped the version number accordingly)", "- [x] Yes (I have stepped the version number accordingly)")
	} else {
		templateContent = strings.ReplaceAll(templateContent, "- [ ] No", "- [x] No")
	}

	return templateContent, nil
}

// Helper function to check if a slice contains a specific string
func contains(slice []string, item string) bool {
	for _, v := range slice {
		if v == item {
			return true
		}
	}
	return false
}
