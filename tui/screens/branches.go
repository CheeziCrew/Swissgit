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
	brLocalStyle  = lipgloss.NewStyle().Foreground(colorFg)
	brRemoteStyle = lipgloss.NewStyle().Foreground(colorYellow)
	brStaleStyle  = lipgloss.NewStyle().Foreground(colorRed).Strikethrough(true)
	brCurrent     = lipgloss.NewStyle().Foreground(colorBlue).Bold(true)
	brRepoName    = lipgloss.NewStyle().Bold(true).Foreground(colorFg)
	brLabel       = lipgloss.NewStyle().Foreground(colorGray)

	brSummaryBox = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorBrMag).
			Padding(0, 2).
			Bold(true)

	brDirtyBox = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorYellow).
			Padding(0, 1)

	brCleanBox = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorGreen).
			Padding(0, 1)

	brErrorBox = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorRed).
			Padding(0, 1)

	brAccent = lipgloss.NewStyle().Foreground(colorBrMag)
	brDim    = lipgloss.NewStyle().Foreground(colorGray)
)

type branchesStep int

const (
	branchesStepProgress branchesStep = iota
	branchesStepResults
)

type branchesTaskDoneMsg struct {
	index  int
	result ops.BranchesResult
}

type branchesReposDiscoveredMsg struct {
	paths []string
}

// BranchesModel handles the branches view.
type BranchesModel struct {
	step      branchesStep
	progress  components.ProgressModel
	results   []ops.BranchesResult
	viewport  viewport.Model
	viewReady bool
	width     int
	height    int
}

func NewBranchesModel() BranchesModel {
	return BranchesModel{
		step: branchesStepProgress,
	}
}

func (m BranchesModel) Init() tea.Cmd {
	return discoverForBranches()
}

func discoverForBranches() tea.Cmd {
	return func() tea.Msg {
		paths, err := git.DiscoverRepos(".")
		if err != nil {
			return branchesReposDiscoveredMsg{}
		}
		return branchesReposDiscoveredMsg{paths: paths}
	}
}

func (m *BranchesModel) startBranchesTasks(paths []string) tea.Cmd {
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
	m.results = make([]ops.BranchesResult, len(paths))

	var cmds []tea.Cmd
	cmds = append(cmds, m.progress.Init())

	for i, p := range paths {
		idx := i
		path := p
		cmds = append(cmds, func() tea.Msg {
			result := ops.GetBranches(path)
			return branchesTaskDoneMsg{index: idx, result: result}
		})
	}

	return tea.Batch(cmds...)
}

func (m BranchesModel) Update(msg tea.Msg) (BranchesModel, tea.Cmd) {
	if wsm, ok := msg.(tea.WindowSizeMsg); ok {
		m.width = wsm.Width
		m.height = wsm.Height
		if !m.viewReady {
			m.viewport = viewport.New(viewport.WithWidth(wsm.Width-2), viewport.WithHeight(wsm.Height-8))
			m.viewReady = true
		} else {
			m.viewport.SetWidth(wsm.Width - 2)
			m.viewport.SetHeight(wsm.Height - 8)
		}
	}

	switch m.step {
	case branchesStepProgress:
		return m.updateProgress(msg)
	case branchesStepResults:
		return m.updateResults(msg)
	}
	return m, nil
}

