package screens

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/CheeziCrew/swissgit/ops"
	"github.com/CheeziCrew/swissgit/tui/components"
)

type cleanupStep int

const (
	cleanupStepDrop cleanupStep = iota
	cleanupStepRepoSelect
	cleanupStepProgress
	cleanupStepResults
)

type cleanupTaskDoneMsg struct {
	index  int
	result ops.CleanupResult
}

// CleanupModel handles the cleanup flow.
type CleanupModel struct {
	step cleanupStep

	dropChanges bool
	dropCursor  int // 0=no, 1=yes

	repoSelect RepoSelectModel
	progress   components.ProgressModel
	results    components.ResultModel
	repos      []string
	viewport   viewport.Model
	viewReady  bool
	height     int
}

func NewCleanupModel() CleanupModel {
	return CleanupModel{
		step:       cleanupStepDrop,
		dropCursor: 0, // default to "no"
	}
}

func (m CleanupModel) Init() tea.Cmd {
	return nil
}

func (m CleanupModel) Update(msg tea.Msg) (CleanupModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.height = msg.Height
		if !m.viewReady {
			m.viewport = viewport.New(msg.Width-6, msg.Height-10)
			m.viewReady = true
		} else {
			m.viewport.Width = msg.Width - 6
			m.viewport.Height = msg.Height - 10
		}
	default:
		_ = msg
	}

	switch m.step {
	case cleanupStepDrop:
		return m.updateDrop(msg)
	case cleanupStepRepoSelect:
		return m.updateRepoSelect(msg)
	case cleanupStepProgress:
		return m.updateProgress(msg)
	case cleanupStepResults:
		return m.updateResults(msg)
	}
	return m, nil
}

func (m CleanupModel) updateDrop(msg tea.Msg) (CleanupModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, key.NewBinding(key.WithKeys("up", "k"))):
			m.dropCursor = 1 - m.dropCursor
		case key.Matches(msg, key.NewBinding(key.WithKeys("down", "j"))):
			m.dropCursor = 1 - m.dropCursor
		case key.Matches(msg, key.NewBinding(key.WithKeys("enter"))):
			m.dropChanges = m.dropCursor == 1
			m.step = cleanupStepRepoSelect
			m.repoSelect = NewRepoSelectModel("cleanup", ".", m.height)
			return m, m.repoSelect.Init()
		case key.Matches(msg, key.NewBinding(key.WithKeys("esc"))):
			return m, func() tea.Msg { return BackToMenuMsg{} }
		}
	}
	return m, nil
}

func (m CleanupModel) updateRepoSelect(msg tea.Msg) (CleanupModel, tea.Cmd) {
	switch msg := msg.(type) {
	case RepoSelectDoneMsg:
		m.repos = msg.Paths
		return m, m.startCleanupTasks()
	case BackToMenuMsg:
		m.step = cleanupStepDrop
		return m, nil
	}
	var cmd tea.Cmd
	m.repoSelect, cmd = m.repoSelect.Update(msg)
	return m, cmd
}

func (m *CleanupModel) startCleanupTasks() tea.Cmd {
	var tasks []components.RepoTask
	for _, p := range m.repos {
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
	m.step = cleanupStepProgress

	var cmds []tea.Cmd
	cmds = append(cmds, m.progress.Init())

	for i, p := range m.repos {
		idx := i
		path := p
		drop := m.dropChanges

		cmds = append(cmds, func() tea.Msg {
			result := ops.CleanupRepo(path, drop, "main")
			return cleanupTaskDoneMsg{index: idx, result: result}
		})
	}

	return tea.Batch(cmds...)
}

func (m CleanupModel) updateProgress(msg tea.Msg) (CleanupModel, tea.Cmd) {
	switch msg := msg.(type) {
	case cleanupTaskDoneMsg:
		status := components.TaskDone
		resultStr := ""
		errStr := ""
		if !msg.result.Success {
			status = components.TaskFailed
			errStr = msg.result.Error
		} else {
			var parts []string
			if msg.result.PrunedBranches > 0 {
				parts = append(parts, fmt.Sprintf("pruned %d branches", msg.result.PrunedBranches))
			}
			if msg.result.DroppedChanges {
				parts = append(parts, "dropped changes")
			}
			resultStr = strings.Join(parts, ", ")
		}

		updateMsg := components.RepoTaskUpdateMsg{
			Index: msg.index, Status: status, Result: resultStr, Error: errStr,
		}
		var cmd tea.Cmd
		m.progress, cmd = m.progress.Update(updateMsg)
		return m, cmd

	case components.AllTasksDoneMsg:
		m.results = components.NewResultModel("Cleanup", m.progress.Tasks)
		m.step = cleanupStepResults
		m.viewport.SetContent(m.results.View())
		return m, nil
	}

	var cmd tea.Cmd
	m.progress, cmd = m.progress.Update(msg)
	return m, cmd
}

func (m CleanupModel) updateResults(msg tea.Msg) (CleanupModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, key.NewBinding(key.WithKeys("esc", "q", "enter"))):
			return m, func() tea.Msg { return BackToMenuMsg{} }
		}
	}

	// Let viewport handle scrolling
	var cmd tea.Cmd
	m.viewport, cmd = m.viewport.Update(msg)
	return m, cmd
}

func (m CleanupModel) View() string {
	var s string
	s += titleStyle.Render("🧹 Cleanup") + "\n\n"

	switch m.step {
	case cleanupStepDrop:
		type dropOption struct {
			label string
			desc  string
		}
		options := []dropOption{
			{label: "No", desc: "keep everything (default)"},
			{label: "Yes", desc: "nuke local changes"},
		}

		s += prLabelStyle.Render("Drop uncommitted changes?") + "\n\n"
		for i, opt := range options {
			line := fmt.Sprintf("%s  %s", menuActiveName.Render(opt.label), menuActiveDesc.Render(opt.desc))
			if i == m.dropCursor {
				s += menuActiveItem.Render(line) + "\n"
			} else {
				inactiveLine := fmt.Sprintf("%s  %s", menuInactiveName.Render(opt.label), menuInactiveDesc.Render(opt.desc))
				s += menuInactiveItem.Render(inactiveLine) + "\n"
			}
		}
		s += "\n"
		s += menuHelpBox.Render("↑↓ navigate  •  enter select  •  esc back")

	case cleanupStepRepoSelect:
		dropLabel := "no"
		if m.dropChanges {
			dropLabel = lipgloss.NewStyle().Foreground(colorRed).Render("yes ⚠")
		}
		s += summaryBlock(summaryLine("drop changes", dropLabel))
		s += m.repoSelect.View()

	case cleanupStepProgress:
		s += m.progress.View()

	case cleanupStepResults:
		if m.viewReady {
			s += m.viewport.View() + "\n"
		} else {
			s += m.results.View() + "\n"
		}
		scrollHint := ""
		if m.viewReady && m.viewport.TotalLineCount() > m.viewport.VisibleLineCount() {
			scrollHint = "  •  ↑↓ scroll"
		}
		s += menuHelpBox.Render("esc/q back to menu" + scrollHint)
	}

	return s
}
