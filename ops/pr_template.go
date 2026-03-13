package ops

import (
	_ "embed" // embed is used for go:embed directives
	"fmt"
	"os"
	"strings"
)

//go:embed PULL_REQUEST_TEMPLATE.md
var pullRequestTemplate string

// defaultChangeTypes are the built-in change categories for a PR.
var defaultChangeTypes = []string{
	"Bug fix",
	"New feature",
	"Removed feature",
	"Code style update (formatting etc.)",
	"Refactoring (no functional changes, no api changes)",
	"Build related changes",
	"Documentation content changes",
}

// ChangeTypes are the available change categories for a PR.
// Deprecated: Use LoadConfig().ChangeTypes instead for configurable types.
var ChangeTypes = defaultChangeTypes

// BuildPullRequestBody fills in the PR template with the selected changes.
func BuildPullRequestBody(changes []string, breakingChange bool) (string, error) {
	cfg := LoadConfig()
	content := pullRequestTemplate
	if cfg.TemplatePath != "" {
		if data, err := os.ReadFile(cfg.TemplatePath); err == nil {
			content = string(data)
		}
	}

	changeSet := make(map[string]bool, len(changes))
	for _, c := range changes {
		changeSet[c] = true
	}

	for _, change := range ChangeTypes {
		old := fmt.Sprintf("- [ ] %s", change)
		replacement := old
		if changeSet[change] {
			replacement = fmt.Sprintf("- [x] %s", change)
		}
		content = strings.ReplaceAll(content, old, replacement)
	}

	if breakingChange {
		content = strings.ReplaceAll(content,
			"- [ ] Yes (I have stepped the version number accordingly)",
			"- [x] Yes (I have stepped the version number accordingly)")
	} else {
		content = strings.ReplaceAll(content,
			"- [ ] No",
			"- [x] No")
	}

	return content, nil
}