func (m BranchesModel) updateProgress(msg tea.Msg) (BranchesModel, tea.Cmd) {
	switch msg := msg.(type) {
	case branchesReposDiscoveredMsg:
		if len(msg.paths) == 0 {
			m.step = branchesStepResults
			m.viewport.SetContent(m.renderResults())
			return m, nil
		}
		return m, m.startBranchesTasks(msg.paths)

	case branchesTaskDoneMsg:
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
		m.step = branchesStepResults
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

func (m BranchesModel) updateResults(msg tea.Msg) (BranchesModel, tea.Cmd) {
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

// hasExtraBranches returns true if the repo has local (non-default) or remote branches.
func hasExtraBranches(r ops.BranchesResult) bool {
	for _, b := range r.LocalBranches {
		if b.Name != r.DefaultBranch {
			return true
		}
	}
	return len(r.RemoteBranches) > 0
}

const maxBranchesShown = 3
const maxBranchNameLen = 40

func truncateBranchName(name string) string {
	if len(name) <= maxBranchNameLen {
		return name
	}
	return name[:maxBranchNameLen-1] + "…"
}

func formatBranchEntry(r ops.BranchesResult) string {
	currentStyle := brCurrent
	if r.CurrentBranch != r.DefaultBranch {
		currentStyle = lipgloss.NewStyle().Foreground(colorRed).Bold(true)
	}

	s := fmt.Sprintf("  %s %s\n", brRepoName.Render(r.RepoName), currentStyle.Render("["+r.CurrentBranch+"]"))

	// Collect local branches (skip default)
	var local []ops.BranchInfo
	for _, b := range r.LocalBranches {
		if b.Name != r.DefaultBranch {
			local = append(local, b)
		}
	}

	// Render local
	if len(local) > 0 {
		shown := local
		if len(shown) > maxBranchesShown {
			shown = shown[:maxBranchesShown]
		}
		for _, b := range shown {
			style := brLocalStyle
			if b.IsStale {
				style = brStaleStyle
			}
			s += fmt.Sprintf("    %s %s\n", brLabel.Render("L"), style.Render(truncateBranchName(b.Name)))
		}
		if len(local) > maxBranchesShown {
			s += fmt.Sprintf("    %s\n", brDim.Render(fmt.Sprintf("  +%d more local", len(local)-maxBranchesShown)))
		}
	}

	// Render remote
	remote := r.RemoteBranches
	if len(remote) > 0 {
		shown := remote
		if len(shown) > maxBranchesShown {
			shown = shown[:maxBranchesShown]
		}
		for _, b := range shown {
			style := brRemoteStyle
			if b.IsStale {
				style = brStaleStyle
			}
			s += fmt.Sprintf("    %s %s\n", brLabel.Render("R"), style.Render(truncateBranchName(b.Name)))
		}
		if len(remote) > maxBranchesShown {
			s += fmt.Sprintf("    %s\n", brDim.Render(fmt.Sprintf("  +%d more remote", len(remote)-maxBranchesShown)))
		}
	}

	return s
}

func (m BranchesModel) renderResults() string {
	sorted := make([]ops.BranchesResult, len(m.results))
	copy(sorted, m.results)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].RepoName < sorted[j].RepoName
	})

	if len(sorted) == 0 {
		return prDimStyle.Render("No git repositories found.")
	}

	var interesting, clean, errored []ops.BranchesResult
	for _, r := range sorted {
		switch {
		case r.Error != "":
			errored = append(errored, r)
		case hasExtraBranches(r):
			interesting = append(interesting, r)
		default:
			clean = append(clean, r)
		}
	}

	var s string

	// Summary banner
	banner := brAccent.Render("Branches")
	banner += brDim.Render("  ")
	banner += fmt.Sprintf("%d repos", len(sorted))
	if len(interesting) > 0 {
		banner += brDim.Render("  ") + brRemoteStyle.Render(fmt.Sprintf("⚡ %d with branches", len(interesting)))
	}
	if len(errored) > 0 {
		banner += brDim.Render("  ") + stErr.Render(fmt.Sprintf("✗ %d errors", len(errored)))
	}
	if len(clean) > 0 {
		banner += brDim.Render("  ") + brDim.Render(fmt.Sprintf("✓ %d clean", len(clean)))
	}
	s += brSummaryBox.Render(banner) + "\n\n"

	// Max box width to prevent wrapping
	boxW := m.width - 6
	if boxW < 40 {
		boxW = 40
	}

	// Errors first
	if len(errored) > 0 {
		brErrName := lipgloss.NewStyle().Bold(true).Foreground(colorFg)
		var content string
		for _, r := range errored {
			content += fmt.Sprintf("  %s %s\n", stErr.Render("✗"), brErrName.Render(r.RepoName))
			content += fmt.Sprintf("    %s\n", brDim.Render(r.Error))
		}
		content = strings.TrimRight(content, "\n")
		s += brErrorBox.MaxWidth(boxW).Render(content) + "\n\n"
	}

	// Repos with extra branches — the interesting stuff
	if len(interesting) > 0 {
		var content string
		for _, r := range interesting {
			content += formatBranchEntry(r)
		}
		content = strings.TrimRight(content, "\n")
		s += brDirtyBox.MaxWidth(boxW).Render(content) + "\n\n"
	}

	// Clean repos — just default branch, compact
	if len(clean) > 0 {
		var content string
		for _, r := range clean {
			content += fmt.Sprintf("  %s %s\n", brDim.Render("✓"), brDim.Render(r.RepoName))
		}
		content = strings.TrimRight(content, "\n")
		s += brCleanBox.MaxWidth(boxW).Render(content) + "\n"
	}

	return s
}

func (m BranchesModel) View() string {
	var s string
	s += titleStyle.Render("🌿 Branches") + "\n\n"

	switch m.step {
	case branchesStepProgress:
		s += m.progress.View()

	case branchesStepResults:
		if m.viewReady {
			s += m.viewport.View() + "\n"
		} else {
			s += m.renderResults() + "\n"
		}
		s += curd.RenderHintBar(st, []curd.Hint{
			{Key: "esc/q", Desc: "menu"},
		})
		return s
	}

	return s
}
