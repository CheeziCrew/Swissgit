package screens

import (
	"fmt"
	"sort"
	"strings"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/CheeziCrew/swissgit/git"
	"github.com/CheeziCrew/swissgit/ops"
	"github.com/CheeziCrew/swissgit/tui/components"
)

var (
	stBranchDefault = lipgloss.NewStyle().Foreground(colorGreen)
	stBranchOther   = lipgloss.NewStyle().Foreground(colorMagenta)
	stName          = lipgloss.NewStyle().Bold(true).Width(35).Foreground(colorFg)
	stClean         = lipgloss.NewStyle().Foreground(colorGray)
	stErr           = lipgloss.NewStyle().Foreground(colorRed)

	stAhead    = lipgloss.NewStyle().Foreground(colorGreen)
	stBehind   = lipgloss.NewStyle().Foreground(colorRed)
	stModified = lipgloss.NewStyle().Foreground(colorYellow)
	stAdded    = lipgloss.NewStyle().Foreground(colorGreen)
	stDeleted  = lipgloss.NewStyle().Foreground(colorRed)
	stUntrack  = lipgloss.NewStyle().Foreground(colorBlue)

	stSummaryBox = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorBrMag).
			Padding(0, 2).
			Bold(true)

	stDirtyBox = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorYellow).
			Padding(0, 1)

	stCleanBox = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorGreen).
			Padding(0, 1)

	stErrorBox = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorRed).
			Padding(0, 1)

	stAccent = lipgloss.NewStyle().Foreground(colorBrMag)
	stDim    = lipgloss.NewStyle().Foreground(colorGray)
)

type statusStep int

const (
	statusStepProgress statusStep = iota
	statusStepResults
)

type statusTaskDoneMsg struct {
	index  int
	result ops.StatusResult
}

type statusReposDiscoveredMsg struct {
	paths []string
}

// StatusModel handles the status view.
type StatusModel struct {
	step      statusStep
	progress  components.ProgressModel
	results   []ops.StatusResult
	viewport  viewport.Model
	viewReady bool
	height    int
}

func NewStatusModel() StatusModel {
	return StatusModel{
		step: statusStepProgress,
	}
}

func (m StatusModel) Init() tea.Cmd {
	return discoverForStatus()
}

func discoverForStatus() tea.Cmd {
	return func() tea.Msg {
		paths, err := git.DiscoverRepos(".")
		if err != nil {
			return statusReposDiscoveredMsg{}
		}
		return statusReposDiscoveredMsg{paths: paths}
	}
}

func (m *StatusModel) startStatusTasks(paths []string) tea.Cmd {
	var tasks []components.RepoTask
	for _, p := range paths {
		name := p
		if n, err := ops.GetRepoNameForPath(p); err == nil {
			name = n
		}
		tasks = append(tasks, components.RepoTask{
			Name:   name,
			Path:   p,
			Status: components.TaskRunning,
		})
	}

	m.progress = components.NewProgressModel(tasks)
	m.results = make([]ops.StatusResult, len(paths))

	var cmds []tea.Cmd
	cmds = append(cmds, m.progress.Init())

	for i, p := range paths {
		idx := i
		path := p
		cmds = append(cmds, func() tea.Msg {
			result := ops.GetRepoStatus(path)
			return statusTaskDoneMsg{index: idx, result: result}
		})
	}

	return tea.Batch(cmds...)
}

func (m StatusModel) Update(msg tea.Msg) (StatusModel, tea.Cmd) {
	if wsm, ok := msg.(tea.WindowSizeMsg); ok {
		m.height = wsm.Height
		if !m.viewReady {
			m.viewport = viewport.New(viewport.WithWidth(wsm.Width-4), viewport.WithHeight(wsm.Height-8))
			m.viewReady = true
		} else {
			m.viewport.SetWidth(wsm.Width - 4)
			m.viewport.SetHeight(wsm.Height - 8)
		}
	}

	switch m.step {
	case statusStepProgress:
		return m.updateProgress(msg)
	case statusStepResults:
		return m.updateResults(msg)
	}
	return m, nil
}

func (m StatusModel) updateProgress(msg tea.Msg) (StatusModel, tea.Cmd) {
	switch msg := msg.(type) {
	case statusReposDiscoveredMsg:
		if len(msg.paths) == 0 {
			m.step = statusStepResults
			m.viewport.SetContent(m.renderResults())
			return m, nil
		}
		return m, m.startStatusTasks(msg.paths)

	case statusTaskDoneMsg:
		m.results[msg.index] = msg.result

		status := components.TaskDone
		errStr := ""
		if msg.result.Error != "" {
			status = components.TaskFailed
			errStr = msg.result.Error
		}

		updateMsg := components.RepoTaskUpdateMsg{
			Index: msg.index, Status: status, Error: errStr,
		}
		var cmd tea.Cmd
		m.progress, cmd = m.progress.Update(updateMsg)
		return m, cmd

	case components.AllTasksDoneMsg:
		m.step = statusStepResults
		m.viewport.SetContent(m.renderResults())
		return m, nil

	case tea.KeyPressMsg:
		switch {
		case key.Matches(msg, key.NewBinding(key.WithKeys("esc", "q"))):
			return m, func() tea.Msg { return BackToMenuMsg{} }
		}
	}

	var cmd tea.Cmd
	m.progress, cmd = m.progress.Update(msg)
	return m, cmd
}

