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
	brMergedStyle = lipgloss.NewStyle().Foreground(colorGreen)
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
	branchesStepConfirmDelete
)

type branchesDeleteDoneMsg struct {
	deleted int
	err     error
}

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
	throttle  *taskThrottle
	viewport  viewport.Model
	viewReady bool
	confirm   curd.ConfirmModel
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
			Status: components.TaskPending,
		})
	}

	m.progress = components.NewProgressModel(tasks)
	m.results = make([]ops.BranchesResult, len(paths))

	var taskCmds []tea.Cmd
	for i, p := range paths {
		idx := i
		path := p
		taskCmds = append(taskCmds, func() tea.Msg {
			result := ops.GetBranches(path)
			return branchesTaskDoneMsg{index: idx, result: result}
		})
	}

	m.throttle = newThrottle(taskCmds)
	initial := m.throttle.Start(&m.progress)
	return tea.Batch(append([]tea.Cmd{m.progress.Init()}, initial...)...)
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
	case branchesStepConfirmDelete:
		return m.updateConfirmDelete(msg)
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
		if next := m.throttle.Dispatch(&m.progress); next != nil {
			return m, tea.Batch(cmd, next)
		}
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

func (m BranchesModel) countMergedBranches() (int, int) {
	branches, repos := 0, 0
	for _, r := range m.results {
		repoHasMerged := false
		for _, b := range r.LocalBranches {
			if b.IsMerged && b.Name != r.DefaultBranch && b.Name != r.CurrentBranch {
				branches++
				repoHasMerged = true
			}
		}
		if repoHasMerged {
			repos++
		}
	}
	return branches, repos
}

func (m BranchesModel) deleteMergedCmd() tea.Cmd {
	return func() tea.Msg {
		total := 0
		for _, r := range m.results {
			if r.Path == "" || r.Error != "" {
				continue
			}
			var toDelete []string
			for _, b := range r.LocalBranches {
				if b.IsMerged && b.Name != r.DefaultBranch && b.Name != r.CurrentBranch {
					toDelete = append(toDelete, b.Name)
				}
			}
			if len(toDelete) > 0 {
				n, err := ops.DeleteMergedBranches(r.Path, toDelete)
				total += n
				if err != nil {
					return branchesDeleteDoneMsg{deleted: total, err: err}
				}
			}
		}
		return branchesDeleteDoneMsg{deleted: total}
	}
}

func (m BranchesModel) updateResults(msg tea.Msg) (BranchesModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch {
		case key.Matches(msg, key.NewBinding(key.WithKeys("esc", "q"))):
			return m, func() tea.Msg { return BackToMenuMsg{} }
		case key.Matches(msg, key.NewBinding(key.WithKeys("D"))):
			branches, repos := m.countMergedBranches()
			if branches == 0 {
				return m, nil
			}
			m.confirm = curd.NewConfirmModel(curd.ConfirmConfig{
				Question: fmt.Sprintf("Delete %d merged branch(es) across %d repo(s)?", branches, repos),
				Caller:   "branches-delete",
				Palette:  palette,
			})
			m.step = branchesStepConfirmDelete
			return m, nil
		}
	}

	var cmd tea.Cmd
	m.viewport, cmd = m.viewport.Update(msg)
	return m, cmd
}

