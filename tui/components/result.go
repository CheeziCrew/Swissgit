package components

import (
	"fmt"
	"strings"

	"charm.land/lipgloss/v2"
)

var (
	summaryTitle = lipgloss.NewStyle().Bold(true).Foreground(cFg)
	resultOk     = lipgloss.NewStyle().Foreground(cGreen)
	resultFail   = lipgloss.NewStyle().Foreground(cRed)
	resultDim    = lipgloss.NewStyle().Foreground(cGray)
	resultAccent = lipgloss.NewStyle().Foreground(cBrMag)

	successBox = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(cGreen).
			Padding(0, 1)

	failBox = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(cRed).
		Padding(0, 1)

	summaryBox = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(cBrMag).
			Padding(0, 2).
			Bold(true)
)

// ResultModel displays a summary of completed operations.
type ResultModel struct {
	Tasks []RepoTask
	Title string
	Width int
}

func NewResultModel(title string, tasks []RepoTask) ResultModel {
	return ResultModel{
		Tasks: tasks,
		Title: title,
		Width: 80,
	}
}

func truncateError(err string, maxLen int) string {
	if len(err) <= maxLen {
		return err
	}
	// The juicy part is usually at the end — find the last meaningful segment
	// e.g. "branch update failed: failed to pull: error: cannot pull with rebase: You have unstaged changes"
	// We want: "cannot pull with rebase: You have unstaged changes"
	if idx := strings.LastIndex(err, "error: "); idx >= 0 {
		tail := err[idx+len("error: "):]
		if len(tail) <= maxLen {
			return tail
		}
	}
	// Fallback: take from the end
	return "…" + err[len(err)-maxLen+1:]
}

func (m ResultModel) View() string {
	var succeeded, failed int
	var okTasks, failTasks []RepoTask
	for _, t := range m.Tasks {
		switch t.Status {
		case TaskDone:
			succeeded++
			okTasks = append(okTasks, t)
		case TaskFailed:
			failed++
			failTasks = append(failTasks, t)
		}
	}

	var s string

	// Summary banner
	banner := resultAccent.Render(m.Title) + resultDim.Render("  ")
	banner += resultOk.Render(fmt.Sprintf("✔ %d", succeeded))
	if failed > 0 {
		banner += resultDim.Render("  ") + resultFail.Render(fmt.Sprintf("✗ %d", failed))
	}
	s += summaryBox.Render(banner) + "\n\n"

	maxErr := m.Width - 10
	if maxErr < 40 {
		maxErr = 40
	}

	// Failures first (they're what you care about)
	if len(failTasks) > 0 {
		var failContent string
		for _, t := range failTasks {
			errMsg := truncateError(t.Error, maxErr)
			failContent += fmt.Sprintf("  %s %s\n", resultFail.Render("✗"), nameStyle.Render(t.Name))
			failContent += fmt.Sprintf("    %s\n", resultDim.Render(errMsg))
		}
		failContent = strings.TrimRight(failContent, "\n")
		s += failBox.Render(failContent) + "\n\n"
	}

	// Successes — compact list
	if len(okTasks) > 0 {
		var okContent string
		for _, t := range okTasks {
			line := fmt.Sprintf("  %s %s", resultOk.Render("✔"), nameStyle.Render(t.Name))
			if t.Result != "" {
				line += "  " + resultDim.Render(t.Result)
			}
			okContent += line + "\n"
		}
		okContent = strings.TrimRight(okContent, "\n")
		s += successBox.Render(okContent) + "\n"
	}

	return s
}
