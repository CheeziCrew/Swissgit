package screens

import (
	"fmt"
	"sort"
	"strings"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/CheeziCrew/curd"

	"github.com/CheeziCrew/swissgit/git"
	"github.com/CheeziCrew/swissgit/ops"
	"github.com/CheeziCrew/swissgit/tui/components"
)

var (
	stBranchDefault = lipgloss.NewStyle().Foreground(colorGreen)
	stBranchOther   = lipgloss.NewStyle().Foreground(colorMagenta)
	stName          = lipgloss.NewStyle().Bold(true).Width(35).Foreground(colorFg)
	stNameActive    = lipgloss.NewStyle().Bold(true).Width(35).Foreground(colorBrMag)
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
	statusStepDiff
)

type statusTaskDoneMsg struct {
	index  int
	result ops.StatusResult
}

type statusReposDiscoveredMsg struct {
	paths []string
}

type statusDiffMsg struct {
	content string
}

// StatusActionMsg is emitted when user presses c/p on a dirty repo.
type StatusActionMsg struct {
	Path   string
	Action string // "commit" or "pullrequest"
}

// StatusModel handles the status view.
type StatusModel struct {
	step       statusStep
	progress   components.ProgressModel
	results    []ops.StatusResult
	dirtyRepos []ops.StatusResult
	cursor     int
	diffView   viewport.Model
	viewport   viewport.Model
	viewReady  bool
	width      int
	height     int
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
		m.width = wsm.Width
		m.height = wsm.Height
		if !m.viewReady {
			m.viewport = viewport.New(viewport.WithWidth(wsm.Width-4), viewport.WithHeight(wsm.Height-8))
			m.diffView = viewport.New(viewport.WithWidth(wsm.Width-4), viewport.WithHeight(wsm.Height-8))
			m.viewReady = true
		} else {
			m.viewport.SetWidth(wsm.Width - 4)
			m.viewport.SetHeight(wsm.Height - 8)
			m.diffView.SetWidth(wsm.Width - 4)
			m.diffView.SetHeight(wsm.Height - 8)
		}
	}

	switch m.step {
	case statusStepProgress:
		return m.updateProgress(msg)
	case statusStepResults:
		return m.updateResults(msg)
	case statusStepDiff:
		return m.updateDiff(msg)
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
		m.computeDirtyRepos()
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

func (m *StatusModel) computeDirtyRepos() {
	sorted := make([]ops.StatusResult, len(m.results))
	copy(sorted, m.results)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].RepoName < sorted[j].RepoName
	})

	m.dirtyRepos = nil
	for _, r := range sorted {
		if r.Error == "" && isDirtyRepo(r) {
			m.dirtyRepos = append(m.dirtyRepos, r)
		}
	}
	m.cursor = 0
}

func (m StatusModel) updateResults(msg tea.Msg) (StatusModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch {
		case key.Matches(msg, key.NewBinding(key.WithKeys("esc", "q"))):
			return m, func() tea.Msg { return BackToMenuMsg{} }
		case key.Matches(msg, key.NewBinding(key.WithKeys("j", "down"))):
			if len(m.dirtyRepos) > 0 {
				m.cursor = (m.cursor + 1) % len(m.dirtyRepos)
				m.viewport.SetContent(m.renderResults())
			}
			return m, nil
		case key.Matches(msg, key.NewBinding(key.WithKeys("k", "up"))):
			if len(m.dirtyRepos) > 0 {
				m.cursor = (m.cursor - 1 + len(m.dirtyRepos)) % len(m.dirtyRepos)
				m.viewport.SetContent(m.renderResults())
			}
			return m, nil
		case key.Matches(msg, key.NewBinding(key.WithKeys("d"))):
			if len(m.dirtyRepos) > 0 {
				repo := m.dirtyRepos[m.cursor]
				return m, func() tea.Msg {
					diff, err := ops.GetDiffStat(repo.Path)
					if err != nil {
						return statusDiffMsg{content: "No diff available: " + err.Error()}
					}
					if diff == "" {
						return statusDiffMsg{content: "No unstaged changes."}
					}
					return statusDiffMsg{content: diff}
				}
			}
			return m, nil
		case key.Matches(msg, key.NewBinding(key.WithKeys("c"))):
			if len(m.dirtyRepos) > 0 {
				repo := m.dirtyRepos[m.cursor]
				return m, func() tea.Msg {
					return StatusActionMsg{Path: repo.Path, Action: "commit"}
				}
			}
			return m, nil
		case key.Matches(msg, key.NewBinding(key.WithKeys("p"))):
			if len(m.dirtyRepos) > 0 {
				repo := m.dirtyRepos[m.cursor]
				return m, func() tea.Msg {
					return StatusActionMsg{Path: repo.Path, Action: "pullrequest"}
				}
			}
			return m, nil
		}

	case statusDiffMsg:
		m.step = statusStepDiff
		m.diffView.SetContent(msg.content)
		m.diffView.GotoTop()
		return m, nil
	}

	var cmd tea.Cmd
	m.viewport, cmd = m.viewport.Update(msg)
	return m, cmd
}