func (m StatusModel) updateResults(msg tea.Msg) (StatusModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch {
		case key.Matches(msg, key.NewBinding(key.WithKeys("esc", "q"))):
			return m, func() tea.Msg { return BackToMenuMsg{} }
		}
	}

	var cmd tea.Cmd
	m.viewport, cmd = m.viewport.Update(msg)
	return m, cmd
}

func formatRepoLine(r ops.StatusResult) string {
	branchStyle := stBranchDefault
	if r.Branch != r.DefaultBranch {
		branchStyle = stBranchOther
	}

	line := stName.Render(r.RepoName) + " " + branchStyle.Render("["+r.Branch+"]")

	var badges []string

	if r.Ahead > 0 {
		badges = append(badges, stAhead.Render(fmt.Sprintf("↑ %d", r.Ahead)))
	}
	if r.Behind > 0 {
		badges = append(badges, stBehind.Render(fmt.Sprintf("↓ %d", r.Behind)))
	}
	if r.Modified > 0 {
		badges = append(badges, stModified.Render(fmt.Sprintf("~ %d", r.Modified)))
	}
	if r.Added > 0 {
		badges = append(badges, stAdded.Render(fmt.Sprintf("+ %d", r.Added)))
	}
	if r.Deleted > 0 {
		badges = append(badges, stDeleted.Render(fmt.Sprintf("- %d", r.Deleted)))
	}
	if r.Untracked > 0 {
		badges = append(badges, stUntrack.Render(fmt.Sprintf("? %d", r.Untracked)))
	}

	if len(badges) > 0 {
		line += "  " + strings.Join(badges, " ")
	}

	return line
}

func (m StatusModel) renderResults() string {
	sorted := make([]ops.StatusResult, len(m.results))
	copy(sorted, m.results)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].RepoName < sorted[j].RepoName
	})

	if len(sorted) == 0 {
		return prDimStyle.Render("No git repositories found.")
	}

	var dirty, clean, errored []ops.StatusResult
	for _, r := range sorted {
		switch {
		case r.Error != "":
			errored = append(errored, r)
		case !r.Clean || r.Branch != r.DefaultBranch || r.Ahead > 0 || r.Behind > 0:
			dirty = append(dirty, r)
		default:
			clean = append(clean, r)
		}
	}

	var s string

	// Summary banner
	banner := stAccent.Render("Status")
	banner += stDim.Render("  ")
	banner += fmt.Sprintf("%d repos", len(sorted))
	if len(dirty) > 0 {
		banner += stDim.Render("  ") + stModified.Render(fmt.Sprintf("⚡ %d dirty", len(dirty)))
	}
	if len(errored) > 0 {
		banner += stDim.Render("  ") + stErr.Render(fmt.Sprintf("✗ %d errors", len(errored)))
	}
	if len(clean) > 0 {
		banner += stDim.Render("  ") + stClean.Render(fmt.Sprintf("✓ %d clean", len(clean)))
	}
	s += stSummaryBox.Render(banner) + "\n\n"

	// Errors first
	if len(errored) > 0 {
		stErrName := lipgloss.NewStyle().Bold(true).Foreground(colorFg)
		var content string
		for _, r := range errored {
			content += fmt.Sprintf("  %s %s\n", stErr.Render("✗"), stErrName.Render(r.RepoName))
			content += fmt.Sprintf("    %s\n", stDim.Render(r.Error))
		}
		content = strings.TrimRight(content, "\n")
		s += stErrorBox.Render(content) + "\n\n"
	}

	// Dirty repos — the interesting stuff
	if len(dirty) > 0 {
		var content string
		for _, r := range dirty {
			content += "  " + formatRepoLine(r) + "\n"
		}
		content = strings.TrimRight(content, "\n")
		s += stDirtyBox.Render(content) + "\n\n"
	}

	// Clean repos — compact
	if len(clean) > 0 {
		var content string
		for _, r := range clean {
			content += fmt.Sprintf("  %s %s\n", stClean.Render("✓"), stDim.Render(r.RepoName))
		}
		content = strings.TrimRight(content, "\n")
		s += stCleanBox.Render(content) + "\n"
	}

	return s
}

func (m StatusModel) View() string {
	var s string
	s += titleStyle.Render("📊 Status") + "\n\n"

	switch m.step {
	case statusStepProgress:
		s += m.progress.View()

	case statusStepResults:
		if m.viewReady {
			s += m.viewport.View() + "\n"
		} else {
			s += m.renderResults() + "\n"
		}
		return s
	}

	return s
}
