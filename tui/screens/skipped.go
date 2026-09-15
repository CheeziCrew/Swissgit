package screens

import (
	"fmt"
	"path/filepath"
	"strings"

	"charm.land/lipgloss/v2"

	"github.com/CheeziCrew/swissgit/git"
)

var skippedWarnStyle = lipgloss.NewStyle().Foreground(colorBrYlow).Bold(true)

// renderSkippedWarning renders a block naming the repos a scan found but could
// not open. Without it those repos drop out of a scan silently, which is how a
// repo with a broken .git/config stays broken for weeks. Returns "" when
// nothing was skipped, so callers can prepend it unconditionally.
func renderSkippedWarning(skipped []git.SkippedRepo) string {
	if len(skipped) == 0 {
		return ""
	}

	var b strings.Builder
	b.WriteString(skippedWarnStyle.Render(fmt.Sprintf("⚠  %d repo(s) skipped — could not be opened", len(skipped))))
	b.WriteString("\n")
	for _, s := range skipped {
		b.WriteString(prDimStyle.Render("   " + filepath.Base(s.Path) + " — " + s.Reason))
		b.WriteString("\n")
	}
	b.WriteString("\n")
	return b.String()
}