func (m StatusModel) updateDiff(msg tea.Msg) (StatusModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch {
		case key.Matches(msg, key.NewBinding(key.WithKeys("esc", "q"))):
			m.step = statusStepResults
			return m, nil
		}
	}

	var cmd tea.Cmd
	m.diffView, cmd = m.diffView.Update(msg)
	return m, cmd
}

func formatRepoLine(r ops.StatusResult) string {
	return formatRepoLineWithCursor(r, false)
}

func formatRepoLineWithCursor(r ops.StatusResult, active bool) string {
	branchStyle := stBranchDefault
	if r.Branch != r.DefaultBranch {
		branchStyle = stBranchOther
	}

	nameStyle := stName
	prefix := "  "
	if active {
		nameStyle = stNameActive
		prefix = "▸ "
	}

	line := prefix + nameStyle.Render(r.RepoName) + " " + branchStyle.Render("["+r.Branch+"]")

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

func statusBanner(total, dirty, errored, clean int) string {
	banner := stAccent.Render("Status")
	banner += stDim.Render("  ")
	banner += fmt.Sprintf("%d repos", total)
	if dirty > 0 {
		banner += stDim.Render("  ") + stModified.Render(fmt.Sprintf("⚡ %d dirty", dirty))
	}
	if errored > 0 {
		banner += stDim.Render("  ") + stErr.Render(fmt.Sprintf("✗ %d errors", errored))
	}
	if clean > 0 {
		banner += stDim.Render("  ") + stClean.Render(fmt.Sprintf("✓ %d clean", clean))
	}
	return banner
}

func renderStatusErrors(errored []ops.StatusResult) string {
	if len(errored) == 0 {
		return ""
	}
	stErrName := lipgloss.NewStyle().Bold(true).Foreground(colorFg)
	var content string
	for _, r := range errored {
		content += fmt.Sprintf("  %s %s\n", stErr.Render("✗"), stErrName.Render(r.RepoName))
		content += fmt.Sprintf("    %s\n", stDim.Render(r.Error))
	}
	return stErrorBox.Render(strings.TrimRight(content, "\n")) + "\n\n"
}

func isDirtyRepo(r ops.StatusResult) bool {
	return !r.Clean || r.Branch != r.DefaultBranch || r.Ahead > 0 || r.Behind > 0
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
		case isDirtyRepo(r):
			dirty = append(dirty, r)
		default:
			clean = append(clean, r)
		}
	}

	var s string
	s += stSummaryBox.Render(statusBanner(len(sorted), len(dirty), len(errored), len(clean))) + "\n\n"
	s += renderStatusErrors(errored)

	if len(dirty) > 0 {
		var content string
		for i, r := range dirty {
			active := i == m.cursor
			content += formatRepoLineWithCursor(r, active) + "\n"
		}
		s += stDirtyBox.Render(strings.TrimRight(content, "\n")) + "\n\n"
	}

	if len(clean) > 0 {
		var content string
		for _, r := range clean {
			content += fmt.Sprintf("  %s %s\n", stClean.Render("✓"), stDim.Render(r.RepoName))
		}
		s += stCleanBox.Render(strings.TrimRight(content, "\n")) + "\n"
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
		hints := []curd.Hint{
			{Key: "j/k", Desc: "move"},
			{Key: "d", Desc: "diff"},
			{Key: "c", Desc: "commit"},
			{Key: "p", Desc: "PR"},
			{Key: "esc/q", Desc: "menu"},
		}
		s += curd.RenderHintBar(st, hints)
		return s

	case statusStepDiff:
		if len(m.dirtyRepos) > 0 && m.cursor < len(m.dirtyRepos) {
			repo := m.dirtyRepos[m.cursor]
			s += stDim.Render("Diff: ") + stName.Render(repo.RepoName) + "\n\n"
		}
		if m.viewReady {
			s += m.diffView.View() + "\n"
		}
		s += curd.RenderHintBar(st, []curd.Hint{
			{Key: "esc", Desc: "back"},
		})
		return s
	}

	return s
}