func (m BranchesModel) updateConfirmDelete(msg tea.Msg) (BranchesModel, tea.Cmd) {
	switch msg := msg.(type) {
	case curd.ConfirmMsg:
		if msg.Confirmed {
			m.step = branchesStepProgress
			return m, m.deleteMergedCmd()
		}
		m.step = branchesStepResults
		return m, nil

	case curd.BackToMenuMsg:
		m.step = branchesStepResults
		return m, nil

	case branchesDeleteDoneMsg:
		// Re-scan after delete
		m.step = branchesStepProgress
		return m, discoverForBranches()
	}

	var cmd tea.Cmd
	m.confirm, cmd = m.confirm.Update(msg)
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

const (
	maxBranchesShown = 3
	maxBranchNameLen = 40
	fmtIndentStr     = "    %s\n"
	fmtNameBranch    = "  %s %s\n"
)

func truncateBranchName(name string) string {
	if len(name) <= maxBranchNameLen {
		return name
	}
	return name[:maxBranchNameLen-1] + "…"
}

func renderBranchList(branches []ops.BranchInfo, label string, defaultStyle lipgloss.Style, category string) string {
	if len(branches) == 0 {
		return ""
	}
	var s string
	shown := branches
	if len(shown) > maxBranchesShown {
		shown = shown[:maxBranchesShown]
	}
	for _, b := range shown {
		style := defaultStyle
		suffix := ""
		switch {
		case b.IsMerged:
			style = brMergedStyle
			suffix = " ✓merged"
		case b.IsStale:
			style = brStaleStyle
		}
		s += fmt.Sprintf("    %s %s%s\n", brLabel.Render(label), style.Render(truncateBranchName(b.Name)), brDim.Render(suffix))
	}
	if len(branches) > maxBranchesShown {
		s += fmt.Sprintf(fmtIndentStr, brDim.Render(fmt.Sprintf("  +%d more %s", len(branches)-maxBranchesShown, category)))
	}
	return s
}

func formatBranchEntry(r ops.BranchesResult) string {
	currentStyle := brCurrent
	if r.CurrentBranch != r.DefaultBranch {
		currentStyle = lipgloss.NewStyle().Foreground(colorRed).Bold(true)
	}

	s := fmt.Sprintf(fmtNameBranch, brRepoName.Render(r.RepoName), currentStyle.Render("["+r.CurrentBranch+"]"))

	var local []ops.BranchInfo
	for _, b := range r.LocalBranches {
		if b.Name != r.DefaultBranch {
			local = append(local, b)
		}
	}

	s += renderBranchList(local, "L", brLocalStyle, "local")
	s += renderBranchList(r.RemoteBranches, "R", brRemoteStyle, "remote")

	return s
}

func branchesBanner(total, interesting, errored, clean, merged int) string {
	banner := brAccent.Render("Branches")
	banner += brDim.Render("  ")
	banner += fmt.Sprintf("%d repos", total)
	if interesting > 0 {
		banner += brDim.Render("  ") + brRemoteStyle.Render(fmt.Sprintf("⚡ %d with branches", interesting))
	}
	if merged > 0 {
		banner += brDim.Render("  ") + brMergedStyle.Render(fmt.Sprintf("✓ %d merged", merged))
	}
	if errored > 0 {
		banner += brDim.Render("  ") + stErr.Render(fmt.Sprintf("✗ %d errors", errored))
	}
	if clean > 0 {
		banner += brDim.Render("  ") + brDim.Render(fmt.Sprintf("✓ %d clean", clean))
	}
	return banner
}

func renderBranchErrors(errored []ops.BranchesResult, boxW int) string {
	if len(errored) == 0 {
		return ""
	}
	brErrName := lipgloss.NewStyle().Bold(true).Foreground(colorFg)
	var content string
	for _, r := range errored {
		content += fmt.Sprintf(fmtNameBranch, stErr.Render("✗"), brErrName.Render(r.RepoName))
		content += fmt.Sprintf(fmtIndentStr, brDim.Render(r.Error))
	}
	return brErrorBox.MaxWidth(boxW).Render(strings.TrimRight(content, "\n")) + "\n\n"
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
	mergedCount := 0
	for _, r := range sorted {
		switch {
		case r.Error != "":
			errored = append(errored, r)
		case hasExtraBranches(r):
			interesting = append(interesting, r)
		default:
			clean = append(clean, r)
		}
		for _, b := range r.LocalBranches {
			if b.IsMerged && b.Name != r.DefaultBranch {
				mergedCount++
			}
		}
	}

	boxW := m.width - 6
	if boxW < 40 {
		boxW = 40
	}

	var s string
	s += brSummaryBox.Render(branchesBanner(len(sorted), len(interesting), len(errored), len(clean), mergedCount)) + "\n\n"
	s += renderBranchErrors(errored, boxW)

	if len(interesting) > 0 {
		var content string
		for _, r := range interesting {
			content += formatBranchEntry(r)
		}
		s += brDirtyBox.MaxWidth(boxW).Render(strings.TrimRight(content, "\n")) + "\n\n"
	}

	if len(clean) > 0 {
		var content string
		for _, r := range clean {
			content += fmt.Sprintf(fmtNameBranch, brDim.Render("✓"), brDim.Render(r.RepoName))
		}
		s += brCleanBox.MaxWidth(boxW).Render(strings.TrimRight(content, "\n")) + "\n"
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
		hints := []curd.Hint{{Key: "esc/q", Desc: "menu"}}
		if branches, _ := m.countMergedBranches(); branches > 0 {
			hints = append([]curd.Hint{{Key: "D", Desc: "delete merged"}}, hints...)
		}
		s += curd.RenderHintBar(st, hints)
		return s

	case branchesStepConfirmDelete:
		s += m.confirm.View()
		return s
	}

	return s
}
